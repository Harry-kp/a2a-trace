package analyzer

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/harry-kp/a2a-trace/internal/store"
)

// Analyzer detects patterns and issues in A2A traffic
type Analyzer struct {
	store          *store.Store
	traceID        string
	slowThreshold  time.Duration
	onInsight      func(*store.Insight)
	requestTimes   map[string]time.Time
	methodCounts   map[string]int
	agentErrors    map[string]int
}

// Config holds analyzer configuration
type Config struct {
	Store         *store.Store
	TraceID       string
	SlowThreshold time.Duration
	OnInsight     func(*store.Insight)
}

// New creates a new Analyzer instance
func New(cfg Config) *Analyzer {
	threshold := cfg.SlowThreshold
	if threshold == 0 {
		threshold = time.Second // Default 1 second
	}

	return &Analyzer{
		store:         cfg.Store,
		traceID:       cfg.TraceID,
		slowThreshold: threshold,
		onInsight:     cfg.OnInsight,
		requestTimes:  make(map[string]time.Time),
		methodCounts:  make(map[string]int),
		agentErrors:   make(map[string]int),
	}
}

// AnalyzeMessage analyzes a message and generates insights
func (a *Analyzer) AnalyzeMessage(msg *store.Message) []*store.Insight {
	var insights []*store.Insight

	if msg.Direction == "request" {
		a.requestTimes[msg.ID] = msg.Timestamp
		a.methodCounts[msg.Method]++
	}

	if msg.Direction == "response" {
		// Check for slow responses
		if insight := a.checkSlowResponse(msg); insight != nil {
			insights = append(insights, insight)
		}

		// Check for errors
		if insight := a.checkError(msg); insight != nil {
			insights = append(insights, insight)
		}

		// Check for protocol violations
		if insight := a.checkProtocolViolation(msg); insight != nil {
			insights = append(insights, insight)
		}
	}

	// Check for retry loops
	if insight := a.checkRetryLoop(msg); insight != nil {
		insights = append(insights, insight)
	}

	// Save and broadcast insights
	for _, insight := range insights {
		if err := a.store.SaveInsight(insight); err == nil {
			if a.onInsight != nil {
				a.onInsight(insight)
			}
		}
	}

	return insights
}

// checkSlowResponse checks if a response is slow
func (a *Analyzer) checkSlowResponse(msg *store.Message) *store.Insight {
	if msg.DurationMs <= a.slowThreshold.Milliseconds() {
		return nil
	}

	return &store.Insight{
		ID:        uuid.New().String(),
		TraceID:   a.traceID,
		MessageID: msg.ID,
		Type:      "warning",
		Category:  "slow_response",
		Title:     "Slow Response Detected",
		Details:   formatSlowResponseDetails(msg),
		Timestamp: time.Now(),
	}
}

// checkError checks for errors in responses
func (a *Analyzer) checkError(msg *store.Message) *store.Insight {
	if msg.Error == "" && msg.StatusCode < 400 {
		return nil
	}

	// Track errors per agent
	a.agentErrors[msg.FromAgent]++

	insightType := "error"
	if msg.StatusCode >= 400 && msg.StatusCode < 500 {
		insightType = "warning"
	}

	return &store.Insight{
		ID:        uuid.New().String(),
		TraceID:   a.traceID,
		MessageID: msg.ID,
		Type:      insightType,
		Category:  "error",
		Title:     formatErrorTitle(msg),
		Details:   formatErrorDetails(msg),
		Timestamp: time.Now(),
	}
}

// checkProtocolViolation checks for A2A protocol violations
func (a *Analyzer) checkProtocolViolation(msg *store.Message) *store.Insight {
	var violations []string

	// Check response body for JSON-RPC compliance
	if msg.Body != "" {
		var resp map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Body), &resp); err == nil {
			// Check for required fields
			if _, ok := resp["jsonrpc"]; !ok {
				violations = append(violations, "Missing 'jsonrpc' field")
			}
			if _, ok := resp["id"]; !ok {
				// id can be null for notifications, but should exist for responses
				if msg.StatusCode >= 200 && msg.StatusCode < 300 {
					if _, hasResult := resp["result"]; hasResult {
						violations = append(violations, "Missing 'id' field in response")
					}
				}
			}
		}
	}

	if len(violations) == 0 {
		return nil
	}

	return &store.Insight{
		ID:        uuid.New().String(),
		TraceID:   a.traceID,
		MessageID: msg.ID,
		Type:      "warning",
		Category:  "protocol_violation",
		Title:     "A2A Protocol Violation",
		Details:   strings.Join(violations, "; "),
		Timestamp: time.Now(),
	}
}

// checkRetryLoop checks for potential retry loops
func (a *Analyzer) checkRetryLoop(msg *store.Message) *store.Insight {
	if msg.Method == "" {
		return nil
	}

	// If we've seen this method more than 5 times in quick succession
	count := a.methodCounts[msg.Method]
	if count > 0 && count%5 == 0 {
		return &store.Insight{
			ID:        uuid.New().String(),
			TraceID:   a.traceID,
			MessageID: msg.ID,
			Type:      "warning",
			Category:  "retry_loop",
			Title:     "Potential Retry Loop Detected",
			Details:   formatRetryLoopDetails(msg.Method, count),
			Timestamp: time.Now(),
		}
	}

	return nil
}

// GetSummary returns a summary of the analysis
func (a *Analyzer) GetSummary() map[string]interface{} {
	insights, _ := a.store.GetInsights(a.traceID)
	messages, _ := a.store.GetMessages(a.traceID)

	// Calculate statistics
	var totalDuration int64
	var errorCount int
	var successCount int

	for _, msg := range messages {
		if msg.Direction == "response" {
			totalDuration += msg.DurationMs
			if msg.Error != "" || msg.StatusCode >= 400 {
				errorCount++
			} else {
				successCount++
			}
		}
	}

	avgDuration := int64(0)
	responseCount := successCount + errorCount
	if responseCount > 0 {
		avgDuration = totalDuration / int64(responseCount)
	}

	return map[string]interface{}{
		"total_messages":    len(messages),
		"total_insights":    len(insights),
		"error_count":       errorCount,
		"success_count":     successCount,
		"avg_duration_ms":   avgDuration,
		"method_counts":     a.methodCounts,
		"agent_error_counts": a.agentErrors,
	}
}

// Helper functions for formatting

func formatSlowResponseDetails(msg *store.Message) string {
	return formatDetails(map[string]interface{}{
		"duration_ms": msg.DurationMs,
		"url":         msg.URL,
		"method":      msg.Method,
		"suggestion":  "Consider adding timeout handling or investigating agent performance",
	})
}

func formatErrorTitle(msg *store.Message) string {
	if msg.StatusCode >= 400 {
		return "HTTP Error " + string(rune(msg.StatusCode))
	}
	return "A2A Error Response"
}

func formatErrorDetails(msg *store.Message) string {
	details := map[string]interface{}{
		"status_code": msg.StatusCode,
		"error":       msg.Error,
		"url":         msg.URL,
		"method":      msg.Method,
	}

	// Parse error body if JSON-RPC error
	if msg.Body != "" {
		var resp store.A2AResponse
		if err := json.Unmarshal([]byte(msg.Body), &resp); err == nil && resp.Error != nil {
			details["error_code"] = resp.Error.Code
			details["error_message"] = resp.Error.Message
		}
	}

	return formatDetails(details)
}

func formatRetryLoopDetails(method string, count int) string {
	return formatDetails(map[string]interface{}{
		"method":     method,
		"call_count": count,
		"suggestion": "Check for proper error handling and backoff logic",
	})
}

func formatDetails(data map[string]interface{}) string {
	bytes, _ := json.MarshalIndent(data, "", "  ")
	return string(bytes)
}

