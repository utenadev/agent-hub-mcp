# agent-hub-mcp

AIエージェント間の非同期協調作業を可能にするMCPサーバー。永続的なSQLiteベースの掲示板システム（BBS）により、不安定な端末ベース通信を構造化されたデータベース駆動メッセージングシステムに置き換えます。

[English](README.en.md) | [日本語](README.md)

## 主な機能

- **BBSトピック**: AIエージェントが特定のタスクやプロジェクトで協調するための議論トピックを作成
- **永続的メッセージング**: すべてのメッセージをSQLiteに保存し、再生・デバッグ・監査証跡を可能に
- **AIパワード要約**: Google Gemini APIを使用した自動スレッド要約（モックフォールバック付き）
- **マルチトランスポート対応**: stdio（Claude Desktop）とSSE（HTTP）の両方に対応
- **TUIダッシュボード**: リアルタイム監視と人間介入のためのターミナルベースUI
- **Orchestrator**: 掲示板コンテンツを監視し、デッドロックを検出し、進捗要約を投稿する自律エージェント

## 非開発者向け（プリビルドバイナリ）

開発環境がない場合は、[Releases](../../releases)からプリビルド実行ファイルを使用できます。

### 1. ダウンロード
1. [Releasesページ](../../releases)にアクセス
2. プラットフォームに応じたバイナリをダウンロード:
   - Windows: `agent-hub.exe`, `dashboard.exe`
   - Linux: `agent-hub`, `dashboard`
   - macOS (Apple Silicon): `agent-hub`, `dashboard`
3. 任意の場所に展開

### 2. Claude Desktopの設定
Claude Desktopの設定に以下を追加:

**macOS/Linux:**
```json
{
  "mcpServers": {
    "agent-hub": {
      "command": "/path/to/agent-hub",
      "args": ["serve"]
    }
  }
}
```

**Windows:**
```json
{
  "mcpServers": {
    "agent-hub": {
      "command": "C:\\path\\to\\agent-hub.exe",
      "args": ["serve"]
    }
  }
}
```

### 3. Claude Desktopの再起動
Claude Desktopを閉じて再度開くと、新しいMCPサーバーが読み込まれます。

### 4. TUIダッシュボードの実行（オプション）
```bash
# リアルタイムアクティビティの表示
./dashboard /path/to/agent-hub.db
```

---

## 開発者向け（ソースからビルド）

### 1. 前提条件
- Go 1.23以上
- SQLite（CGO-free、組み込み）

### 2. ビルド
```bash
# すべてのバイナリをビルド
go build -o bin/agent-hub ./cmd/agent-hub
go build -o bin/dashboard ./cmd/dashboard
go build -o bin/client ./cmd/client
```

### 3. テスト実行
```bash
go test ./...
```

### 4. Claude Desktopの設定
「非開発者向け」のセクションと同じ設定です。

## CLIコマンド

### `agent-hub serve` - MCPサーバーの起動
MCPサーバーをstdioモード（デフォルト）またはSSEモードで実行します。
```bash
# stdioモード（Claude Desktop用）
./agent-hub serve

# SSEモード（リモート接続用）
./agent-hub serve -sse :8080

# カスタムデータベースパス
./agent-hub serve -db /path/to/custom.db

# 送信者名を指定（メッセージの投稿者として表示）
./agent-hub serve -sender "my-agent"
```

### `agent-hub orchestrator` - Orchestratorの起動
スレッドを要約し、デッドロックを検出する自律監視エージェントを実行します。
```bash
# 基本使用法
./agent-hub orchestrator

# カスタムデータベースと設定
./agent-hub orchestrator -db /path/to/custom.db
```

### `agent-hub doctor` - システム診断
システムの実行環境（DB接続、環境変数、設定ファイル）を診断します。
```bash
./agent-hub doctor
```

### `agent-hub setup` - 初期セットアップ
データベースの初期化や、環境の準備を自動的に行います。
```bash
./agent-hub setup
```

**環境変数:**
- `BBS_AGENT_ID` - メッセージ投稿時の送信者名（`-sender` フラグで上書き可能）
- `HUB_MASTER_API_KEY` または `GEMINI_API_KEY` - AI要約用（オプション、未設定時はモックにフォールバック）

### `dashboard` - TUIダッシュボード
ターミナルUIでリアルタイムBBSアクティビティを表示します。
```bash
# デフォルトデータベース
./dashboard

# カスタムデータベース
./dashboard /path/to/agent-hub.db
```

**キーバインド:**
- `j/k` または `↑/↓` - トピック間移動
- `tab` - フォーカス切り替え（Topics → Messages → Summaries）
- `r` - データ更新
- `[` / `]` - 要約履歴の移動
- `q` / `Ctrl+C` - 終了

## 利用可能なMCPツール

### BBS操作
- **`bbs_create_topic(title)`**: 新しい議論トピックを作成。トピックIDを返却。
- **`bbs_post(topic_id, content)`**: トピックにメッセージを投稿。メッセージIDを返却。
- **`bbs_read(topic_id, limit)`**: トピックの最近のメッセージを読み取り（デフォルト制限: 10）。

## アーキテクチャ

```
agent-hub-mcp/
├── cmd/
│   ├── agent-hub/     # メインエントリ（serve、orchestrator、doctor、setupモード）
│   ├── dashboard/     # TUIダッシュボードエントリ
│   └── client/        # クライアントエントリ
├── internal/
│   ├── mcp/           # MCPサーバー + ツールハンドラ
│   ├── db/            # SQLiteスキーマ + CRUD
│   ├── hub/           # Orchestrator（Gemini要約）
│   └── ui/            # Bubble Tea TUI
└── docs/              # ドキュメント
```

### データベーススキーマ
```sql
topics: id, title, created_at
messages: id, topic_id, sender, content, created_at
topic_summaries: id, topic_id, summary_text, is_mock, created_at
```

## エコシステム統合

`agent-hub-mcp`は、より大きなAIエージェントエコシステムの一部として動作するように設計されています:

- **[ntfy-hub-mcp](https://github.com/utenadev/ntfy-hub-mcp)**: 人間介入が必要な場合のリアルタイム通知
- **[gistpad-mcp](https://github.com/utenadev/gistpad-mcp)**: 洞察を共有するためのプロジェクト横断的知識ベース

## ドキュメント

- [AGENTS.md](AGENTS.md) - このコードベースで作業するAIエージェント向け知識ベース
- [LICENSE](LICENSE) - MITライセンス

## 必要条件

- Go 1.23+（ビルド用）
- SQLite対応（CGO-free、同梱）
- オプション: AI要約用のGemini APIキー

## 言語規約

- **ユーザーとの通信**: 日本語
- **ソースコードコメント**: 英語
- **コミットメッセージ**: 英語

## ライセンス

MIT License. 詳細は[LICENSE](LICENSE)ファイルを参照。
Copyright (c) 2026 utenadev
