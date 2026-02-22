package db

import (
	"fmt"
)

// Message represents a message in a topic.
type Message struct {
	ID        int
	TopicID   int
	Sender    string
	Content   string
	CreatedAt string
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
