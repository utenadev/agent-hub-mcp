package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yklcs/agent-hub-mcp/internal/config"
	"github.com/yklcs/agent-hub-mcp/internal/db"
	"github.com/yklcs/agent-hub-mcp/internal/mcp"
)

// runServe starts the MCP server.
func (a *App) runServe(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Suppress flag errors during testing
	dbPath := fs.String("db", config.DefaultDBPath(), "Path to SQLite database")
	sseAddr := fs.String("sse", "", "Enable SSE mode on address (e.g., :8080)")
	senderFlag := fs.String("sender", "", "Default sender name for messages (overrides BBS_AGENT_ID env var)")
	roleFlag := fs.String("role", "", "Agent role (overrides BBS_AGENT_ROLE env var)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Determine sender: flag > env var > default ("unknown")
	sender := *senderFlag
	if sender == "" {
		sender = os.Getenv("BBS_AGENT_ID")
		if sender == "" {
			sender = "unknown"
		}
	}

	// Determine role: flag > env var > default ("agent")
	role := *roleFlag
	if role == "" {
		role = os.Getenv("BBS_AGENT_ROLE")
		if role == "" {
			role = "agent"
		}
	}

	database, err := db.Open(*dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	// Check database integrity before starting
	if results, err := database.CheckIntegrity(); err != nil {
		// Print detailed error information to stderr
		fmt.Fprintf(stderr, "Database check failed:\n")
		for table, exists := range results {
			status := "OK"
			if !exists {
				status = "MISSING"
			}
			fmt.Fprintf(stderr, "  - %s: %s\n", table, status)
		}
		return fmt.Errorf("database integrity check failed: %w (run 'agent-hub setup' if this is a new installation)", err)
	}

	// Register/update agent presence
	if err := database.UpsertAgentPresence(sender, role); err != nil {
		fmt.Fprintf(stderr, "Warning: failed to register agent presence: %v\n", err)
	}

	fmt.Fprintf(stderr, "Database opened: %s\n", *dbPath)
	fmt.Fprintf(stderr, "Agent: name=%s, role=%s\n", sender, role)

	srv := mcp.NewServer(database, sender, role)

	if *sseAddr != "" {
		host := *sseAddr
		if host[0] == ':' {
			host = "localhost" + host
		}
		fmt.Fprintf(stderr, "Starting MCP server on SSE http://%s...\n", *sseAddr)
		fmt.Fprintf(stderr, "\n--- SSE Connection Info ---\n")
		fmt.Fprintf(stderr, "SSE Endpoint:     http://%s/sse\n", host)
		fmt.Fprintf(stderr, "Message Endpoint: http://%s/message\n", host)
		fmt.Fprintf(stderr, "---------------------------\n\n")
		if err := srv.ServeSSE(*sseAddr); err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	} else {
		fmt.Fprintln(stderr, "Starting MCP server on stdio...")
		if err := srv.Serve(); err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	return nil
}
