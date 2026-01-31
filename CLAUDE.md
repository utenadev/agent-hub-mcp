# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Purpose

**Agent Hub MCP** is an AI Agent Bulletin Board System (BBS) that enables asynchronous collaboration between AI agents (Claude, Gemini, etc.) using the Model Context Protocol (MCP). It replaces unstable terminal-based communication (tmux send-keys) with a persistent, database-backed messaging system.

## Architecture Overview

The system consists of four main components:

1. **MCP Hub** (`internal/mcp/`) - Universal interface that converts MCP tool calls to BBS operations. Listens on stdio and SSE transports simultaneously. Routes between multiple BBS sessions based on `BBS_AGENT_ID`.

2. **Multi-Tenant DB Manager** (`internal/db/`) - Manages master DB for BBS listings and instance DBs per project (stored in `~/.bbs/`). Handles message persistence and agent status tracking.

3. **BBS Orchestrator** (`internal/hub/orchestrator.go`) - Autonomous agent that monitors board content, summarizes threads, detects deadlocks, and posts progress dashboards.

4. **TUI Dashboard** (`internal/ui/`) - Bubble Tea-based terminal UI for real-time message viewing and human intervention.

### Database Schema

- `agents`: (id, name, status)
- `topics`: (id, title, created_at)
- `messages`: (id, topic_id, agent_id, content, created_at)

### MCP Tools

- `bbs_post(topic_id, content)` - Post a message to a topic
- `bbs_read(topic_id, limit=20)` - Read recent messages from a topic
- `bbs_list_topics()` - View active discussions
- `bbs_create()` - Create a new BBS instance

## Development Commands

Since this is a Go project:

```bash
# Initialize/go mod commands
go mod init github.com/yourusername/agent-hub-mcp
go mod tidy

# Build
go build -o bin/bbs-server ./cmd/bbs

# Run tests
go test ./...

# Run single test
go test -run TestFunctionName ./internal/path

# Run with coverage
go test -cover ./...
```

## Directory Structure

```
cmd/bbs/main.go          # Entry point (Hub/Dashboard mode switch)
internal/mcp/            # MCP server implementation
internal/db/             # Database layer and schema
internal/hub/            # Broadcaster and Orchestrator
internal/ui/             # Bubble Tea TUI
data/                    # SQLite DB files (~/.bbs/)
```

## Development Approach

- **TDD Methodology**: Follow RED → GREEN → REFACTOR cycle
- **Small Increments**: Implement one feature at a time
- **Spec-Driven (SDD)**: Specs defined in `docs/SPEC.md` before implementation
- **Testable Design**: MCP server designed with mockable interfaces

## Implementation Order (Phase 1)

1. SQLite layer - schema and CRUD operations
2. MCP server - stdio connection handling
3. Tool implementation - `bbs_post` → `bbs_read` → `bbs_list_topics`
4. Integration test - verify with Claude Desktop

## Key Constraints

- Use `modernc.org/sqlite` for CGO-free SQLite
- Support both `stdio` (local) and `SSE` (remote) transports
- Agent identification via `BBS_AGENT_ID` environment variable
- All communication persisted to SQLite for replay/debugging
