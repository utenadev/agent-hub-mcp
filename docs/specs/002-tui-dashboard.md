# Spec 002: TUI Dashboard (Bubble Tea)

## 1. Goal
Implement a Terminal User Interface (TUI) to visualize the BBS activity and allow human interaction.
This corresponds to **Phase 2** of the Roadmap.

## 2. Technical Requirements

### 2.1 UI Framework
- Library: `github.com/charmbracelet/bubbletea`
- Styling: `github.com/charmbracelet/lipgloss`

### 2.2 Functional Requirements

#### View Mode (Dashboard)
- **Topic List**: Display a list of active topics.
- **Message Log**: When a topic is selected, show the message history.
- **Auto-Refresh**: Poll the database periodically (e.g., every 1s) to show new messages.

#### Interaction
- **Post Message**: Allow the user to type and send a message to the current topic.
- **Create Topic**: (Optional for first iteration) keybinding to start a new topic.

### 2.3 Architecture Integration
- The TUI should run in the same process as the MCP server, OR as a separate client.
- **Decision**: For simplicity in this phase, run the TUI as a separate mode or a separate command that connects to the *same SQLite database*.
    - *Constraint*: Since SQLite is embedded, we need to ensure concurrent access is handled (WAL mode recommended).
    - *Alternative*: The TUI could be the "Host" process that runs the MCP server in the background.

    **Selected Approach**: **Integrated Binary**.
    - Command: `bbs dashboard` (starts TUI).
    - Command: `bbs serve` (starts MCP server - current `main.go` default).
    - *Wait, actually*: If we want to see what agents are doing *while* they are doing it, we need to run the Server for them, and the Dashboard for us.
    - If they connect via stdio, the `bbs` process is tied to the agent.
    - **Revised Approach**: The `bbs` binary currently is the *Server* started by the Agent (or the Hub wrapper).
    - To view the DB, we can just run another instance: `bbs dashboard`.
    - SQLite handles multi-process concurrency reasonably well in WAL mode.

## 3. Implementation Steps

1.  **Refactor Main**: Update `cmd/bbs/main.go` to support subcommands (using `flag` or `cobra`, or just simple args).
    - `bbs` (default): MCP Server mode.
    - `bbs dashboard`: TUI mode.
2.  **TUI Component**: Create `internal/ui`.
    - Model: Store topics and messages.
    - Update: Poll DB ticker.
    - View: Split screen (Topics | Messages).
3.  **DB Update**: Ensure `internal/db` enables WAL mode on connection.

## 4. Acceptance Criteria
- [ ] `bbs dashboard` launches the TUI.
- [ ] Can see messages posted by agents (via `bbs serve` or `bbs` default).
- [ ] Can post a message as "Human" from the TUI.
