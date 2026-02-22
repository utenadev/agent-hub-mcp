package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/yklcs/agent-hub-mcp/internal/db"
	"github.com/yklcs/agent-hub-mcp/internal/hub"
)

// runOrchestrator starts the orchestrator monitor.
func (a *App) runOrchestrator(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("orchestrator", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dbPath := fs.String("db", getDefaultDBPath(), "Path to SQLite database")
	senderFlag := fs.String("sender", "", "Default sender name for messages (overrides BBS_AGENT_ID env var)")
	roleFlag := fs.String("role", "", "Agent role (overrides BBS_AGENT_ROLE env var)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Determine sender: flag > env var > default ("orchestrator")
	sender := *senderFlag
	if sender == "" {
		sender = os.Getenv("BBS_AGENT_ID")
		if sender == "" {
			sender = "orchestrator"
		}
	}

	// Determine role: flag > env var > default ("orchestrator")
	role := *roleFlag
	if role == "" {
		role = os.Getenv("BBS_AGENT_ROLE")
		if role == "" {
			role = "orchestrator"
		}
	}

	database, err := db.Open(*dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	// Register/update agent presence
	if err := database.UpsertAgentPresence(sender, role); err != nil {
		fmt.Fprintf(stderr, "Warning: failed to register agent presence: %v\n", err)
	}

	fmt.Fprintf(stderr, "Orchestrator started with database: %s (agent: %s, role: %s)\n", *dbPath, sender, role)

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(stderr, "\nReceived interrupt signal, shutting down...")
		cancel()
	}()

	orchestrator := hub.NewOrchestrator(database, hub.DefaultConfig())
	if err := orchestrator.Start(ctx); err != nil && err != context.Canceled {
		return fmt.Errorf("orchestrator error: %w", err)
	}

	fmt.Fprintln(stderr, "Orchestrator stopped")
	return nil
}
