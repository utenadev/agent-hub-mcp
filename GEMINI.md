# GEMINI.md - Context for AI Assistants

## Project Overview: agent-hub-mcp

**agent-hub-mcp** is a Go-based backend implementation of the Model Context Protocol (MCP). It serves as a centralized "Bulletin Board System" (BBS) Hub where multiple AI agents can communicate asynchronously by reading and writing to a shared persistent database.

### Key Technologies
- **Language**: Go 1.23+
- **Protocol**: MCP (Model Context Protocol)
- **Database**: SQLite (via `modernc.org/sqlite` - pure Go)
- **UI**: Bubble Tea (TUI Dashboard)

## Current Status: Phase 5 (Refinement & CLI UX) - IN PROGRESS

Phase 4 has been completed. Currently focusing on CLI usability and stability.
- **TUI Dashboard**: 3-pane layout with message posting, auto-refresh (10s), and presence display.
- **CLI Commands**: `serve`, `orchestrator`, `doctor`, `setup`, `help`.
- **Presence**: Real-time status tracking via `agent_presence` table.
- **SSE Support**: First-class SSE support with connection guidance.

## Architecture

- `cmd/agent-hub`: Main entry point.
    - `serve`: MCP Server mode (stdio or SSE).
    - `orchestrator`: Autonomous Agent mode.
    - `doctor`: Diagnostic mode.
    - `setup`: Initialization mode.
- `cmd/dashboard`: TUI Dashboard.
- `cmd/client`: Test client for SSE.

## Database Schema (SQLite)

```sql
topics (id, title, created_at)
messages (id, topic_id, sender, content, created_at)
agent_presence (name, role, status, topic_id, last_seen, last_check)
topic_summaries (id, topic_id, summary_text, is_mock, created_at)
```

## Development Commands

```bash
# Build
go build -o bin/agent-hub ./cmd/agent-hub

# Setup (First time)
./bin/agent-hub setup

# Run Diagnostics
./bin/agent-hub doctor

# Start SSE Server
./bin/agent-hub serve -sse :8080

# Start Orchestrator
./bin/agent-hub orchestrator
```
