package main

import (
	"fmt"
	"io"
)

// runHelp displays usage information.
func (a *App) runHelp(stdout io.Writer) {
	fmt.Fprintln(stdout, "Agent Hub MCP - Multi-Agent Collaboration BBS")
	fmt.Fprintln(stdout, "\nUsage:")
	fmt.Fprintln(stdout, "  agent-hub <command> [args...]")
	fmt.Fprintln(stdout, "\nCommands:")
	fmt.Fprintln(stdout, "  serve         Start the MCP server (stdio or SSE mode)")
	fmt.Fprintln(stdout, "  orchestrator  Start the autonomous monitor/summarizer")
	fmt.Fprintln(stdout, "  doctor        Run system diagnostics")
	fmt.Fprintln(stdout, "  setup         Initialize database and configuration")
	fmt.Fprintln(stdout, "  help          Show this help message")
	fmt.Fprintln(stdout, "\nGlobal Flags (available for most commands):")
	fmt.Fprintln(stdout, "  -db string    Path to SQLite database (default: "+getDefaultDBPath()+")")
	fmt.Fprintln(stdout, "\nServe Flags:")
	fmt.Fprintln(stdout, "  -sse string   Enable SSE mode on address (e.g., :8080)")
	fmt.Fprintln(stdout, "  -sender name  Default sender name for messages")
	fmt.Fprintln(stdout, "  -role role    Agent role")
	fmt.Fprintln(stdout, "\nSSE Connection Example:")
	fmt.Fprintln(stdout, "  When running with '-sse :8080', connect your MCP client to:")
	fmt.Fprintln(stdout, "  http://localhost:8080/sse")
}
