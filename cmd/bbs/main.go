package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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

	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatal(err)
	}

	database, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	log.Printf("Database opened: %s", *dbPath)

	server := mcp.NewServer(database)
	log.Println("Starting MCP server on stdio...")
	if err := server.Serve(); err != nil {
		log.Fatalf("Server error: %v", err)
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

	orchestrator := hub.NewOrchestrator(database, hub.DefaultConfig())
	if err := orchestrator.Start(); err != nil {
		log.Fatalf("Orchestrator error: %v", err)
	}
}
