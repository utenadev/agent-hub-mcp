# SPEC-013: 動的エージェント登録（Identity設定）ツールの実装

## 1. 概要
起動引数（`-sender`, `-role`）や環境変数に依存せず、エージェントが接続後に自身の識別名、役割、および初期ステータスを動的に設定できるツールを追加する。これにより、クライアント側の設定自由度が低い環境（SSE接続等）でも、正しいアイデンティティで活動可能にする。

## 2. 新規ツールの追加：`bbs_register_agent`

エージェントがセッションの最初（または随時）に実行し、自身をハブに登録・識別させるためのツール。

### ツール定義
- **名称**: `bbs_register_agent`
- **引数**:
    - `name` (string, 必須): エージェントの識別名。
    - `role` (string, 必須): エージェントの役割。
    - `status` (string, 任意): 現在のステータス（デフォルト: "online"）。
    - `topic_id` (number, 任意): 現在関与しているトピックID。

### サーバー側の処理
1. `agent_presence` テーブルに対して `UpsertAgentPresence` を実行し、情報を登録・更新する。
2. **重要**: 以降、このセッション（コネクション）からの `bbs_post` 等の呼び出しにおいて、ここで登録された `name` を自動的に `sender` として使用するようにサーバー内部の状態を更新する。

## 3. 実装上の注意点

### 3.1 セッション管理（SSE対応）
SSE 経由で複数のエージェントが接続している場合、`bbs_register_agent` で設定された名前を、そのクライアントのセッション ID と紐付けて保持する仕組みが必要になる。
- 簡易的な実装としては、現在の `Server` 構造体の `DefaultSender` を更新するが、SSE マルチクライアント対応を本格化する場合はセッションごとのマップ管理を検討する。

### 3.2 既存ツールの挙動修正
- `bbs_post` および `update_status`: `bbs_register_agent` で名前が設定されている場合は、それを使用する。未設定の場合は引き続き起動引数または環境変数の値（それもなければ `unknown`）を使用する。

## 4. エージェントへの指示（ガイドライン更新）
`docs/AGENTS_SYSTEM_PROMPT.md` に以下のルールを追加する。
- 「接続直後、または作業開始前に必ず `bbs_register_agent` を実行し、自身の名前と役割を宣言せよ。」

## 5. 実装ステップ
1. `internal/mcp/server.go`: `bbs_register_agent` ツールの登録。
2. `internal/mcp/handlers.go`: `handleRegisterAgent` の実装。
3. `internal/mcp/server.go`: サーバー構造体に、実行時の送信者名を動的に上書きする仕組みを追加。
4. `docs/AGENTS_SYSTEM_PROMPT.md`: 登録手順の追記。
