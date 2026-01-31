package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yklcs/agent-hub-mcp/internal/db"
	"github.com/yklcs/agent-hub-mcp/internal/hub"
	"github.com/yklcs/agent-hub-mcp/internal/mcp"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: bbs <command> [args...]\nCommands: serve, orchestrator")
	}

	command := os.Args[1]

	switch command {
	case "serve":
		runServe()
	case "orchestrator":
		runOrchestrator()
	default:
		log.Fatalf("Unknown command: %s\nCommands: serve, orchestrator", command)
	}
}

// runServe starts the MCP server.
func runServe() {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	dbPath := fs.String("db", "agent-hub.db", "Path to SQLite database")
	sseAddr := fs.String("sse", "", "Enable SSE mode on address (e.g., :8080)")

	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatal(err)
	}

	database, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	log.Printf("Database opened: %s", *dbPath)

	srv := mcp.NewServer(database)

	if *sseAddr != "" {
		log.Printf("Starting MCP server on SSE %s...", *sseAddr)
		if err := srv.ServeSSE(*sseAddr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		log.Println("Starting MCP server on stdio...")
		if err := srv.Serve(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}

// runOrchestrator starts the orchestrator monitor.
func runOrchestrator() {
	fs := flag.NewFlagSet("orchestrator", flag.ExitOnError)
	dbPath := fs.String("db", "agent-hub.db", "Path to SQLite database")

	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatal(err)
	}

	database, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
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
		log.Println("\nReceived interrupt signal, shutting down...")
		cancel()
	}()

	orchestrator := hub.NewOrchestrator(database, hub.DefaultConfig())
	if err := orchestrator.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Orchestrator error: %v", err)
	}

	log.Println("Orchestrator stopped")
}
