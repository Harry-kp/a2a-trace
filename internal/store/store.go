package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// Store manages SQLite database operations for traces
type Store struct {
	db *sql.DB
	mu sync.RWMutex
}

// New creates a new Store instance with an in-memory or file-based SQLite database
func New(dbPath string) (*Store, error) {
	if dbPath == "" {
		dbPath = ":memory:"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

// migrate creates the database schema
func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS traces (
		id TEXT PRIMARY KEY,
		started_at TIMESTAMP NOT NULL,
		command TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'running'
	);

	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		trace_id TEXT NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		direction TEXT NOT NULL,
		from_agent TEXT,
		to_agent TEXT,
		method TEXT,
		url TEXT,
		headers TEXT,
		body TEXT,
		duration_ms INTEGER DEFAULT 0,
		status_code INTEGER DEFAULT 0,
		error TEXT,
		request_id TEXT,
		content_type TEXT,
		size INTEGER DEFAULT 0,
		FOREIGN KEY (trace_id) REFERENCES traces(id)
	);

	CREATE TABLE IF NOT EXISTS agents (
		id TEXT PRIMARY KEY,
		url TEXT UNIQUE NOT NULL,
		name TEXT,
		description TEXT,
		version TEXT,
		skills TEXT,
		first_seen TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS insights (
		id TEXT PRIMARY KEY,
		trace_id TEXT NOT NULL,
		message_id TEXT,
		type TEXT NOT NULL,
		category TEXT NOT NULL,
		title TEXT NOT NULL,
		details TEXT,
		timestamp TIMESTAMP NOT NULL,
		FOREIGN KEY (trace_id) REFERENCES traces(id)
	);

	CREATE INDEX IF NOT EXISTS idx_messages_trace_id ON messages(trace_id);
	CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
	CREATE INDEX IF NOT EXISTS idx_insights_trace_id ON insights(trace_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

// CreateTrace creates a new trace session
func (s *Store) CreateTrace(command string) (*Trace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	trace := &Trace{
		ID:        uuid.New().String(),
		StartedAt: time.Now(),
		Command:   command,
		Status:    "running",
	}

	_, err := s.db.Exec(
		"INSERT INTO traces (id, started_at, command, status) VALUES (?, ?, ?, ?)",
		trace.ID, trace.StartedAt, trace.Command, trace.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace: %w", err)
	}

	return trace, nil
}

// UpdateTraceStatus updates the status of a trace
func (s *Store) UpdateTraceStatus(traceID, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("UPDATE traces SET status = ? WHERE id = ?", status, traceID)
	return err
}

// GetTrace retrieves a trace by ID
func (s *Store) GetTrace(traceID string) (*Trace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trace := &Trace{}
	err := s.db.QueryRow(
		"SELECT id, started_at, command, status FROM traces WHERE id = ?",
		traceID,
	).Scan(&trace.ID, &trace.StartedAt, &trace.Command, &trace.Status)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return trace, nil
}

// SaveMessage saves an A2A message to the database
func (s *Store) SaveMessage(msg *Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	_, err := s.db.Exec(`
		INSERT INTO messages (
			id, trace_id, timestamp, direction, from_agent, to_agent,
			method, url, headers, body, duration_ms, status_code, error,
			request_id, content_type, size
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.TraceID, msg.Timestamp, msg.Direction, msg.FromAgent, msg.ToAgent,
		msg.Method, msg.URL, msg.Headers, msg.Body, msg.DurationMs, msg.StatusCode, msg.Error,
		msg.RequestID, msg.ContentType, msg.Size,
	)
	return err
}

// GetMessages retrieves all messages for a trace
func (s *Store) GetMessages(traceID string) ([]*Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT id, trace_id, timestamp, direction, from_agent, to_agent,
			method, url, headers, body, duration_ms, status_code, error,
			request_id, content_type, size
		FROM messages WHERE trace_id = ? ORDER BY timestamp ASC`,
		traceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		var fromAgent, toAgent, method, url, headers, body, errStr, requestID, contentType sql.NullString
		err := rows.Scan(
			&msg.ID, &msg.TraceID, &msg.Timestamp, &msg.Direction,
			&fromAgent, &toAgent, &method, &url, &headers, &body,
			&msg.DurationMs, &msg.StatusCode, &errStr, &requestID,
			&contentType, &msg.Size,
		)
		if err != nil {
			return nil, err
		}
		msg.FromAgent = fromAgent.String
		msg.ToAgent = toAgent.String
		msg.Method = method.String
		msg.URL = url.String
		msg.Headers = headers.String
		msg.Body = body.String
		msg.Error = errStr.String
		msg.RequestID = requestID.String
		msg.ContentType = contentType.String
		messages = append(messages, msg)
	}

	return messages, nil
}

// SaveAgent saves or updates an agent
func (s *Store) SaveAgent(agent *Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if agent.ID == "" {
		agent.ID = uuid.New().String()
	}

	_, err := s.db.Exec(`
		INSERT INTO agents (id, url, name, description, version, skills, first_seen)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			version = excluded.version,
			skills = excluded.skills`,
		agent.ID, agent.URL, agent.Name, agent.Description, agent.Version, agent.Skills, agent.FirstSeen,
	)
	return err
}

// GetAgents retrieves all discovered agents
func (s *Store) GetAgents() ([]*Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT id, url, name, description, version, skills, first_seen
		FROM agents ORDER BY first_seen DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*Agent
	for rows.Next() {
		agent := &Agent{}
		var name, desc, version, skills sql.NullString
		err := rows.Scan(&agent.ID, &agent.URL, &name, &desc, &version, &skills, &agent.FirstSeen)
		if err != nil {
			return nil, err
		}
		agent.Name = name.String
		agent.Description = desc.String
		agent.Version = version.String
		agent.Skills = skills.String
		agents = append(agents, agent)
	}

	return agents, nil
}

// SaveInsight saves an insight to the database
func (s *Store) SaveInsight(insight *Insight) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if insight.ID == "" {
		insight.ID = uuid.New().String()
	}

	_, err := s.db.Exec(`
		INSERT INTO insights (id, trace_id, message_id, type, category, title, details, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		insight.ID, insight.TraceID, insight.MessageID, insight.Type, insight.Category,
		insight.Title, insight.Details, insight.Timestamp,
	)
	return err
}

// GetInsights retrieves all insights for a trace
func (s *Store) GetInsights(traceID string) ([]*Insight, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT id, trace_id, message_id, type, category, title, details, timestamp
		FROM insights WHERE trace_id = ? ORDER BY timestamp DESC`,
		traceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insights []*Insight
	for rows.Next() {
		insight := &Insight{}
		var messageID sql.NullString
		err := rows.Scan(
			&insight.ID, &insight.TraceID, &messageID, &insight.Type,
			&insight.Category, &insight.Title, &insight.Details, &insight.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		insight.MessageID = messageID.String
		insights = append(insights, insight)
	}

	return insights, nil
}

// ExportTrace exports a trace as JSON
func (s *Store) ExportTrace(traceID string) ([]byte, error) {
	trace, err := s.GetTrace(traceID)
	if err != nil {
		return nil, err
	}

	messages, err := s.GetMessages(traceID)
	if err != nil {
		return nil, err
	}

	insights, err := s.GetInsights(traceID)
	if err != nil {
		return nil, err
	}

	export := map[string]interface{}{
		"trace":    trace,
		"messages": messages,
		"insights": insights,
	}

	return json.MarshalIndent(export, "", "  ")
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

