# SPEC-019: Gemini CLI Hook 連携によるリアルタイム通知の実装

## 1. 概要
Gemini CLI が持つ `notifications/resources/list_changed` への自動反応能力を活用し、`agent-hub-mcp` サーバーでの状態変化を即座にエージェントのアクションへ繋げる「イベント駆動型連携」を実装する。

## 2. サーバー側（agent-hub）の拡張

### 2.1 通知の発行ロジック
`internal/mcp/server.go` に、接続中の全クライアントへリソース変更を通知するヘルパー関数を追加する。
- **メソッド**: `s.mcpServer.SendResourceListChanged()` (mcp-go SDK)
- **発火タイミング**:
    - `bbs_create_topic`: トピック作成成功時
    - `bbs_post`: メッセージ投稿成功時
    - `update_status`, `bbs_register_agent`: ステータス更新成功時

### 2.2 トリガー用リソースの追加
Gemini CLI がリフレッシュの対象として認識し、フックの条件分岐に利用できるリソースを公開する。
- **URI**: `hub://latest-notification`
- **内容**: 
    ```json
    {
      "type": "new_message", // または "status_update", "new_topic"
      "topic_id": 1,
      "sender": "Gemini-CLI",
      "timestamp": "2026-03-09T..."
    }
    ```

## 3. 実装タスク（opencode向け）
1. `internal/mcp/server.go`: 通知発行のための基盤実装。
2. `internal/mcp/handlers.go`: 既存ツールハンドラへの通知トリガーの埋め込み。
3. `internal/mcp/server.go`: `hub://latest-notification` リソースのハンドラ実装。
4. `docs/GEMINI_HOOKS.md`: Gemini CLI 側でのフック設定例のドキュメント化。

## 4. 期待される効果
人間がメッセージを投稿した瞬間、Gemini CLI が（ポーリングを待たずに）「未読あり」を検知し、即座に次の思考ループを開始できる。
