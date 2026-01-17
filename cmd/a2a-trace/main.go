package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/harry-kp/a2a-trace/internal/analyzer"
	"github.com/harry-kp/a2a-trace/internal/cli"
	"github.com/harry-kp/a2a-trace/internal/process"
	"github.com/harry-kp/a2a-trace/internal/proxy"
	"github.com/harry-kp/a2a-trace/internal/store"
	"github.com/harry-kp/a2a-trace/internal/websocket"
)

//go:embed ui/out/*
var uiFS embed.FS

func main() {
	// Parse CLI arguments
	cfg, err := cli.ParseArgs()
	if err != nil {
		os.Exit(1)
	}

	// Print banner
	cli.PrintBanner(cfg)

	// Initialize store
	dataStore, err := store.New(cfg.DBPath)
	if err != nil {
		cli.PrintError("Failed to initialize database", err)
		os.Exit(1)
	}
	defer dataStore.Close()

	// Create trace session
	trace, err := dataStore.CreateTrace(fmt.Sprintf("%v", cfg.Command))
	if err != nil {
		cli.PrintError("Failed to create trace", err)
		os.Exit(1)
	}

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize analyzer
	analyzer := analyzer.New(analyzer.Config{
		Store:         dataStore,
		TraceID:       trace.ID,
		SlowThreshold: time.Second,
		OnInsight: func(insight *store.Insight) {
			wsHub.BroadcastInsight(insight)
			if cfg.Verbose {
				log.Printf("Insight: %s - %s", insight.Category, insight.Title)
			}
		},
	})

	// Set up HTTP server with UI and WebSocket
	mux := http.NewServeMux()

	// Determine if we need to share the mux with proxy
	var additionalMux *http.ServeMux
	if cfg.UIPort == cfg.Port && !cfg.NoUI {
		additionalMux = mux
	}

	// Initialize proxy
	proxyServer := proxy.New(proxy.Config{
		Port:          cfg.Port,
		Store:         dataStore,
		TraceID:       trace.ID,
		AdditionalMux: additionalMux,
		OnMessage: func(msg *store.Message) {
			wsHub.BroadcastMessage(msg)
			analyzer.AnalyzeMessage(msg)
			if cfg.Verbose {
				log.Printf("[%s] %s %s (%dms)", msg.Direction, msg.Method, msg.URL, msg.DurationMs)
			}
		},
		OnAgent: func(agent *store.Agent) {
			wsHub.BroadcastAgent(agent)
			if cfg.Verbose {
				log.Printf("Discovered agent: %s (%s)", agent.Name, agent.URL)
			}
		},
	})

	// WebSocket endpoint
	mux.HandleFunc("/ws", wsHub.HandleWebSocket)

	// API endpoints
	mux.HandleFunc("/api/messages", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		messages, _ := dataStore.GetMessages(trace.ID)
		writeJSON(w, messages)
	})
	mux.HandleFunc("/api/agents", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		agents, _ := dataStore.GetAgents()
		writeJSON(w, agents)
	})
	mux.HandleFunc("/api/trace", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		t, _ := dataStore.GetTrace(trace.ID)
		writeJSON(w, t)
	})
	mux.HandleFunc("/api/insights", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		insights, _ := dataStore.GetInsights(trace.ID)
		writeJSON(w, insights)
	})
	mux.HandleFunc("/api/summary", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		summary := analyzer.GetSummary()
		writeJSON(w, summary)
	})
	mux.HandleFunc("/api/export", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		data, _ := dataStore.ExportTrace(trace.ID)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=trace-%s.json", trace.ID))
		w.Write(data)
	})

	// Serve embedded UI
	if !cfg.NoUI {
		uiContent, err := fs.Sub(uiFS, "ui/out")
		if err != nil {
			// UI not embedded, serve placeholder
			mux.HandleFunc("/ui", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(placeholderHTML))
			})
			mux.HandleFunc("/ui/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(placeholderHTML))
			})
		} else {
			fileServer := http.FileServer(http.FS(uiContent))
			mux.Handle("/ui/", http.StripPrefix("/ui/", fileServer))
			mux.HandleFunc("/ui", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/ui/", http.StatusMovedPermanently)
			})
		}
	}

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start HTTP server for UI
	uiServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.UIPort),
		Handler: mux,
	}

	var wg sync.WaitGroup

	// Start UI server if port is different from proxy
	if cfg.UIPort != cfg.Port && !cfg.NoUI {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := uiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				cli.PrintError("UI server error", err)
			}
		}()
	}

	// Start proxy server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := proxyServer.Start(); err != nil && err != http.ErrServerClosed {
			cli.PrintError("Proxy server error", err)
		}
	}()

	// Give servers time to start
	time.Sleep(100 * time.Millisecond)

	// Initialize process manager
	procMgr, err := process.New(process.Config{
		Command:   cfg.Command,
		ProxyPort: cfg.Port,
		OutputHandler: func(line string, isStderr bool) {
			// Output is already printed by the process manager
		},
	})
	if err != nil {
		cli.PrintError("Failed to create process manager", err)
		os.Exit(1)
	}

	// Start the user's command
	if err := procMgr.Start(); err != nil {
		cli.PrintError("Failed to start command", err)
		os.Exit(1)
	}

	fmt.Printf("📍 Process started (PID: %d)\n\n", procMgr.PID())

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for process to exit or signal
	exitCode := 0
	done := make(chan struct{})

	go func() {
		code, err := procMgr.Wait()
		if err != nil {
			cli.PrintError("Process error", err)
			exitCode = 1
		} else {
			exitCode = code
		}
		close(done)
	}()

	select {
	case <-done:
		// Process exited naturally
	case sig := <-sigChan:
		fmt.Printf("\n📍 Received %v, shutting down...\n", sig)
		_ = procMgr.Stop()
		<-done
	}

	// Update trace status
	_ = dataStore.UpdateTraceStatus(trace.ID, "completed")

	// Print summary
	summary := analyzer.GetSummary()
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  A2A Trace Summary")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Messages:    %v\n", summary["total_messages"])
	fmt.Printf("  Insights:    %v\n", summary["total_insights"])
	fmt.Printf("  Errors:      %v\n", summary["error_count"])
	fmt.Printf("  Avg Latency: %vms\n", summary["avg_duration_ms"])
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Stop servers
	_ = proxyServer.Stop()
	if cfg.UIPort != cfg.Port {
		_ = uiServer.Close()
	}

	os.Exit(exitCode)
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	jsonData, _ := json.Marshal(data)
	w.Write(jsonData)
}

const placeholderHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>A2A Trace</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            color: #e4e4e7;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            text-align: center;
            padding: 2rem;
        }
        h1 {
            font-size: 3rem;
            margin-bottom: 1rem;
            background: linear-gradient(90deg, #3b82f6, #8b5cf6);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        p {
            color: #a1a1aa;
            margin-bottom: 2rem;
            font-size: 1.125rem;
        }
        .status {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            padding: 0.75rem 1.5rem;
            background: rgba(34, 197, 94, 0.1);
            border: 1px solid rgba(34, 197, 94, 0.3);
            border-radius: 9999px;
            color: #22c55e;
        }
        .dot {
            width: 8px;
            height: 8px;
            background: #22c55e;
            border-radius: 50%;
            animation: pulse 2s infinite;
        }
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }
        .info {
            margin-top: 2rem;
            padding: 1rem;
            background: rgba(255,255,255,0.05);
            border-radius: 0.5rem;
            font-family: monospace;
            font-size: 0.875rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔍 A2A Trace</h1>
        <p>Visual debugger for multi-agent systems</p>
        <div class="status">
            <span class="dot"></span>
            Tracing Active
        </div>
        <div class="info">
            <p>UI coming soon! For now, use the API:</p>
            <p>GET /api/messages - List all messages</p>
            <p>GET /api/agents - List discovered agents</p>
            <p>GET /api/insights - List insights</p>
            <p>GET /api/export - Export trace as JSON</p>
            <p>WS /ws - WebSocket for real-time updates</p>
        </div>
    </div>
</body>
</html>`

