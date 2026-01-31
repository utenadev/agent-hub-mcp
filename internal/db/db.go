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
