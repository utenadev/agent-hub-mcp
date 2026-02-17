# SPEC-008: CLI機能のブラッシュアップと診断機能の強化

## 1. 概要
SPEC-007で導入された `doctor` および `setup` コマンドを改善し、設定ディレクトリの自動管理、データベース整合性チェックの厳密化、およびClaude Desktop連携の利便性を向上させる。

## 2. 修正・拡張内容

### 2.1 データベース整合性チェックの厳密化
`internal/db/db.go` の `CheckIntegrity()` を拡張し、単一のテーブルだけでなく、システム稼働に必要な全テーブルの存在を確認する。
- **対象テーブル:** `topics`, `messages`, `topic_summaries`, `agent_presence`

### 2.2 セットアップ機能の強化 (`setup`)
- **設定ディレクトリの作成:** `~/.config/agent-hub-mcp/` ディレクトリが存在しない場合は自動作成する。
- **設定ファイルの雛形作成:** 上記ディレクトリに `config.json` が存在しない場合、空の雛形（Gemini APIキーのプレースホルダ等を含む）を作成する。
- **Claude Desktop設定ヘルパー:**
    - 実行バイナリの絶対パスを自動取得する。
    - 実行環境（OS）を判別し、適切な `claude_desktop_config.json` の保存場所を提示する。
    - そのままコピー＆ペースト可能なJSONスニペットを表示する。

### 2.3 診断機能の強化 (`doctor`)
- **ディレクトリ権限のチェック:** 設定ディレクトリおよびデータベースファイルに対する読み書き権限をチェックする。
- **詳細なエラー表示:** 整合性チェックに失敗した場合、どのテーブルが欠損しているかを具体的に表示する。

## 3. 実装上の注意点

- **OS固有のパス処理:** `os.UserConfigDir()` を使用して、OSごとの標準的な設定ディレクトリ（Linux: `~/.config`, Windows: `AppData/Roaming`, macOS: `Library/Application Support`）を適切に扱う。
- **絶対パスの取得:** `os.Executable()` を使用して、現在実行中のバイナリパスを正確に特定する。

## 4. 実装ステップ
1. `internal/db/db.go`: `CheckIntegrity()` のチェック対象テーブルを拡充。
2. `cmd/agent-hub/main.go`:
    - `runSetup` 内でディレクトリ作成と `config.json` 雛形作成ロジックを追加。
    - `runSetup` の最後に、絶対パスとOS情報を元にしたClaude設定ヘルパーを表示する処理を追加。
3. `doctor` の出力メッセージの微調整。
