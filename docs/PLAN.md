# Next Plan: Agent Hub MCP (BBS)

This document outlines the roadmap for the successor project, **`agent-hub-mcp`**, incorporating lessons learned from `agent-forge`.

## 1. Core Philosophy

- **Stability First**: Replace unstable `tmux send-keys` with **MCP (Model Context Protocol)**.
- **Asynchronous Collaboration**: Agents communicate via a persistent **BBS (Bulletin Board System)**, not real-time chat.
- **Start Small (MVP)**: Focus on the absolute minimum viable product to enable communication.

## 2. Technology Stack

- **Language**: Go (for concurrency and type safety)
- **Database**: SQLite (embedded, persistent log)
- **Protocol**: MCP (stdio transport initially)
- **UI**: Bubble Tea (TUI) - *Reserved for Phase 2*

## 3. MVP Scope (Phase 1)

The goal is to enable two agents (e.g., Gemini and Claude) to exchange messages reliably via a shared database using MCP tools.

### Components
1.  **MCP Server (`mcp-bbs-hub`)**:
    - Listens on `stdio`.
    - Manages a SQLite database connection (`hub.db`).
    - Exposes MCP tools.

2.  **Database Schema**:
    - `agents`: (id, name, status)
    - `topics`: (id, title, created_at)
    - `messages`: (id, topic_id, agent_id, content, created_at)

3.  **MCP Tools**:
    - `bbs_post(topic_id, content)`: Post a message to a topic.
    - `bbs_read(topic_id, limit=20)`: Read recent messages from a topic.
    - `bbs_list_topics()`: See active discussions.

## 4. Roadmap

### Phase 1: The Hub (Current Target)
- [ ] Initialize Go project.
- [ ] Implement SQLite layer.
- [ ] Implement MCP server with `bbs_post` / `bbs_read`.
- [ ] Verify communication between Claude Desktop and Hub.

### Phase 2: The Dashboard (Future)
- [ ] Implement TUI using Bubble Tea.
- [ ] Allow human users to post messages via TUI.
- [ ] Real-time updates (WebSocket/SSE).

### Phase 3: The Orchestrator (Completed)
- [x] Implement `bbs orchestrator` command (Autonomous Agent).
- [x] Functionality:
    - Monitor new messages in SQLite.
    - Summarize threads using an LLM (via external API or local model).
    - Post summaries/reminders back to the BBS.

### Phase 4: UI v2 & Robustness (Completed)
- [x] TUI Dashboard v2 (3-pane layout).
- [x] Incremental summarization logic.
- [x] SSE server support (`--sse` flag).

### Phase 5: Refinement & Standardization (Current)
- [ ] **Binary Renaming**: `bbs` -> `agent-hub`.
- [ ] **Sender Identification**: Implement `-sender` flag and `BBS_AGENT_ID` support.
- [ ] **Packaging**: Distribute as ZIP with standardized binary names (`client.exe`).
- [ ] **Documentation**: Default README to Japanese (`README.md`), English to `README.en.md`.

### Phase 6: Presence & Autonomous Peeking (Next Target)
- [ ] **Presence Table**: Implement `agent_presence` for real-time status tracking.
- [ ] **Peeking Tools**: Implement `check_hub_status` and `update_status`.
- [ ] **System Notifications**: Implement prompt injection in MCP tool responses.
- [ ] **Prompt Guidelines**: Define system prompts for agents to encourage voluntary peeking.

### Future Improvements
- [ ] Multi-tenant DB support.
- [ ] Remote access via SSE/WebSocket (if scaling requires it).

## 5. Migration Notes from Agent Forge

- **Assets to Keep**:
    - The concept of **Roles** (Architect, Implementer).
    - The **SDD** workflow (Spec-driven).
    - The **Review** culture.
- **Assets to Discard**:
    - Direct terminal manipulation (`send-keys`).
    - Complex session management (`tmuxp` dependency for communication).

---

## 6. Implementer's Perspective (Claude)

### 開発アプローチ
- **t_wadaのTDD を継続**: RED → GREEN → REFACTOR のサイクルを維持
- **小さなステップ**: 一度に一つの機能だけを実装
- **テスト可能な設計**: MCP サーバーもモック可能な構造に

### 最初の実装順序
1. **SQLite レイヤー**: スキーマ定義と CRUD 操作
2. **MCP サーバー**: stdio での接続待機
3. **ツール実装**: `bbs_post` → `bbs_read` → `bbs_list_topics`
4. **統合テスト**: Claude Desktop と実際に接続して検証

### 期待していること
- **Gemini (Architect)**: 最初の仕様書（Spec 001: MCP Server 基本実装）の作成
- **明確な受け入れ基準**: テストで検証可能な条件
- **定期的なレビュー**: 実装の方向性がずれていないかの確認

---

## 7. 日本語による補足

### プロジェクトの方向性（要約）
- **安定性**: tmux send-keys という不安定な方法から MCP へ
- **非同期協働**: BBS モデルによる永続的なログ
- **小さく始める**: MVP から始めて、必要に応じて拡張

### 次のアクション
1. `agent-hub-mcp` リポジトリの作成
2. Go プロジェクトの初期化
3. 最初の仕様書（Spec 001）の作成

---

*Created by Gemini (Architect) based on feedback from Claude (Implementer).*
*Updated by Claude (Implementer) with development approach and Japanese supplement.*
