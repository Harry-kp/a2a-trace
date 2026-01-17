package proxy

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/harry-kp/a2a-trace/internal/store"
)

// MessageHandler is called when a message is intercepted
type MessageHandler func(msg *store.Message)

// AgentHandler is called when an agent is discovered
type AgentHandler func(agent *store.Agent)

// Proxy is an HTTP proxy that intercepts A2A traffic
type Proxy struct {
	server      *http.Server
	interceptor *Interceptor
	store       *store.Store
	traceID     string
	port        int
	onMessage   MessageHandler
	onAgent     AgentHandler
	client      *http.Client
}

// Config holds proxy configuration
type Config struct {
	Port       int
	Store      *store.Store
	TraceID    string
	OnMessage  MessageHandler
	OnAgent    AgentHandler
}

// New creates a new Proxy instance
func New(cfg Config) *Proxy {
	// Create HTTP client with custom transport
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: false},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &Proxy{
		interceptor: NewInterceptor(),
		store:       cfg.Store,
		traceID:     cfg.TraceID,
		port:        cfg.Port,
		onMessage:   cfg.OnMessage,
		onAgent:     cfg.OnAgent,
		client: &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
		},
	}
}

// Start starts the proxy server
func (p *Proxy) Start() error {
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API endpoints for UI
	mux.HandleFunc("/api/messages", p.handleGetMessages)
	mux.HandleFunc("/api/agents", p.handleGetAgents)
	mux.HandleFunc("/api/trace", p.handleGetTrace)
	mux.HandleFunc("/api/export", p.handleExport)

	// Proxy all other requests
	mux.HandleFunc("/", p.handleProxy)

	p.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", p.port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("🔍 A2A Trace proxy starting on port %d", p.port)
	return p.server.ListenAndServe()
}

// Stop gracefully stops the proxy server
func (p *Proxy) Stop() error {
	if p.server == nil {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return p.server.Shutdown(ctx)
}

// handleProxy handles proxied requests
func (p *Proxy) handleProxy(w http.ResponseWriter, r *http.Request) {
	// Check for CONNECT (HTTPS tunneling)
	if r.Method == "CONNECT" {
		p.handleConnect(w, r)
		return
	}

	// Get target URL from request
	targetURL := r.URL.String()
	if !strings.HasPrefix(targetURL, "http") {
		// If using as forward proxy, URL should be absolute
		// Otherwise, use Host header
		targetURL = "http://" + r.Host + r.URL.RequestURI()
	}

	// Read request body
	reqBody, newReqBody, err := p.interceptor.ReadBody(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	r.Body = newReqBody

	// Parse request for A2A
	var reqMsg *store.Message
	if p.interceptor.IsA2ARequest(r) || len(reqBody) > 0 {
		reqMsg = p.interceptor.ParseRequest(r, reqBody, p.traceID)
		
		// Store request
		if err := p.store.SaveMessage(reqMsg); err != nil {
			log.Printf("Failed to save request: %v", err)
		}
		
		// Notify handler
		if p.onMessage != nil {
			p.onMessage(reqMsg)
		}
	}

	startTime := time.Now()

	// Create the proxied request
	proxyReq, err := http.NewRequest(r.Method, targetURL, bytes.NewReader(reqBody))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Remove proxy-specific headers
	proxyReq.Header.Del("Proxy-Connection")
	proxyReq.Header.Del("Proxy-Authenticate")
	proxyReq.Header.Del("Proxy-Authorization")

	// Send request
	resp, err := p.client.Do(proxyReq)
	if err != nil {
		// Log error and return
		if reqMsg != nil {
			errMsg := &store.Message{
				TraceID:    p.traceID,
				Timestamp:  time.Now(),
				Direction:  "response",
				URL:        targetURL,
				Error:      err.Error(),
				DurationMs: time.Since(startTime).Milliseconds(),
				RequestID:  reqMsg.ID,
			}
			_ = p.store.SaveMessage(errMsg)
			if p.onMessage != nil {
				p.onMessage(errMsg)
			}
		}
		http.Error(w, fmt.Sprintf("Proxy error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	// Parse response for A2A
	if reqMsg != nil {
		respMsg := p.interceptor.ParseResponse(resp, respBody, reqMsg, duration)
		
		// Store response
		if err := p.store.SaveMessage(respMsg); err != nil {
			log.Printf("Failed to save response: %v", err)
		}
		
		// Notify handler
		if p.onMessage != nil {
			p.onMessage(respMsg)
		}

		// Check if this is an agent card response
		if strings.Contains(r.URL.Path, "/.well-known/agent.json") {
			if agent := p.interceptor.ParseAgentCard(respBody, targetURL); agent != nil {
				_ = p.store.SaveAgent(agent)
				if p.onAgent != nil {
					p.onAgent(agent)
				}
			}
		}
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Write status code and body
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

// handleConnect handles HTTPS CONNECT tunneling
func (p *Proxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	// For HTTPS, we just tunnel without intercepting
	// (intercepting HTTPS requires certificate setup)
	
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

// API handlers for UI

func (p *Proxy) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	messages, err := p.store.GetMessages(p.traceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json, _ := json.Marshal(messages)
	w.Write(json)
}

func (p *Proxy) handleGetAgents(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	agents, err := p.store.GetAgents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json, _ := json.Marshal(agents)
	w.Write(json)
}

func (p *Proxy) handleGetTrace(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	trace, err := p.store.GetTrace(p.traceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json, _ := json.Marshal(trace)
	w.Write(json)
}

func (p *Proxy) handleExport(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	data, err := p.store.ExportTrace(p.traceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=trace-%s.json", p.traceID))
	w.Write(data)
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// CreateReverseProxy creates a reverse proxy for a specific target
func CreateReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}
	
	return proxy
}


