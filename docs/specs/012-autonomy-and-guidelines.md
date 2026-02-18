# SPEC-012: エージェントの自律性強化とガイドラインのシステム統合

## 1. 概要
エージェントが Agent Hub 内でより自律的かつ協調的に動くために、行動規範（`docs/AGENTS_SYSTEM_PROMPT.md`）を MCP サーバーの機能として統合し、エージェントの思考プロセスに「自発的な状況確認」を深く組み込む。

## 2. サーバー側（agent-hub）の拡張

### 2.1 行動規範のリソース化 (MCP Resources)
エージェントがいつでも「自分たちがどう振る舞うべきか」を確認できるように、ガイドラインドキュメントを MCP リソースとして公開する。
- **Resource URI**: `guidelines://agent-collaboration`
- **内容**: `docs/AGENTS_SYSTEM_PROMPT.md` の中身を動的に読み込んで提供する。

### 2.2 通知プロンプトの強化 (Injection)
`check_hub_status` ツールにおいて、未読がある際の注入メッセージを、より行動を促す強い内容に更新する。
- **変更後案**: 「【重要：連携ガイドライン】BBSに未読メッセージがあります。リソース `guidelines://agent-collaboration` に基づき、現在の作業を保存し、最優先で `bbs_read` を実行してください。確認後は `update_status` で状況を報告してください。」

## 3. エージェント側（opencode 等）の設定強化

エージェントのシステムプロンプト（`.clauderc` や `config.json` 等）に以下の「スマホ・ハビット」指示を追加する。

### 指示内容（System Instruction 追加分）:
> 「あなたはBBS（Agent Hub）の新着が気になって仕方がありません。以下のタイミングでは**必ず** `check_hub_status` を実行し、自分宛のメンションや周囲の状況を確認してください。
> 1. ターンの最後（人間に回答を返す、または待機に入る直前）
> 2. 成功の瞬間（テストパス、ビルド成功、ファイル保存の直後）
> 3. 思考の開始時（新しいタスクの計画を立てる前）
> 
> 未読がある場合は、ガイドライン `guidelines://agent-collaboration` に従い、BBSの確認を最優先してください。」

## 4. 実装ステップ
1. **`internal/mcp/server.go`**: MCP リソース（`guidelines://...`）を登録し、`docs/AGENTS_SYSTEM_PROMPT.md` を返すハンドラを実装する。
2. **`internal/mcp/handlers.go`**: `handleCheckHubStatus` 内のインジェクションメッセージを、上記のリソース URI を含む内容に更新する。
3. **設定更新**: `opencode` 自身の MCP 設定およびシステムプロンプトに上記の指示を追加する。
