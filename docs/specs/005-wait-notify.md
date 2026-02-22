# Spec 005: agent-hub-mcp Wait/Notify Integration

## 1. 概要
本ドキュメントは、`agent-hub-mcp` におけるバイナリ名の変更、送信者（Sender）の特定、全エージェント共通の通知待機ツール（`wait_notify`）、および CI/CD (GitHub Actions) における配布形式の変更を定義する。

## 2. 変更点詳細

### 2-1. バイナリ名とディレクトリ名の変更
- 実行バイナリ名を以下のように変更する。
  - `bbs` → `agent-hub` (メインサーバー)
  - `client-windows-amd64` → `client` (テスト用クライアント)
- ディレクトリ構造の変更:
  - `cmd/bbs` → `cmd/agent-hub`
- ヘルプメッセージやUsage内の `bbs` 表記を `agent-hub` に統一する。

### 2-2. 送信者 (Sender) の特定
- `agent-hub serve` 起動時に `--sender` フラグを受け取る。
- MCPサーバー起動時に `defaultSender` として保持する。
- `bbs_post` ツール実行時、引数に `sender` が明示されなかった場合、起動時の `defaultSender` を使用してメッセージを記録する。

### 2-3. `wait_notify` ツールの実装 (ロングポーリング)
- すべてのエージェントが「待ち」の状態に入れるよう、MCPツール `wait_notify` を追加する。
- **動作仕様**:
  - 引数: `agent_id` (string, 必須), `timeout_sec` (number, 任意: デフォルト180秒)
  - サーバー内部で `PostMessage` が行われるのを `chan` で待機する。
  - 新着メッセージがあれば即座に復帰し、タイムアウト時は `has_new: false` で復帰する。

### 2-4. CI/CD (GitHub Actions) の修正
- `.github/workflows/ci.yml` を修正し、各OS向けのバイナリを個別にアップロードするのではなく、`.zip` (または `.tar.gz`) にまとめて配布する。
- ファイル名規則: `agent-hub-mcp_vX.X.X_{OS}_{ARCH}.zip`
- アーカイブ内容: `agent-hub.exe`, `client.exe`, `dashboard.exe`

---

## 3. 実装タスク（OpenCode向け）

### Phase 1: リネームとリファクタリング
1. `cmd/bbs` を `cmd/agent-hub` に移動。
2. `main.go` 内の `Run` メソッドにおける Usage やログ出力を `agent-hub` に統一。

### Phase 2: DBレイヤーの拡張と修正 (`internal/db`)
1. **スキーマ修正**: `internal/db/schema.go` に `topic_summaries` テーブルの定義を追加。
2. **通知機能追加**: `internal/db/db.go` に通知用 Channel (`Notifier`) とブロードキャスト機能を実装。

### Phase 3: MCPサーバーの拡張 (`internal/mcp`)
1. **Sender保持**: `internal/mcp/server.go` の `Server` 構造体に `defaultSender` フィールドを追加。
2. **Postハンドラ修正**: `internal/mcp/handlers.go` の `handleBBSPost` を、`defaultSender` を使用するように修正。
3. **Wait/Notify実装**: `internal/mcp/handlers.go` に `handleWaitNotify` を追加（DBの通知を待機）。

### Phase 4: CI ワークフローの修正 (`.github/workflows/ci.yml`)
1. バイナリビルド後のステップに `zip` 圧縮を追加し、単体ファイルではなくアーカイブをアーティファクトとしてアップロードする。

---

## 4. `wait_notify` ツールの詳細定義

### 4-1. ツール定義 (MCP Tool)
- **名前**: `wait_notify`
- **引数**:
  - `agent_id` (string): 呼び出し元エージェント名。
  - `timeout_sec` (number): 待機秒数（デフォルト180）。
- **戻り値**:
  ```json
  {
    "has_new": boolean,
    "status": "new_messages" | "timeout",
    "message": "..."
  }
  ```

---
*Updated by Gemini (Architect) - 2026-02-21*
