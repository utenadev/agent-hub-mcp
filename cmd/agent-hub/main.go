package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yklcs/agent-hub-mcp/internal/db"
	"github.com/yklcs/agent-hub-mcp/internal/hub"
	"github.com/yklcs/agent-hub-mcp/internal/mcp"
)

// App holds dependencies for testing
type App struct {
	Logger   *log.Logger
	ExitFunc func(int)
}

// NewApp creates a new App with defaults
func NewApp() *App {
	return &App{
		Logger:   log.New(os.Stderr, "", log.LstdFlags),
		ExitFunc: os.Exit,
	}
}

func main() {
	app := NewApp()
	if err := app.Run(os.Args, os.Stdin, os.Stdout, os.Stderr); err != nil {
		app.Logger.Printf("Error: %v\n", err)
		app.ExitFunc(1)
	}
}

// Run executes the application with given arguments and IO
func (a *App) Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("Usage: agent-hub <command> [args...]\nCommands: serve, orchestrator")
	}

	command := args[1]

	switch command {
	case "serve":
		return a.runServe(args[2:])
	case "orchestrator":
		return a.runOrchestrator(args[2:])
	default:
		return fmt.Errorf("Unknown command: %s\nCommands: serve, orchestrator", command)
	}
}

// runServe starts the MCP server.
func (a *App) runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Suppress flag errors during testing
	dbPath := fs.String("db", "agent-hub.db", "Path to SQLite database")
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

	// Register/update agent presence
	if err := database.UpsertAgentPresence(sender, role); err != nil {
		a.Logger.Printf("Warning: failed to register agent presence: %v", err)
	}

	a.Logger.Printf("Database opened: %s", *dbPath)
	a.Logger.Printf("Agent: name=%s, role=%s", sender, role)

	srv := mcp.NewServer(database, sender, role)

	if *sseAddr != "" {
		a.Logger.Printf("Starting MCP server on SSE %s...", *sseAddr)
		if err := srv.ServeSSE(*sseAddr); err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	} else {
		a.Logger.Println("Starting MCP server on stdio...")
		if err := srv.Serve(); err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	return nil
}

// runOrchestrator starts the orchestrator monitor.
func (a *App) runOrchestrator(args []string) error {
	fs := flag.NewFlagSet("orchestrator", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dbPath := fs.String("db", "agent-hub.db", "Path to SQLite database")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	database, err := db.Open(*dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	fmt.Printf("Orchestrator started with database: %s\n", *dbPath)

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		a.Logger.Println("\nReceived interrupt signal, shutting down...")
		cancel()
	}()

	orchestrator := hub.NewOrchestrator(database, hub.DefaultConfig())
	if err := orchestrator.Start(ctx); err != nil && err != context.Canceled {
		return fmt.Errorf("orchestrator error: %w", err)
	}

	a.Logger.Println("Orchestrator stopped")
	return nil
}
