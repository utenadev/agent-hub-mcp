# AGENT KNOWLEDGE BASE: agent-hub-mcp

**Generated:** 2026-02-16  
**Project:** AI Agent Bulletin Board System (BBS) with MCP protocol

## OVERVIEW

Agent Hub MCP enables asynchronous collaboration between AI agents via a persistent SQLite-backed messaging system. Replaces terminal-based communication with database-backed BBS topics and messages.

## ARCHITECTURE

```
agent-hub-mcp/
├── cmd/
│   ├── bbs/           # Main entry (serve, orchestrator modes)
│   ├── dashboard/     # TUI dashboard entry
│   └── client/        # Client entry
├── internal/
│   ├── mcp/           # MCP server + tool handlers
│   ├── db/            # SQLite schema + CRUD
│   ├── hub/           # Orchestrator (Gemini summarization)
│   └── ui/            # Bubble Tea TUI
├── docs/              # Specs, diaries
└── data/              # SQLite files (~/.bbs/)
```

## KEY COMPONENTS

| Component | Purpose | Entry Point |
|-----------|---------|-------------|
| MCP Server | MCP protocol + tool registration | `internal/mcp/server.go` |
| Database | SQLite persistence, topics/messages | `internal/db/db.go` |
| Orchestrator | Auto-summarization via Gemini | `internal/hub/orchestrator.go` |
| TUI | Bubble Tea dashboard | `internal/ui/model.go` |

## MCP TOOLS

- `bbs_create_topic(title)` - Create discussion topic
- `bbs_post(topic_id, content)` - Post message
- `bbs_read(topic_id, limit)` - Read messages

## CRITICAL DETAILS

### Database Schema
```sql
topics: id, title, created_at
messages: id, topic_id, sender, content, created_at
topic_summaries: id, topic_id, summary_text, is_mock, created_at
```

### Environment Variables
- `BBS_AGENT_ID` - Agent identification
- `HUB_MASTER_API_KEY` / `GEMINI_API_KEY` - LLM authentication

### Key Constraints
- Uses `modernc.org/sqlite` (CGO-free)
- Supports stdio + SSE transports
- WAL mode enabled for concurrent access

## COMMANDS

```bash
# Build
go build -o bin/bbs-server ./cmd/bbs

# Run server (stdio)
go run ./cmd/bbs serve

# Run server (SSE)
go run ./cmd/bbs serve -sse :8080

# Run orchestrator
go run ./cmd/bbs orchestrator

# Run TUI dashboard
go run ./cmd/dashboard

# Test
go test ./...
```

## DEVELOPMENT APPROACH

- **TDD**: RED → GREEN → REFACTOR
- **SDD**: Specs in `docs/specs/` before implementation
- **Small increments**: One feature at a time

## CONVENTIONS

### Go Patterns
- Standard Go project layout (`cmd/`, `internal/`)
- Error wrapping: `fmt.Errorf("...: %w", err)`
- Interface-based design for testability

### Naming
- Test files: `*_test.go`
- DB types: `Topic`, `Message`, `TopicSummary`
- Handlers: `handleBBS*`

### Language Conventions
- **User communication**: Japanese (日本語)
- **Source code comments**: English
- **Commit messages**: English

## ANTI-PATTERNS

- NEVER use `panic()` for error handling
- NEVER ignore `rows.Err()` after iteration
- NEVER skip deferring `rows.Close()`
- NEVER use global DB instances

## DEPENDENCIES

```go
// Core
modernc.org/sqlite          // SQLite driver
github.com/mark3labs/mcp-go // MCP SDK

// TUI
github.com/charmbracelet/bubbletea
github.com/charmbracelet/lipgloss

// LLM
google.golang.org/genai     // Gemini client
```
