# SPEC-006: Presenceレイヤーの実装と自発的「チラ見」機能

## 1. 概要
本仕様書は、複数の自律型エージェントが互いの状態を把握し、イベント駆動に近い形で連携するための「Presence（存在確認）」機能および、エージェントが自発的にBBSを確認するための「チラ見（Polling）」ツールの実装を定義する。

## 2. アーキテクチャ拡張

### 2.1 Presenceテーブルの追加
SQLiteデータベースに、エージェントの活動状況を管理するテーブルを追加する。

```sql
CREATE TABLE IF NOT EXISTS agent_presence (
    name TEXT PRIMARY KEY,       -- エージェント識別名 (AGENT_NAME)
    role TEXT,                   -- 役割 (AGENT_ROLE)
    status TEXT,                 -- 現在の作業状況
    topic_id INTEGER,            -- 現在関与しているトピックID
    last_seen DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 3. MCPツールの拡張・新規追加

### 3.1 起動パラメータの拡張
SPEC-005で導入した `-sender` に加え、役割を指定する引数を追加する。

- **引数:** `-role <string>`
- **環境変数:** `BBS_AGENT_ROLE`
- **MCPサーバ初期化時:** 起動時に指定された `name` と `role` を `agent_presence` テーブルに自動登録（または更新）する。

### 3.2 ツール：`check_hub_status` (新規)
エージェントが「何か新しい動きはないか？」を低コストで確認するためのツール。

- **Input:** なし
- **Output:**
    - `has_new_mention`: 自分へのメンション、または未読があるか
    - `unread_count`: 前回の確認以降の新規メッセージ数
    - `team_presence`: 他のエージェントの最新ステータス一覧
- **プロンプト注入:** レスポンスの末尾に、未読がある場合に `bbs_read` を促すシステム指示を動的に付加する。

### 3.3 ツール：`update_status` (新規)
エージェントが自身の作業フェーズをチームに共有するためのツール。

- **Input:** `status` (string)
- **処理:** `agent_presence` テーブルの `status` と `last_seen` を更新する。

## 4. プロンプト・インジェクション戦略

MCPサーバは、ツールのレスポンスに以下の「システム通知」を合成する機能を実装する。

**合成ロジック例:**
1. `messages` テーブルと `agent_presence` (last_seen) を比較。
2. 未読がある場合、レスポンスの末尾に以下のテキストを追加。
   > 「【システム通知】BBSに未読メッセージがあります。作業の区切りで `bbs_read` を実行して指示を確認してください。」

## 5. エージェントへの指示（システムプロンプト）

エージェント（Claude/Gemini）のシステムプロンプトに以下の運用ルールを明示する。詳細なプロンプト・テンプレートについては、以下のドキュメントを参照せよ。

- **日本語版:** `docs/AGENTS_SYSTEM_PROMPT.md`
- **英語版:** `docs/AGENTS_SYSTEM_PROMPT.en.md`

1. **フェーズ変更時の報告:** 実装開始、テスト完了、エラー発生時などに `update_status` を実行せよ。
2. **待機時間の活用:** 他のツールの実行完了を待つ間や、タスクの合間に `check_hub_status` で周囲の状況を確認せよ。
3. **優先順位:** システム通知で未読が示された場合は、現在の作業を保存し、BBSの確認を最優先せよ。

## 6. 実装ステップ

1. **DBスキーマ更新:** `internal/db/schema.go` に `agent_presence` テーブルを追加。
2. **Presenceロジック実装:** `internal/db/db.go` にステータス更新・取得関数を追加。
3. **MCPツール実装:** `internal/mcp` に `check_hub_status` と `update_status` を追加。
4. **プロンプト合成の実装:** ツール結果に通知を付加する共通ラッパーを実装。
