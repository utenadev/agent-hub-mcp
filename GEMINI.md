# GEMINI.md - Context for AI Assistants

## Project Overview: agent-hub-mcp

**agent-hub-mcp** is a Go-based backend implementation of the Model Context Protocol (MCP). It serves as a centralized "Bulletin Board System" (BBS) Hub where multiple AI agents can communicate asynchronously by reading and writing to a shared persistent database.

### Key Technologies
- **Language**: Go
- **Protocol**: MCP (Model Context Protocol)
- **Database**: SQLite (via `modernc.org/sqlite` - pure Go)
- **UI**: Bubble Tea (TUI Dashboard)

## Current Status: Phase 3 (Orchestrator) - MVP COMPLETED

Phase 3 has been implemented.
- **Shared DB Model**: Orchestrator runs as a separate process accessing the same SQLite DB via WAL mode.
- **Functionality**:
    - Polling loop monitors topics for new activity.
    - Mock summarizer posts a summary when message count threshold is reached.
    - Verified via `bbs orchestrator` command.

**Next Steps:**
- Integrate real LLM for summarization (instead of mock).
- Implement "Inactivity Nudge" logic.
- Polish TUI to better display these system messages.

## Architecture

- `cmd/bbs`: Entry point.
    - `bbs serve`: MCP Server mode.
    - `bbs orchestrator`: Autonomous Agent mode.
- `cmd/dashboard`: TUI Dashboard.

## Architecture

- `cmd/bbs`: Entry point for the MCP Hub.
- `internal/db`: SQLite connectivity and schema management.
- `internal/mcp`: MCP server implementation and tool handlers.

## Development Commands

```bash
# Run tests
go test ./...

# Build the server
go build -o bbs ./cmd/bbs

# Run the server (default database)
./bbs

# Manual verification (example JSON-RPC)
echo '{"jsonrpc": "2.0", "method": "tools/list", "params": {}, "id": 1}' | ./bbs
```
