# 独立監査レポート: プロジェクト整合性と品質の検証

**監査日**: 2026-02-22  
**監査官**: Amp (独立監査官として参加)  
**対象イシュー**: bd-26w  
**対象バージョン**: v0.0.7相当 (Unreleased含む)

---

## エグゼクティブサマリー

agent-hub-mcp プロジェクトの4項目（DBパス一貫性、スキーマ同期、ドキュメント完備性、テストカバレッジ）を精査した。
**重大バグ2件**、設計上の問題3件、ドキュメント乖離3件を検出。スキーマの同期については問題なし。

| 区分 | 件数 | 重大度 |
|------|------|--------|
| 重大バグ | 2 | 🔴 Critical |
| 設計問題 | 3 | 🟡 Medium |
| ドキュメント乖離 | 3 | 🟡 Medium |
| テスト不足 | 4 | 🟡 Medium |

---

## 1. DBパスの一貫性

### 結論: 不整合あり

#### ISSUE-1: `getDefaultDBPath()` の重複定義 🟡

**場所**: `cmd/agent-hub/main.go:59-65`

`internal/config.DefaultDBPath()` と完全に同一のロジックが `cmd/agent-hub/main.go` にローカル関数として重複定義されている。

```go
// cmd/agent-hub/main.go:59-65 (ローカル版)
func getDefaultDBPath() string {
    configDir, err := os.UserConfigDir()
    if err != nil {
        return "agent-hub.db"
    }
    return filepath.Join(configDir, "agent-hub-mcp", "agent-hub.db")
}
```

**影響範囲**:
- `cmd/agent-hub/serve.go:17` → ローカル版を使用
- `cmd/agent-hub/orchestrator.go:20` → ローカル版を使用
- `cmd/agent-hub/doctor.go:16` → ローカル版を使用
- `cmd/agent-hub/setup.go:18` → ローカル版を使用
- `cmd/agent-hub/help.go:20` → ローカル版を使用
- `cmd/dashboard/main.go:41` → **`config.DefaultDBPath()` を使用** ✅

現時点ではロジックが一致しているため実害はないが、将来片方のみ変更された場合にサイレントな不整合が発生するリスクがある。

**推奨**: ローカル版を削除し、全箇所で `config.DefaultDBPath()` を使用する。

#### ISSUE-2: Orchestrator の config パスがハードコード 🟡

**場所**: `internal/hub/orchestrator.go:49`

```go
configPath := filepath.Join(homeDir, ".config", "agent-hub-mcp", "config.json")
```

`os.UserConfigDir()` ではなく `homeDir + ".config"` をハードコードしている。

**影響**: macOS では `~/Library/Application Support/agent-hub-mcp/config.json` が正しいパスだが、このコードは `~/.config/agent-hub-mcp/config.json` を参照する。Windows でも同様の不整合が発生する。

**推奨**: `config.DefaultConfigPath()` を使用する。

#### ISSUE-3: Dashboard テストの期待値不一致 🟡

**場所**: `cmd/dashboard/main_test.go:44`

```go
if path != "agent-hub.db" {  // ← 旧パスをハードコード
    t.Errorf("expected default path 'agent-hub.db', got: %s", path)
}
```

`dashboard/main.go` は `config.DefaultDBPath()` を使用しており `~/.config/agent-hub-mcp/agent-hub.db` を返すが、テストは旧パス `"agent-hub.db"` を期待している。

**影響**: `go test ./cmd/dashboard/` が常に FAIL する。

---

## 2. スキーマの同期

### 結論: 問題なし ✅

#### スキーマ定義 (`internal/db/schema.go`)

| テーブル | カラム | CRUD対応 |
|----------|--------|----------|
| `topics` | id, title, created_at | `CreateTopic`, `ListTopics` ✅ |
| `messages` | id, topic_id, sender, content, created_at | `PostMessage`, `GetMessages`, `CountUnreadMessages` ✅ |
| `agent_presence` | name, role, status, topic_id, last_seen, last_check | `UpsertAgentPresence`, `UpdateAgentStatus`, `UpdateAgentCheckTime`, `GetAgentPresence`, `ListAllAgentPresence` ✅ |
| `topic_summaries` | id, topic_id, summary_text, is_mock, created_at | `SaveSummary`, `GetLatestSummary`, `GetSummariesByTopic` ✅ |

- `CheckIntegrity()` が4テーブル全てを検証 ✅
- 全 SQL クエリが schema.go のカラム定義と一致 ✅
- `rows.Err()` チェック、`rows.Close()` defer が全箇所で適切 ✅
- `sql.NullInt64` による nullable カラムの処理が適切 ✅

#### ただし、CRUD ロジック内に重大バグあり（後述 BUG-1）

---

## 3. 重大バグ

### BUG-1: `GetLatestSummary` のエラー比較が永久に失敗する 🔴

**場所**: `internal/db/summary.go:44`

```go
func (db *DB) GetLatestSummary(topicID int64) (*TopicSummary, error) {
    // ...
    err := row.Scan(...)
    if err != nil {
        if err == fmt.Errorf("sql: no rows in result set") {  // ← BUG
            return nil, nil
        }
        return nil, fmt.Errorf("failed to get latest summary: %w", err)
    }
    // ...
}
```

**問題**: `fmt.Errorf()` は毎回新しい `error` オブジェクトを生成する。Go の `==` 演算子はポインタ比較となるため、この条件は**絶対に `true` にならない**。

**正しいコード**:
```go
if err == sql.ErrNoRows {
    return nil, nil
}
```

**影響**: サマリーが存在しないトピックに対して `GetLatestSummary` を呼ぶと、`nil, nil` ではなく `nil, error` を返す。これにより下流の増分要約ロジックに影響する（BUG-2 参照）。

### BUG-2: `generateSummary` でエラーが握りつぶされている 🔴

**場所**: `internal/hub/orchestrator.go:273-276`

```go
func (o *Orchestrator) generateSummary(ctx context.Context, topicID int64) error {
    // ...
    latestSummary, err := o.db.GetLatestSummary(topicID)  // 行273: err を宣言

    messages, err := o.db.GetMessages(topicID, 50)         // 行276: err を上書き！
    if err != nil {
        return err  // ← GetLatestSummary のエラーは検査されない
    }
    // ...
}
```

**問題**: `GetLatestSummary` の返り値 `err` が次行の `GetMessages` で上書きされ、チェックされない。

**BUG-1 との複合影響**:
1. BUG-1 により、サマリー未存在時に `GetLatestSummary` は常にエラーを返す
2. BUG-2 により、そのエラーは無視される
3. `latestSummary` は常に `nil` になる
4. 結果として、増分要約 (`llmIncrementalSummarizer`) が**一度も使われず**、常にフルスキャン要約 (`llmSummarizer`) が実行される

**実害**: 増分要約機能が完全に死んでいる。Gemini API のトークン消費が本来より多い可能性がある。

---

## 4. ドキュメントの乖離

### DOC-1: `internal/mcp/AGENTS.md` が大幅に古い 🟡

**現状のドキュメント**: ツール3件のみ記載
```
bbs_create_topic, bbs_post, bbs_read
```

**実装 (`server.go`)**: ツール7件
```
bbs_create_topic, bbs_post, bbs_read,
check_hub_status, update_status, bbs_register_agent, wait_notify
```

また、Server 構造体の記述が古い:
- ドキュメント: `Server { mcpServer, db }`
- 実装: `Server { mcpServer, db, DefaultSender, DefaultRole, CurrentSender, notifier }`
- "Sender defaults to 'unknown' (TODO: use BBS_AGENT_ID)" は既に解決済み

### DOC-2: `internal/hub/AGENTS.md` の API キー優先順位が不完全 🟡

**ドキュメント** (3段階):
1. `~/.config/agent-hub-mcp/config.json`
2. `HUB_MASTER_API_KEY` env var
3. `GEMINI_API_KEY` env var

**実装** (`orchestrator.go:45-76`, 4段階):
1. `~/.config/agent-hub-mcp/config.json`
2. **`Config.APIKey` (explicitly set)** ← 欠落
3. `HUB_MASTER_API_KEY` env var
4. `GEMINI_API_KEY` env var

### DOC-3: `docs/AGENT_HUB_USAGE.md` が不完全 🟡

4/7 ツールのみ言及。以下が欠落:
- `bbs_create_topic` の独立セクション（ステップ1で `bbs_register_agent` のみ言及）
- `wait_notify` ツール（全く言及なし）

---

## 5. テストカバレッジ

### 現状

```
cmd/agent-hub    : 61.0%  ⚠️
cmd/client       : 79.2%  ✅
cmd/dashboard    : 70.6%  ❌ (1 FAIL)
internal/config  :  0.0%  ❌ (テストファイルなし)
internal/db      : 22.6%  ❌
internal/hub     : 38.8%  ⚠️
internal/mcp     : 42.6%  ⚠️
internal/ui      : 32.5%  ⚠️
```

### テスト不足の詳細

#### `internal/config` (0.0%) ❌
テストファイルが存在しない。以下が未テスト:
- `DefaultDBPath()`, `DefaultConfigPath()`, `DefaultConfigDir()`
- `New()`, `Load()`, `LoadFromFile()`, `SaveToFile()`
- `GetSender()`, `GetRole()`

#### `internal/db` (22.6%) ❌
テスト済み: `Open`, `CreateTopic`, `PostMessage`, `GetMessages`, `ListTopics`

未テスト:
- **summary.go**: `SaveSummary`, `GetLatestSummary`, `GetSummariesByTopic`
- **presence.go**: `UpsertAgentPresence`, `UpdateAgentStatus`, `UpdateAgentCheckTime`, `GetAgentPresence`, `ListAllAgentPresence`
- **message.go**: `CountUnreadMessages`
- **db.go**: `CheckIntegrity`
- **notifier.go**: `Notifier` 全体 (`Register`, `Unregister`, `Notify`, `NotifyAll`, `Wait`, `Count`)

#### `cmd/dashboard` (FAIL) ❌
`TestDashboardApp_Run_DefaultDBPath` が `config.DefaultDBPath()` の戻り値変更に追従していない。

---

## 6. その他の所見

### 軽微な問題

- `watch_bbs.sh` がハードコード `agent-hub.db`（カレントディレクトリ）を参照しているが、これはスクリプト用途として妥当と判断。
- `AGENTS.md`（プロジェクトルート）の `MCP TOOLS` セクションにも `wait_notify` が記載されていない。

### 良い点

- `rows.Err()` / `rows.Close()` の処理は全 CRUD 関数で適切 ✅
- エラーラッピング (`fmt.Errorf("...: %w", err)`) が一貫 ✅
- WAL モード有効化が `Open()` で確実に実行 ✅
- CGO-free SQLite (`modernc.org/sqlite`) の選定は適切 ✅
- Notifier の並行性制御 (`sync.RWMutex`) は適切 ✅

---

## 推奨アクション（優先順）

| # | 優先度 | 内容 | 対象ファイル |
|---|--------|------|-------------|
| 1 | 🔴 P0 | `GetLatestSummary` の `fmt.Errorf` → `sql.ErrNoRows` 修正 | `internal/db/summary.go:44` |
| 2 | 🔴 P0 | `generateSummary` のエラー握りつぶし修正 | `internal/hub/orchestrator.go:273-276` |
| 3 | 🟡 P1 | `getDefaultDBPath()` 重複削除、`config.DefaultDBPath()` に統一 | `cmd/agent-hub/main.go` + 各サブコマンド |
| 4 | 🟡 P1 | Orchestrator の config パスを `config.DefaultConfigPath()` に変更 | `internal/hub/orchestrator.go:49` |
| 5 | 🟡 P1 | Dashboard テスト修正 | `cmd/dashboard/main_test.go:44` |
| 6 | 🟡 P2 | `internal/config` のテスト追加 | 新規: `internal/config/config_test.go` |
| 7 | 🟡 P2 | `internal/db` の presence/summary/notifier テスト追加 | `internal/db/db_test.go` 拡張 |
| 8 | 🟡 P2 | MCP AGENTS.md をツール7件に更新 | `internal/mcp/AGENTS.md` |
| 9 | 🟡 P3 | `AGENT_HUB_USAGE.md` に全ツール記載 | `docs/AGENT_HUB_USAGE.md` |

---

*以上、独立監査官による報告を終了する。*
