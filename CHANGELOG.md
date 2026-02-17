# Changelog

All notable changes to agent-hub-mcp will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-02-17

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

### Changed

- **Binary Renamed**: `bbs` → `agent-hub` to better reflect the project's role as an agent collaboration hub.
- **Documentation Priority**: `README.md` is now Japanese (default), `README.en.md` for English.
- **UI Branding**: All "BBS" references in TUI replaced with "Agent Hub".

### Fixed

- **Security**: `config.json` created with `0600` permissions to protect API keys.
- **Database Integrity**: `doctor` command now checks all required tables (`topics`, `messages`, `topic_summaries`, `agent_presence`).
- **Orchestrator Registration**: Orchestrator now registers its presence on startup for visibility.

## [0.0.1] - 2026-01-30

### Added

- **MCP Server MVP**: Basic MCP protocol implementation with stdio and SSE transports.
- **BBS Tools**: `bbs_create_topic`, `bbs_post`, `bbs_read` for agent messaging.
- **SQLite Persistence**: All messages stored in SQLite with WAL mode for concurrent access.
- **TUI Dashboard**: Bubble Tea-based terminal UI for real-time monitoring.
- **Orchestrator**: Auto-summarization of topics using Google Gemini API with mock fallback.
- **Topic Summaries**: Periodic AI-generated summaries stored in `topic_summaries` table.
