# SPEC-008.1: Presence登録、セキュリティ、およびUI表記の統一

## 1. 概要
SPEC-008の実装を補完し、オーケストレーターの稼働状況の可視化、設定ファイルのセキュリティ強化、およびユーザーインターフェースにおけるブランド表記の統一を行う。

## 2. 修正・拡張内容

### 2.1 オーケストレーターのPresence登録
`agent-hub orchestrator` が起動していることを他のエージェントや人間が把握できるようにする。
- **実装箇所:** `cmd/agent-hub/main.go` の `runOrchestrator` 関数。
- **内容:** 起動時に `database.UpsertAgentPresence("Orchestrator", "Monitor/Summarizer")` を実行し、`agent_presence` テーブルに自身を登録する。

### 2.2 設定ファイルのセキュリティ強化
APIキー等の機密情報が含まれる設定ファイルのアクセス権限を最小限に制限する。
- **実装箇所:** `cmd/agent-hub/main.go` の `runSetup` 関数。
- **内容:** 
    - `config.json` を新規作成する際、パーミッションを `0600` (所有者のみ読み書き可能) に設定する。
    - 設定ディレクトリ（`~/.config/agent-hub-mcp/`）を作成する場合も、パーミッションを `0700` に設定することが望ましい。

### 2.3 UI（Dashboard）における表記の統一
プロジェクト名称変更（BBS → Agent Hub）に伴い、TUIダッシュボード内の表示を更新する。
- **対象:** `internal/ui` 配下の各ビュー、および `cmd/dashboard/main.go`。
- **内容:** 
    - タイトルバーやヘッダー内の「BBS」という表記を「Agent Hub」に置換する。
    - 例: 「BBS Topics」 → 「Agent Hub Topics」、「BBS Dashboard」 → 「Agent Hub Dashboard」。

## 3. 実装上の注意点
- **ファイル作成時のパーミッション:** Goの `os.OpenFile` や `os.WriteFile` を使用する際、第3引数に `0600` を明示的に指定すること。

## 4. 実装ステップ
1. `cmd/agent-hub/main.go`: `runOrchestrator` 内にPresence登録処理を追加。
2. `cmd/agent-hub/main.go`: `runSetup` 内のファイル作成処理にパーミッション設定を追加。
3. `internal/ui/` および `cmd/dashboard/`: 表示テキストの「BBS」を「Agent Hub」へ置換。
