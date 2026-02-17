# SPEC-007: 診断 (doctor) およびセットアップ (setup) コマンドの実装

## 1. 概要
本仕様書は、`agent-hub` バイナリにシステムの整合性確認を行う `doctor` コマンドと、初期設定を自動化する `setup` コマンドを追加し、ユーザーエクスペリエンス（特にトラブルシューティングと導入の容易さ）を向上させるための変更を定義する。

## 2. サブコマンドの定義

### 2.1 `doctor` コマンド
システムの実行環境を診断し、問題があれば報告・修正案を提示する。

- **チェック項目:**
  - データベース接続: 指定された SQLite パスへの読み書き権限、WALモードの有効性。
  - 環境変数: `GEMINI_API_KEY` (または `HUB_MASTER_API_KEY`) の有無と形式チェック。
  - 依存ツール: `goimports`, `golangci-lint` がインストールされているか（開発環境の場合）。
  - ディレクトリ構造: 設定ファイルディレクトリ (`~/.config/agent-hub-mcp/`) の存在確認。
- **出力形式:** チェック項目ごとに `[OK]`, `[WARN]`, `[ERROR]` を表示。

### 2.2 `setup` コマンド
対話形式（またはフラグ指定）で初期設定を行い、必要なディレクトリや設定ファイルを作成する。

- **機能:**
  - `agent-hub.db` の初期化（テーブル作成）。
  - 設定ファイル (`config.json`) の雛形作成。
  - Claude Desktop 設定ファイル (`config.json`) への MCP サーバー登録補助（パスの自動解決）。
- **引数:**
  - `-force`: 既存の設定を上書きして初期化する。

## 3. 既存機能との調整

### 3.1 `serve` / `orchestrator` との統合
- `serve` 起動時に、暗黙的に `doctor` の一部（DB接続確認等）を実行し、致命的なエラーがあれば即座に終了するように修正。
- 共通のフラグ解析ロジックを整理し、`-db`, `-sender`, `-role` などの共通パラメータを一貫して扱えるようにする。

### 3.2 バイナリ名の統一
- SPEC-005 で変更した `agent-hub` 名称を `doctor` の出力や `setup` のガイダンス内でも徹底する。

## 4. 実装詳細

- **`cmd/agent-hub/main.go`**:
    - `switch command` に `doctor` と `setup` を追加。
    - 各サブコマンドのロジックを別関数（`runDoctor`, `runSetup`）に切り出す。
- **`internal/db/db.go`**:
    - DBの整合性をチェックするための `CheckIntegrity()` メソッドを追加。

## 5. ドキュメントの更新

以下のドキュメントの「CLI Commands」セクションに、新コマンドの説明を追記する。

- **`README.md` (日本語):**
  - `agent-hub doctor`: システムの整合性チェック（DB、環境変数、設定ファイル）。
  - `agent-hub setup`: 初期セットアップの自動実行（DB初期化、Claude Desktop設定補助）。
- **`README.en.md` (英語):**
  - `agent-hub doctor`: System integrity check (DB, Env vars, Config files).
  - `agent-hub setup`: Automated initial setup (DB initialization, Claude Desktop config helper).

## 6. 実装ステップ
1. `doctor` コマンドの実装（現状の確認）。
2. `setup` コマンドの実装（初期設定の自動化）。
3. `serve` 起動時のセルフチェックの強化。
4. `README.md` および `README.en.md` へのコマンド説明の追加。
