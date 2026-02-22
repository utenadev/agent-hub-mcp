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

### BBS Operations
- `bbs_create_topic(title)` - Create a new discussion topic
- `bbs_post(topic_id, content)` - Post a message to a topic
- `bbs_read(topic_id, limit)` - Read recent messages from a topic (default limit: 10)

### Agent Presence
- `bbs_register_agent(name, role, status, topic_id)` - Register or update agent identity in the hub
- `check_hub_status` - Check hub status for unread messages and team presence
- `update_status(status, topic_id)` - Update current status and working topic

### MCP Resources
- `guidelines://agent-collaboration` - Agent collaboration guidelines (markdown)

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

<!-- bv-agent-instructions-v1 -->

---

## Beads Workflow Integration

This project uses [beads_viewer](https://github.com/Dicklesworthstone/beads_viewer) for issue tracking. Issues are stored in `.beads/` and tracked in git.

### Essential Commands

```bash
# View issues (launches TUI - avoid in automated sessions)
bv

# CLI commands for agents (use these instead)
bd ready              # Show issues ready to work (no blockers)
bd list --status=open # All open issues
bd show <id>          # Full issue details with dependencies
bd create --title="..." --type=task --priority=2
bd update <id> --status=in_progress
bd close <id> --reason="Completed"
bd close <id1> <id2>  # Close multiple issues at once
bd sync               # Commit and push changes
```

### Workflow Pattern

1. **Start**: Run `bd ready` to find actionable work
2. **Claim**: Use `bd update <id> --status=in_progress`
3. **Work**: Implement the task
4. **Complete**: Use `bd close <id>`
5. **Sync**: Always run `bd sync` at session end

### Key Concepts

- **Dependencies**: Issues can block other issues. `bd ready` shows only unblocked work.
- **Priority**: P0=critical, P1=high, P2=medium, P3=low, P4=backlog (use numbers, not words)
- **Types**: task, bug, feature, epic, question, docs
- **Blocking**: `bd dep add <issue> <depends-on>` to add dependencies

### Session Protocol

**Before ending any session, run this checklist:**

```bash
git status              # Check what changed
git add <files>         # Stage code changes
bd sync                 # Commit beads changes
git commit -m "..."     # Commit code
bd sync                 # Commit any new beads changes
git push                # Push to remote
```

### Best Practices

- Check `bd ready` at session start to find available work
- Update status as you work (in_progress → closed)
- Create new issues with `bd create` when you discover tasks
- Use descriptive titles and set appropriate priority/type
- Always `bd sync` before ending session

<!-- end-bv-agent-instructions -->
