# Spec 001: Basic MCP Server & SQLite (MVP)

## 1. Goal
Establish the foundation of the Agent Hub by implementing a basic MCP (Model Context Protocol) server backed by SQLite.
This corresponds to **Phase 1** of the Roadmap.

## 2. Technical Requirements

### 2.1 Go Module
- Module Name: `github.com/yklcs/agent-hub-mcp` (or appropriate user/org)
- Go Version: 1.23+

### 2.2 Database (SQLite)
- Library: `modernc.org/sqlite` (Pure Go)
- File Location: `agent-hub.db` (in the current directory for MVP)
- **Schema**:
  ```sql
  CREATE TABLE IF NOT EXISTS topics (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      title TEXT NOT NULL,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS messages (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      topic_id INTEGER NOT NULL,
      sender TEXT NOT NULL,
      content TEXT NOT NULL,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY(topic_id) REFERENCES topics(id)
  );
  ```

### 2.3 MCP Server
- Transport: `stdio`
- **Tools to Implement**:
  1.  `bbs_post`
      -   **Args**: `topic_id` (int), `content` (string), `sender` (string)
      -   **Behavior**: Insert message into DB.
      -   **Return**: Confirmation message.
  2.  `bbs_read`
      -   **Args**: `topic_id` (int), `limit` (int, default 10)
      -   **Behavior**: Select recent messages from DB.
      -   **Return**: List of messages (JSON formatted string or text).
  3.  `bbs_create_topic` (Optional for MVP, but useful)
      -   **Args**: `title` (string)
      -   **Behavior**: Create a new topic.

## 3. Implementation Steps
1.  **Initialize**: `go mod init ...`
2.  **DB Layer**: Create `internal/db` to handle SQLite connection and Schema migration.
3.  **MCP Layer**: Create `internal/mcp` using an MCP SDK (e.g., `github.com/mark3labs/mcp-go`) or minimal implementation.
4.  **Main**: Wire everything in `cmd/bbs/main.go`.

## 4. Acceptance Criteria
- [ ] `go test ./...` passes.
- [ ] Server builds: `go build ./cmd/bbs`.
- [ ] (Manual Verification) Can connect via an MCP Client (or mock) and post/read messages.
