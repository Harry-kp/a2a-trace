package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/harry-kp/a2a-trace/internal/store"
)

// Interceptor parses and classifies A2A protocol messages
type Interceptor struct{}

// NewInterceptor creates a new Interceptor instance
func NewInterceptor() *Interceptor {
	return &Interceptor{}
}

// IsA2ARequest checks if a request is an A2A protocol request
func (i *Interceptor) IsA2ARequest(r *http.Request) bool {
	// Check content type
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return false
	}

	// A2A uses POST for JSON-RPC
	if r.Method != "POST" {
		// GET to /.well-known/agent.json is also A2A
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/.well-known/agent.json") {
			return true
		}
		return false
	}

	return true
}

// ParseRequest parses an HTTP request into an A2A message
func (i *Interceptor) ParseRequest(r *http.Request, body []byte, traceID string) *store.Message {
	msg := &store.Message{
		TraceID:     traceID,
		Timestamp:   time.Now(),
		Direction:   "request",
		URL:         r.URL.String(),
		ContentType: r.Header.Get("Content-Type"),
		Size:        int64(len(body)),
		Body:        string(body),
	}

	// Parse headers
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	headersJSON, _ := json.Marshal(headers)
	msg.Headers = string(headersJSON)

	// Extract target agent from URL
	msg.ToAgent = extractAgentFromURL(r.URL.String())

	// Parse JSON-RPC to extract method
	var a2aReq store.A2ARequest
	if err := json.Unmarshal(body, &a2aReq); err == nil {
		msg.Method = a2aReq.Method
		if a2aReq.ID != nil {
			msg.RequestID = formatRequestID(a2aReq.ID)
		}
	}

	return msg
}

// ParseResponse parses an HTTP response into an A2A message
func (i *Interceptor) ParseResponse(resp *http.Response, body []byte, requestMsg *store.Message, duration time.Duration) *store.Message {
	msg := &store.Message{
		TraceID:     requestMsg.TraceID,
		Timestamp:   time.Now(),
		Direction:   "response",
		URL:         requestMsg.URL,
		FromAgent:   requestMsg.ToAgent,
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Size:        int64(len(body)),
		Body:        string(body),
		DurationMs:  duration.Milliseconds(),
		RequestID:   requestMsg.RequestID,
	}

	// Parse headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	headersJSON, _ := json.Marshal(headers)
	msg.Headers = string(headersJSON)

	// Parse JSON-RPC response for errors
	var a2aResp store.A2AResponse
	if err := json.Unmarshal(body, &a2aResp); err == nil {
		if a2aResp.Error != nil {
			msg.Error = a2aResp.Error.Message
		}
	}

	// Check HTTP error
	if resp.StatusCode >= 400 {
		msg.Error = http.StatusText(resp.StatusCode)
	}

	return msg
}

// ParseAgentCard parses an agent card response
func (i *Interceptor) ParseAgentCard(body []byte, url string) *store.Agent {
	var card store.AgentCard
	if err := json.Unmarshal(body, &card); err != nil {
		return nil
	}

	skillsJSON, _ := json.Marshal(card.Skills)

	return &store.Agent{
		URL:         url,
		Name:        card.Name,
		Description: card.Description,
		Version:     card.Version,
		Skills:      string(skillsJSON),
		FirstSeen:   time.Now(),
	}
}

// ReadBody reads and restores an HTTP request/response body
func (i *Interceptor) ReadBody(body io.ReadCloser) ([]byte, io.ReadCloser, error) {
	if body == nil {
		return nil, nil, nil
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, nil, err
	}
	body.Close()

	// Return a new reader for the original body
	return data, io.NopCloser(bytes.NewReader(data)), nil
}

// extractAgentFromURL extracts the agent identifier from a URL
func extractAgentFromURL(urlStr string) string {
	// Remove protocol and path, keep host
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "https://")
	
	// Get just the host part
	if idx := strings.Index(urlStr, "/"); idx != -1 {
		urlStr = urlStr[:idx]
	}
	
	return urlStr
}

// formatRequestID converts the JSON-RPC id to a string
func formatRequestID(id interface{}) string {
	switch v := id.(type) {
	case string:
		return v
	case float64:
		return strings.TrimSuffix(strings.TrimSuffix(
			strings.Replace(string(rune(int(v))), ".", "", 1), "0"), ".")
	default:
		data, _ := json.Marshal(id)
		return string(data)
	}
}

// ClassifyMethod returns a human-readable description of an A2A method
func ClassifyMethod(method string) string {
	methodDescriptions := map[string]string{
		"tasks/create":   "Create Task",
		"tasks/get":      "Get Task Status",
		"tasks/cancel":   "Cancel Task",
		"tasks/send":     "Send Message",
		"tasks/sendSubscribe": "Send & Subscribe",
		"tasks/resubscribe":   "Resubscribe to Task",
	}

	if desc, ok := methodDescriptions[method]; ok {
		return desc
	}
	return method
}

