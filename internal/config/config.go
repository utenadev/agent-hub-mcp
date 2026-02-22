// Package config provides centralized configuration management.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all application configuration.
type Config struct {
	// Database
	DBPath string `json:"db_path"`

	// Agent Identity
	AgentID   string `json:"agent_id"`
	AgentRole string `json:"agent_role"`

	// API Keys
	GeminiAPIKey string `json:"gemini_api_key"`
}

// DefaultDBPath returns the standard database path.
func DefaultDBPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "agent-hub.db"
	}
	return filepath.Join(configDir, "agent-hub-mcp", "agent-hub.db")
}

// DefaultConfigPath returns the standard config file path.
func DefaultConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(configDir, "agent-hub-mcp", "config.json")
}

// DefaultConfigDir returns the standard config directory.
func DefaultConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "."
	}
	return filepath.Join(configDir, "agent-hub-mcp")
}

// New creates a new Config with default values.
func New() *Config {
	return &Config{
		DBPath:    DefaultDBPath(),
		AgentID:   os.Getenv("BBS_AGENT_ID"),
		AgentRole: os.Getenv("BBS_AGENT_ROLE"),
		GeminiAPIKey: func() string {
			if key := os.Getenv("GEMINI_API_KEY"); key != "" {
				return key
			}
			return os.Getenv("HUB_MASTER_API_KEY")
		}(),
	}
}

// Load loads configuration from environment variables and optional config file.
// Precedence: flag > env var > config file > default
func Load(configPath string) (*Config, error) {
	cfg := New()

	// Load from config file if exists
	if configPath != "" {
		if err := cfg.LoadFromFile(configPath); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	return cfg, nil
}

// LoadFromFile loads configuration from a JSON file.
func (c *Config) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var fileConfig Config
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge file config (only override if not empty)
	if fileConfig.DBPath != "" {
		c.DBPath = fileConfig.DBPath
	}
	if fileConfig.AgentID != "" {
		c.AgentID = fileConfig.AgentID
	}
	if fileConfig.AgentRole != "" {
		c.AgentRole = fileConfig.AgentRole
	}
	if fileConfig.GeminiAPIKey != "" {
		c.GeminiAPIKey = fileConfig.GeminiAPIKey
	}

	return nil
}

// SaveToFile saves the configuration to a JSON file.
func (c *Config) SaveToFile(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetSender returns the agent ID, falling back to default if not set.
func (c *Config) GetSender(defaultSender string) string {
	if c.AgentID != "" {
		return c.AgentID
	}
	return defaultSender
}

// GetRole returns the agent role, falling back to default if not set.
func (c *Config) GetRole(defaultRole string) string {
	if c.AgentRole != "" {
		return c.AgentRole
	}
	return defaultRole
}
