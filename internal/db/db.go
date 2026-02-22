package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB wraps sql.DB with our schema.
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
