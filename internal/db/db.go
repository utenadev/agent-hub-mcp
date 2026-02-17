package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.go
var schemaFS embed.FS

type DB struct {
	*sql.DB
}

// Open opens a SQLite database at the given path.
// If the file doesn't exist, it will be created.
func Open(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{DB: sqlDB}

	// Enable WAL mode for better concurrent access
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Create schema
	if err := db.CreateSchema(); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return db, nil
}

// CreateSchema creates the database tables.
func (db *DB) CreateSchema() error {
	_, err := db.Exec(schemaSQL)
	return err
}

// Topic represents a discussion topic.
type Topic struct {
	ID        int
	Title     string
	CreatedAt string
}

// Message represents a message in a topic.
type Message struct {
	ID        int
	TopicID   int
	Sender    string
	Content   string
	CreatedAt string
}

// TopicSummary represents a summary of a topic.
type TopicSummary struct {
	ID          int
	TopicID     int
	SummaryText string
	IsMock      bool
	CreatedAt   string
}

// CreateTopic creates a new topic and returns its ID.
func (db *DB) CreateTopic(title string) (int64, error) {
	result, err := db.Exec("INSERT INTO topics (title) VALUES (?)", title)
	if err != nil {
		return 0, fmt.Errorf("failed to create topic: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// PostMessage posts a message to a topic.
func (db *DB) PostMessage(topicID int64, sender, content string) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO messages (topic_id, sender, content) VALUES (?, ?, ?)",
		topicID, sender, content,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to post message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// GetMessages retrieves recent messages from a topic.
func (db *DB) GetMessages(topicID int64, limit int) ([]Message, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := db.Query(
		"SELECT id, topic_id, sender, content, created_at FROM messages WHERE topic_id = ? ORDER BY id DESC LIMIT ?",
		topicID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.TopicID, &m.Sender, &m.Content, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// ListTopics retrieves all topics.
func (db *DB) ListTopics() ([]Topic, error) {
	rows, err := db.Query("SELECT id, title, created_at FROM topics ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query topics: %w", err)
	}
	defer rows.Close()

	var topics []Topic
	for rows.Next() {
		var t Topic
		if err := rows.Scan(&t.ID, &t.Title, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan topic: %w", err)
		}
		topics = append(topics, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating topics: %w", err)
	}

	return topics, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}

// CheckIntegrity verifies the database health and configuration.
// Returns detailed information about which tables are missing.
func (db *DB) CheckIntegrity() (map[string]bool, error) {
	requiredTables := []string{"topics", "messages", "topic_summaries", "agent_presence"}
	results := make(map[string]bool)
	var missingTables []string

	for _, table := range requiredTables {
		var name string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&name)

		if err != nil {
			if err == sql.ErrNoRows {
				results[table] = false
				missingTables = append(missingTables, table)
			} else {
				return results, fmt.Errorf("failed to check table %s: %w", table, err)
			}
		} else {
			results[table] = true
		}
	}

	if len(missingTables) > 0 {
		return results, fmt.Errorf("missing tables: %v", missingTables)
	}

	// Check journal mode
	var mode string
	err := db.QueryRow("PRAGMA journal_mode").Scan(&mode)
	if err != nil {
		return results, fmt.Errorf("failed to check journal mode: %w", err)
	}
	if mode != "wal" {
		return results, fmt.Errorf("database is not in WAL mode (current: %s)", mode)
	}

	return results, nil
}

// SaveSummary saves a summary for a topic.
func (db *DB) SaveSummary(topicID int64, summaryText string, isMock bool) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO topic_summaries (topic_id, summary_text, is_mock) VALUES (?, ?, ?)",
		topicID, summaryText, isMock,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to save summary: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// GetLatestSummary retrieves the latest summary for a topic.
func (db *DB) GetLatestSummary(topicID int64) (*TopicSummary, error) {
	row := db.QueryRow(
		"SELECT id, topic_id, summary_text, is_mock, created_at FROM topic_summaries WHERE topic_id = ? ORDER BY created_at DESC LIMIT 1",
		topicID,
	)

	var s TopicSummary
	err := row.Scan(&s.ID, &s.TopicID, &s.SummaryText, &s.IsMock, &s.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No summary found
		}
		return nil, fmt.Errorf("failed to get latest summary: %w", err)
	}

	return &s, nil
}

// GetSummariesByTopic retrieves all summaries for a topic, ordered by most recent first.
func (db *DB) GetSummariesByTopic(topicID int64) ([]TopicSummary, error) {
	rows, err := db.Query(
		"SELECT id, topic_id, summary_text, is_mock, created_at FROM topic_summaries WHERE topic_id = ? ORDER BY created_at DESC",
		topicID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query summaries: %w", err)
	}
	defer rows.Close()

	var summaries []TopicSummary
	for rows.Next() {
		var s TopicSummary
		if err := rows.Scan(&s.ID, &s.TopicID, &s.SummaryText, &s.IsMock, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan summary: %w", err)
		}
		summaries = append(summaries, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating summaries: %w", err)
	}

	return summaries, nil
}

// AgentPresence represents an agent's presence status.
type AgentPresence struct {
	Name      string
	Role      string
	Status    string
	TopicID   *int64
	LastSeen  string
	LastCheck string
}

// UpsertAgentPresence registers or updates an agent's presence.
func (db *DB) UpsertAgentPresence(name, role string) error {
	_, err := db.Exec(
		`INSERT INTO agent_presence (name, role, status, last_seen) 
		 VALUES (?, ?, 'online', CURRENT_TIMESTAMP)
		 ON CONFLICT(name) DO UPDATE SET
		 role = excluded.role,
		 status = excluded.status,
		 last_seen = CURRENT_TIMESTAMP`,
		name, role,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert agent presence: %w", err)
	}
	return nil
}

// UpdateAgentStatus updates an agent's status and current topic.
func (db *DB) UpdateAgentStatus(name, status string, topicID *int64) error {
	var err error
	if topicID != nil {
		_, err = db.Exec(
			"UPDATE agent_presence SET status = ?, topic_id = ?, last_seen = CURRENT_TIMESTAMP WHERE name = ?",
			status, *topicID, name,
		)
	} else {
		_, err = db.Exec(
			"UPDATE agent_presence SET status = ?, topic_id = NULL, last_seen = CURRENT_TIMESTAMP WHERE name = ?",
			status, name,
		)
	}
	if err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}
	return nil
}

// UpdateAgentCheckTime updates an agent's last check time.
func (db *DB) UpdateAgentCheckTime(name string) error {
	_, err := db.Exec(
		"UPDATE agent_presence SET last_check = CURRENT_TIMESTAMP WHERE name = ?",
		name,
	)
	if err != nil {
		return fmt.Errorf("failed to update agent check time: %w", err)
	}
	return nil
}

// GetAgentPresence retrieves an agent's presence information.
func (db *DB) GetAgentPresence(name string) (*AgentPresence, error) {
	row := db.QueryRow(
		"SELECT name, role, status, topic_id, last_seen, last_check FROM agent_presence WHERE name = ?",
		name,
	)

	var p AgentPresence
	var topicID sql.NullInt64
	err := row.Scan(&p.Name, &p.Role, &p.Status, &topicID, &p.LastSeen, &p.LastCheck)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get agent presence: %w", err)
	}

	if topicID.Valid {
		p.TopicID = &topicID.Int64
	}

	return &p, nil
}

// ListAllAgentPresence retrieves all agents' presence information.
func (db *DB) ListAllAgentPresence() ([]AgentPresence, error) {
	rows, err := db.Query(
		"SELECT name, role, status, topic_id, last_seen, last_check FROM agent_presence ORDER BY last_seen DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query agent presence: %w", err)
	}
	defer rows.Close()

	var presences []AgentPresence
	for rows.Next() {
		var p AgentPresence
		var topicID sql.NullInt64
		if err := rows.Scan(&p.Name, &p.Role, &p.Status, &topicID, &p.LastSeen, &p.LastCheck); err != nil {
			return nil, fmt.Errorf("failed to scan agent presence: %w", err)
		}
		if topicID.Valid {
			p.TopicID = &topicID.Int64
		}
		presences = append(presences, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating agent presence: %w", err)
	}

	return presences, nil
}

// CountUnreadMessages counts messages since the agent's last check.
func (db *DB) CountUnreadMessages(agentName string) (int64, error) {
	var count int64
	err := db.QueryRow(
		`SELECT COUNT(*) FROM messages 
		 WHERE created_at > (
			 SELECT COALESCE(last_check, '1970-01-01 00:00:00') 
			 FROM agent_presence 
			 WHERE name = ?
		 )`,
		agentName,
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count unread messages: %w", err)
	}

	return count, nil
}
