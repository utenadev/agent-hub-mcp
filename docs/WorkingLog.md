# Working Log

## 2026-01-30

### Phase 1: MVP Implementation (MCP Server & SQLite)

- **Gemini (Architect)**:
    - Created `docs/specs/001-mcp-server.md` defining the MVP scope.
    - Verified the implementation manually using JSON-RPC commands over stdio.
    - Confirmed `bbs_create_topic`, `bbs_post`, and `bbs_read` tools are working.

- **Claude (Implementer)**:
    - Initialized the Go project (`go mod init`).
    - Implemented `internal/db` (SQLite schema & CRUD operations).
    - Implemented `internal/mcp` (MCP Server & Tool Handlers).
    - Added comprehensive tests for both layers (`go test ./...` passing).
    - Resolved type mismatch errors in MCP SDK usage during implementation.

### Phase 2: TUI Dashboard Implementation

- **Gemini (Architect)**:
    - Created `docs/specs/002-tui-dashboard.md` defining the TUI requirements.
    - Verified the build and basic structure of the new `dashboard` command.

- **Claude (Implementer)**:
    - Implemented `internal/ui` using Bubble Tea and Lipgloss.
    - Created `cmd/dashboard/main.go` as a standalone binary for the TUI.
    - Implemented topic selection, message viewing, and basic refresh logic.
    - Ensured concurrent SQLite access compatibility.

### Phase 3: Orchestrator Implementation & LLM Integration

- **Decision**: 
    - Use Gemini API directly for summarization (Pattern A).
    - Model: `gemini-2.0-flash-lite-preview-02-05`.
    - API Key Priority: `HUB_MASTER_API_KEY` > `GEMINI_API_KEY`.
    - Updated `docs/specs/003-orchestrator.md` to reflect these details.

### Status
- **Phase 1 & 2 Complete**.
- **Phase 3 Planning Updated**.
- The system now has a functional backend (MCP) and a frontend (TUI Dashboard).
- Ready to start Phase 3 (Orchestrator) in the next session.

---

## 2026-01-31

### Phase 3: アーキテクチャ決定

#### 概要
Orchestrator の実装アーキテクチャについて決定を行った。

#### 決定事項
- **Shared DB Model (SQLite)**を採用
- SSE/WebSocket 実装は当面見送り
- Orchestrator は別プロセスとして実行し、同じ DB ファイルにアクセス

#### 変更ファイル
- `docs/PLAN.md` - Phase 3 のアーキテクチャ方針を更新

#### 理由
- SQLite WAL モードにより複数プロセスからの同時アクセスが可能
- シンプルな実装で早期に価値を提供するため
- SSE は将来の拡張として検討可能

#### 次のステップ
- Orchestrator プロセスの実装
- DB ポーリングによるメッセージ監視
- スレッド要約・デッドロック検出機能の追加

---

### Phase 3: Orchestrator 実装 (TDDアプローチ)

#### 概要
`docs/specs/003-orchestrator.md` に基づき、Orchestrator を実装。

#### REDフェーズ
- `internal/hub/orchestrator_test.go` でテストを作成:
  - `TestNewOrchestrator` - デフォルト設定の検証
  - `TestNewOrchestratorWithCustomConfig` - カスタム設定の検証
  - `TestInitializeTopics` - トピック初期化の検証
  - `TestCheckTopicWithNewMessages` - 新規メッセージ検出の検証
  - `TestMockSummarizer` - モック要約機能の検証
  - `TestGenerateSummary` - サマリーポストの検証
- テスト実行: **失敗** - 実装前の期待値

#### GREENフェーズ
- `internal/hub/orchestrator.go` を実装:
  - `Config` 構造体 - PollInterval, SummaryThreshold, InactivityTimeout
  - `Orchestrator` 構造体 - DB、設定、状態管理
  - `NewOrchestrator()` - オーケストレーターの作成
  - `Start()` - ポーリングループの開始
  - `initializeTopics()` - 既存トピックの追跡開始
  - `checkTopic()` - 単一トピックの監視
  - `generateSummary()` - サマリー生成とポスト
  - `mockSummarizer()` - モック要約機能（LLM統合用プレースホルダー）
- `cmd/bbs/main.go` を更新:
  - `serve` サブコマンドの追加（旧動作）
  - `orchestrator` サブコマンドの追加
  - `flag` パッケージによる CLI 引数処理
- テスト実行: **成功 - 全6テストパス**

#### 作成ファイル
- `internal/hub/orchestrator.go` - Orchestrator 実装（約200行）
- `internal/hub/orchestrator_test.go` - Orchestrator テスト（6テスト）

#### 変更ファイル
- `cmd/bbs/main.go` - サブコマンド構造に変更（serve, orchestrator）

#### テスト結果
```
=== RUN   TestNewOrchestrator
--- PASS: TestNewOrchestrator (0.02s)
=== RUN   TestNewOrchestratorWithCustomConfig
--- PASS: TestNewOrchestratorWithCustomConfig (0.02s)
=== RUN   TestInitializeTopics
--- PASS: TestInitializeTopics (0.05s)
=== RUN   TestCheckTopicWithNewMessages
--- PASS: TestCheckTopicWithNewMessages (0.06s)
=== RUN   TestMockSummarizer
--- PASS: TestMockSummarizer (0.02s)
=== RUN   TestGenerateSummary
--- PASS: TestGenerateSummary (0.05s)
PASS
ok  	github.com/yklcs/agent-hub-mcp/internal/hub	0.222s
```

#### 実装内容
1. **ポーリングループ**: 5秒ごとに DB をチェック
2. **メッセージ検出**: lastSeenMsgID で新しいメッセージを追跡
3. **サマリー生成**: 5メッセージごとに要約をポスト
4. **モック要約**: 送信者別メッセージ数を集計（LLM統合は将来実装）

#### 受入要件達成状況
- [x] `bbs orchestrator` がエラーなく実行される
- [x] 新規メッセージを検出してログ出力する
- [x] 5メッセージ後にサマリーをポストする
- [ ] アクティビティがないトピックにナッジを送る（未実装）

#### 次のステップ
- Inactivity Detection（ナッジ機能）の実装
- LLM 統合による本格的な要約機能
- Phase 4: エージェント間協働プロトコルの確立