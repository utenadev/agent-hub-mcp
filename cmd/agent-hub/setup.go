package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/yklcs/agent-hub-mcp/internal/config"
	"github.com/yklcs/agent-hub-mcp/internal/db"
)

// runSetup initializes the system.
func (a *App) runSetup(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	dbPath := fs.String("db", config.DefaultDBPath(), "Path to SQLite database")
	force := fs.Bool("force", false, "Overwrites existing configuration if present")
	if err := fs.Parse(args); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "--- Agent Hub Setup ---")

	// 1. Initialize Database
	if _, err := os.Stat(*dbPath); err == nil && !*force {
		fmt.Fprintf(stdout, "[*] Database already exists at %s. Use -force to overwrite.\n", *dbPath)
	} else {
		fmt.Fprintf(stdout, "[*] Initializing Database at %s... ", *dbPath)
		database, err := db.Open(*dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		database.Close()
		fmt.Fprintln(stdout, "OK")
	}

	// 2. Create Config Directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}
	agentHubDir := filepath.Join(configDir, "agent-hub-mcp")
	fmt.Fprintf(stdout, "[*] Creating Config Directory (%s)... ", agentHubDir)
	if err := os.MkdirAll(agentHubDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	fmt.Fprintln(stdout, "OK")

	// 3. Create config.json template
	configPath := filepath.Join(agentHubDir, "config.json")
	if _, err := os.Stat(configPath); err == nil && !*force {
		fmt.Fprintf(stdout, "[*] Config file already exists at %s. Use -force to overwrite.\n", configPath)
	} else {
		fmt.Fprintf(stdout, "[*] Creating Config Template (%s)... ", configPath)
		config := map[string]string{
			"gemini_api_key": "YOUR_GEMINI_API_KEY_HERE",
			"default_sender": "",
			"default_role":   "",
		}
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
		fmt.Fprintln(stdout, "OK")
	}

	// 4. Claude Desktop Config Helper
	fmt.Fprintln(stdout, "\n--- Claude Desktop Configuration ---")

	// Get executable path
	execPath, err := os.Executable()
	if err != nil {
		execPath = "agent-hub"
	}

	fmt.Fprintln(stdout, "Add the following to your Claude Desktop configuration:")
	fmt.Fprintln(stdout)

	// OS-specific config path
	var claudeConfigPath string
	switch runtime.GOOS {
	case "darwin":
		claudeConfigPath = "~/Library/Application Support/Claude/claude_desktop_config.json"
	case "windows":
		claudeConfigPath = "%APPDATA%/Claude/claude_desktop_config.json"
	default: // linux and others
		claudeConfigPath = "~/.config/Claude/claude_desktop_config.json"
	}

	fmt.Fprintf(stdout, "Config file location: %s\n", claudeConfigPath)
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "JSON snippet to add:")
	fmt.Fprintln(stdout, "```json")
	fmt.Fprintf(stdout, "{\n  \"mcpServers\": {\n    \"agent-hub\": {\n      \"command\": \"%s\",\n      \"args\": [\"serve\", \"-db\", \"%s\"]\n    }\n  }\n}\n", execPath, *dbPath)
	fmt.Fprintln(stdout, "```")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Setup complete! You can now run 'agent-hub serve' or 'agent-hub doctor'.")

	return nil
}
