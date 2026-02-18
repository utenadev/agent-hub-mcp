# SPEC-011: CLI改善、スキーマ同期、およびSSEドキュメントの強化

## 1. 概要
v0.0.2でのフィードバックに基づき、データベーススキーマの欠落を修正し、CLIの利便性（ヘルプ、デフォルトパス）を向上させ、主要な接続方式であるSSEに関する情報を充実させる。

## 2. 修正・拡張内容

### 2.1 スキーマの同期 (`internal/db/schema.go`)
`doctor`コマンドで指摘されていた `topic_summaries` テーブルの欠落を解消する。
```sql
CREATE TABLE IF NOT EXISTS topic_summaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    topic_id INTEGER NOT NULL,
    summary_text TEXT NOT NULL,
    is_mock BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(topic_id) REFERENCES topics(id)
);
```

### 2.2 デフォルトDBパスの変更
`-db` フラグが指定されない場合のデフォルト値を `agent-hub.db` (カレントディレクトリ) から、`setup` で作成される標準ディレクトリ（例: `~/.config/agent-hub-mcp/agent-hub.db`）に変更する。

### 2.3 ヘルプ機能の強化
- `agent-hub help` サブコマンドの実装。
- 引数なし起動時に、利用可能なサブコマンド一覧とSSE接続のヒントを表示。

### 2.4 SSE接続ガイダンスの表示
`serve -sse` 起動時に、クライアントが接続すべきURLをコンソルに明示する。
- 表示例: `SSE Endpoint: http://localhost:8080/sse`

### 2.5 GEMINI.md の更新
バイナリ名、ディレクトリ名、最新のスキーマ定義を反映する。

## 3. 実装ステップ
1. `internal/db/schema.go`: `topic_summaries` テーブルを追加。
2. `cmd/agent-hub/main.go`:
    - デフォルトDBパス解決ロジックの追加。
    - `runHelp` 関数の実装。
    - `runServe` 起動時のメッセージ強化。
3. `GEMINI.md`: 全体的な記述の最新化。
