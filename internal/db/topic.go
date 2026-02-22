package db

import (
	"fmt"
)

// Topic represents a discussion topic.
type Topic struct {
	ID        int
	Title     string
	CreatedAt string
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
