package store

import (
	"time"
)

// Trace represents a single tracing session
type Trace struct {
	ID        string    `json:"id"`
	StartedAt time.Time `json:"started_at"`
	Command   string    `json:"command"`
	Status    string    `json:"status"` // "running", "completed", "error"
}

// Message represents an A2A protocol message (request or response)
type Message struct {
	ID          string    `json:"id"`
	TraceID     string    `json:"trace_id"`
	Timestamp   time.Time `json:"timestamp"`
	Direction   string    `json:"direction"` // "request" or "response"
	FromAgent   string    `json:"from_agent"`
	ToAgent     string    `json:"to_agent"`
	Method      string    `json:"method"` // A2A method like "tasks/create"
	URL         string    `json:"url"`
	Headers     string    `json:"headers"` // JSON string
	Body        string    `json:"body"`    // Full JSON body
	DurationMs  int64     `json:"duration_ms"`
	StatusCode  int       `json:"status_code"`
	Error       string    `json:"error,omitempty"`
	RequestID   string    `json:"request_id,omitempty"` // Links response to request
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
}

// Agent represents a discovered A2A agent
type Agent struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	Skills      string `json:"skills,omitempty"` // JSON array
	FirstSeen   time.Time `json:"first_seen"`
}

// A2ARequest represents a parsed A2A JSON-RPC request
type A2ARequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	ID      interface{} `json:"id"`
	Params  interface{} `json:"params,omitempty"`
}

// A2AResponse represents a parsed A2A JSON-RPC response
type A2AResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *A2AError   `json:"error,omitempty"`
}

// A2AError represents an A2A protocol error
type A2AError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// AgentCard represents the A2A agent card (/.well-known/agent.json)
type AgentCard struct {
	Name            string       `json:"name"`
	Description     string       `json:"description,omitempty"`
	URL             string       `json:"url"`
	Version         string       `json:"version,omitempty"`
	ProtocolVersion string       `json:"protocol_version,omitempty"`
	Capabilities    *Capabilities `json:"capabilities,omitempty"`
	Skills          []Skill      `json:"skills,omitempty"`
}

// Capabilities represents agent capabilities
type Capabilities struct {
	Streaming              bool `json:"streaming,omitempty"`
	PushNotifications      bool `json:"push_notifications,omitempty"`
	StateTransitionHistory bool `json:"state_transition_history,omitempty"`
}

// Skill represents an A2A agent skill
type Skill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Examples    []string `json:"examples,omitempty"`
}

// Insight represents an automatically detected issue or pattern
type Insight struct {
	ID        string    `json:"id"`
	TraceID   string    `json:"trace_id"`
	MessageID string    `json:"message_id,omitempty"`
	Type      string    `json:"type"` // "error", "warning", "info"
	Category  string    `json:"category"` // "slow_response", "retry_loop", "protocol_violation"
	Title     string    `json:"title"`
	Details   string    `json:"details"`
	Timestamp time.Time `json:"timestamp"`
}

// WebSocketMessage represents a message sent to the UI
type WebSocketMessage struct {
	Type    string      `json:"type"` // "message", "agent", "insight", "trace_status"
	Payload interface{} `json:"payload"`
}

