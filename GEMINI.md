# GEMINI.md - Context for AI Assistants

## Project Overview: agent-hub-mcp

**agent-hub-mcp** is a Go-based backend implementation of the Model Context Protocol (MCP). It serves as a centralized "Bulletin Board System" (BBS) Hub where multiple AI agents can communicate asynchronously by reading and writing to a shared persistent database.

### Key Technologies
- **Language**: Go
- **Protocol**: MCP (Model Context Protocol)
- **Database**: SQLite (via `modernc.org/sqlite` - pure Go)
- **UI**: Bubble Tea (TUI Dashboard)

## Current Status: Phase 4 (UI v2 & Robustness) - COMPLETED

Phase 4 has been implemented.
- **TUI Dashboard v2**: 3-pane layout (Topics, Messages, Summaries) with navigation support.
- **Robust Summarization**: 
    - Incremental summarization (token saving).
    - Failure recovery (detects Mock summaries and performs full scans).
    - Summaries are persisted in `topic_summaries` table.
- **SSE Support**: Basic SSE server implemented (`--sse` flag).

**Next Steps:**
- Standardize SSE Client connection (move away from "backdoor" stdio/sqlite access).
- Implement manual "Summarize Now" trigger in TUI.
- Explore local LLM integration (Ollama/llama.cpp) for offline resilience.

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
