package db

import (
	"database/sql"
	"fmt"
)

// TopicSummary represents a summary of a topic.
type TopicSummary struct {
	ID          int
	TopicID     int
	SummaryText string
	IsMock      bool
	CreatedAt   string
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
