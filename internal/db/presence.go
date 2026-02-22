package db

import (
	"database/sql"
	"fmt"
)

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
