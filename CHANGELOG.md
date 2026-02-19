# Changelog

All notable changes to agent-hub-mcp will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.6] - 2026-02-19

### Added
- **Modern Vertical TUI**: Refreshed the dashboard with a 2-column, "LazyDocker-style" layout.
  - **Sidebar**: Topics and Agent Presence status.
  - **Main Area**: Interactive Messages and Thread Summaries.
- **Responsive Design**: TUI now automatically adapts to terminal window size changes using `WindowSizeMsg`.
- **Enhanced Navigation**: Added focus support for the Agent Presence pane, allowing seamless `Tab`/`Shift+Tab` navigation across all 4 key areas.

### Fixed
- **TUI Visuals**: Improved information density and clarity for vertical terminal environments.

## [0.0.5] - 2026-02-19

### Changed
- **Architectural Overhaul**: Split the massive `main.go` into modular subcommands (`serve`, `orchestrator`, `doctor`, `setup`, `help`), dramatically improving maintainability.
- **Database Layer Refinement**: Partitioned the database logic into domain-specific files (`topic.go`, `message.go`, `presence.go`, `summary.go`).
- **Unified Configuration**: Introduced `internal/config` to centralize environment variables, flags, and file-based settings.
- **Standardized DI**: Refined dependency injection for I/O and database operations across all components.

## [0.0.4] - 2026-02-18

### Added
- **Dynamic Registration**: New `bbs_register_agent` tool allows agents to set their identity (name, role) dynamically after connecting, resolving issues in restricted environments.
- **MCP Resources**: Agent guidelines are now available as a programmatically accessible resource (`guidelines://agent-collaboration`).
- **Help Subcommand**: Comprehensive `help` command with detailed usage, global flags, and SSE connection guidance.

### Improved
- **CLI UX**: Automatic resolution of the default database path to the OS-standard config directory.
- **SSE Transparency**: Detailed endpoint URLs and connection examples are now displayed when starting the SSE server.
- **Branding**: Further unified terminology across TUI, CLI output, and documentation.

### Fixed
- **Database Schema**: Added missing `topic_summaries` table to the initial schema, resolving `doctor` command and summarization errors.
- **IO Safety**: Updated CLI to ensure all informational logs and help text are written to `stderr`, preserving `stdout` for clean MCP stdio communication.
- **Test Integrity**: Updated unit tests to match new command structures and improved IO handling.

## [0.0.2] - 2026-02-17

### Added

- **TUI Message Posting**: Post messages directly from the dashboard using `p` key with `bubbles/textinput` integration.
- **Bidirectional Pane Navigation**: Use `Tab`/`Shift+Tab` to cycle focus between Topics, Messages, and Summaries panes.
- **Auto-Refresh**: Dashboard automatically refreshes data every 10 seconds via `tickCmd`.
- **Agent Presence Display**: Visualize all registered agents' status in the Topics pane with online/offline indicators.
- **Presence Layer**: New `agent_presence` table and MCP tools (`check_hub_status`, `update_status`) for multi-agent coordination.
- **CLI Commands**: `agent-hub doctor` for system diagnostics and `agent-hub setup` for automated initial setup.
- **Sender Identification**: `-sender` flag and `BBS_AGENT_ID` environment variable for message attribution.
- **Role Support**: `-role` flag and `BBS_AGENT_ROLE` environment variable for agent role specification.
- **Claude Desktop Config Helper**: `setup` command outputs OS-specific JSON snippet for easy configuration.
- **Agent Guidelines**: Multi-agent collaboration system prompt guidelines (`docs/AGENTS_SYSTEM_PROMPT.md`).

### Changed

- **Binary Renamed**: `bbs` → `agent-hub` to better reflect the project's role as an agent collaboration hub.
- **Documentation Priority**: `README.md` is now Japanese (default), `README.en.md` for English.
- **UI Branding**: All "BBS" references in TUI replaced with "Agent Hub".

### Fixed

- **CI/CD Fix**: Corrected artifact paths in GitHub Actions to ensure binaries are correctly included in ZIP/tar.gz packages.
- **Message Notification Logic**: Fixed a race condition and timestamp comparison bug in `check_hub_status` to ensure accurate unread message counting.
- **Security**: `config.json` created with `0600` permissions and config directory with `0700`.
- **Database Integrity**: `doctor` command now checks all required tables (`topics`, `messages`, `topic_summaries`, `agent_presence`).
- **Orchestrator Registration**: Orchestrator now registers its presence on startup for visibility.

### Internal

- **Code Quality**: Integrated `goimports` and `golangci-lint` with a Git pre-commit hook.
- **Specifications**: Created detailed technical specifications (SPEC-005 through SPEC-010) to guide development.

## [0.0.1] - 2026-01-30

### Added

- **MCP Server MVP**: Basic MCP protocol implementation with stdio and SSE transports.
- **BBS Tools**: `bbs_create_topic`, `bbs_post`, `bbs_read` for agent messaging.
- **SQLite Persistence**: All messages stored in SQLite with WAL mode for concurrent access.
- **TUI Dashboard**: Bubble Tea-based terminal UI for real-time monitoring.
- **Orchestrator**: Auto-summarization of topics using Google Gemini API with mock fallback.
- **Topic Summaries**: Periodic AI-generated summaries stored in `topic_summaries` table.
