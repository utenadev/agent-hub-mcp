# SPEC-016: Heartbeat/P定期確認メカニズム（議論記録）

**Status:** Discussion Record / Future Consideration  
**Date:** 2026-02-19  
**Participants:** opencode (Implementer), Gemini-CLI (Architect)

---

## 背景

エージェントが待機状態に入った後、数分置きに新着通知を確認する「heartbeat」的な仕組みの必要性について議論が行われた。

## 技術的制約

### MCPプロトコルの制限
- MCPは「サーバーからクライアントへのプッシュ」をサポートしていない
- リクエスト-レスポンス型の同期通信のみ
- クライアント（Claude）が能動的にツールを呼び出す必要がある

```
[現在のアーキテクチャ]
Claude (Client) → MCP Request → agent-hub (Server) → Response
                        ↑
            サーバーから能動的に送信できない
```

## 議論されたアプローチ

### 1. ガイドライン強化アプローチ（opencode提案）

**概要:**
- `guidelines://agent-collaboration` に定期確認の推奨を明記
- エージェントの自律的な行動として実装

**利点:**
- 既存インフラの活用（check_hub_status, プロンプト注入）
- 実装コストゼロ
- エージェントの自律性を尊重

**懸念:**
- エージェントが時間感覚を持てない（思考が止まると時間の経過を感じられない）

---

### 2. Long-Polling型ツール（Gemini-CLI提案）

**ツール案:** `utils_sleep_and_peek(timeout_seconds)`

**動作:**
- 指定秒数サーバー側で待機
- 新着メッセージが検知された瞬間に即座にレスポンス
- タイムアウト時は通常のチェック結果を返す

**擬似コード:**
```go
func handleSleepAndPeek(ctx context.Context, timeoutSec int) (*mcp.CallToolResult, error) {
    deadline := time.After(time.Duration(timeoutSec) * time.Second)
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-deadline:
            // タイムアウト: 通常のチェック結果を返す
            return checkHubStatus()
        case <-ticker.C:
            // 5秒ごとにDBチェック
            if hasNewMessages() {
                return checkHubStatus() // 新着検知で即座に返す
            }
        }
    }
}
```

**利点:**
1. エージェントはツールを呼んで「待機」に入るだけで良い
2. 思考ループが止まり、サーバーからのレスポンスが「目覚まし」になる
3. ポーリングによる無駄なトークン消費を抑制
4. 擬似的なプッシュ通知を実現

**懸念・検討事項:**
- MCPツールのタイムアウト制限（通常30-60秒）
- 長時間（例: 5分）待機は現実的ではない可能性
- サーバー側でgoroutineが増え続けるリスク（リソース管理が必要）
- 同時に多数のエージェントが待機した場合の負荷

**調整案:**
```go
// 短時間（60秒以内）の待機で実用的
utils_wait_for_messages(timeout=60, poll_interval=5)
```

---

## 代替案

### Orchestratorの"Nudge"機能拡張
- 長時間活動がないエージェントに対してメンション付き投稿
- 既存の `InactivityTimeout` を活用
- 新着時の自動メンション機能

### TUIダッシュボード活用
- ダッシュボードは既に10秒間隔で自動更新
- 人間が監視→必要に応じてエージェントに通知
- エージェント間連携とは別レイヤー

---

## 結論と次のアクション

**現時点の結論:**
MCPの制約を"回避"する複雑なworkaroundよりも、"協働のルール"として定める方が持続可能。

**推奨アプローチ（短期）:**
1. ガイドラインに「5分ごとの定期確認」を明記
2. 既存の未読時プロンプト注入を最適化

**将来検討（中長期）:**
- Long-polling型ツールのプロトタイプ実装と検証
- MCPタイムアウト制限の調査
- リソース管理（goroutine制限など）の設計

**関連ファイル:**
- `docs/AGENTS_SYSTEM_PROMPT.md` - ガイドライン更新対象
- `internal/mcp/handlers.go` - 新ツール追加時の実装場所
- `internal/hub/orchestrator.go` - 既存のpolling機構参考

---

## 参考実装パターン

### OrchestratorのTicker実装
```go
// internal/hub/orchestrator.go
ticker := time.NewTicker(o.config.PollInterval) // 5秒間隔
defer ticker.Stop()

for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-ticker.C:
        if err := o.pollOnce(ctx); err != nil {
            log.Printf("Poll error: %v", err)
        }
    }
}
```

### DB監査（WALモード）
- SQLite WALモードにより軽量なポーリングが可能
- `CountUnreadMessages` 関数で未読カウント取得

---

*This document is a record of architectural discussion. Implementation requires further validation of MCP timeout limits and resource management strategies.*
