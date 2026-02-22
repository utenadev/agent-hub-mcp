# SPEC-018: 監査レポートに基づく不具合修正と整合性の確保

## 1. 概要
独立監査レポート（`docs/report/20260222_audit_integrity.md`）で指摘された重大バグおよび設計上の不整合を修正し、システムの信頼性と一貫性を回復させる。

## 2. 修正内容（優先度順）

### 2.1 【P0】増分要約ロジックの修正
- **`internal/db/summary.go`**: `GetLatestSummary` 内のエラー判定を `if err == sql.ErrNoRows` に修正（`fmt.Errorf` との比較を削除）。
- **`internal/hub/orchestrator.go`**: `generateSummary` 内で `GetLatestSummary` のエラーを正しくチェックし、`err` 変数の上書きを避ける。

### 2.2 【P1】パス解決ロジックの一元化
- **`cmd/agent-hub/main.go`**: ローカル関数 `getDefaultDBPath()` を削除し、すべて **`internal/config.DefaultDBPath()`** に置き換える。
- **`internal/hub/orchestrator.go`**: 設定ファイルの読み込みパスを **`config.DefaultConfigPath()`** に変更（`.config` のハードコードを削除）。

### 2.3 【P1】テストの整合性修正
- **`cmd/dashboard/main_test.go`**: 期待するデフォルトパスを `config.DefaultDBPath()` の結果と一致するように修正。

## 3. 実装上のルール
- **作業ブランチ**: `fix/audit-remediation`
- **成果物の検証**: 修正後、`go test ./...` を実行し、少なくとも `cmd/dashboard` のテストがパスすることを確認すること。
- **コミット方針**: リモート保護ルールに従い、ローカルでの `commit` まで行い、BBSで報告すること。

## 4. 実装ステップ
1. P0バグ（Summary/Orchestrator）の修正。
2. P1不整合（パス解決/テスト期待値）の修正。
3. 全パッケージのテスト実行。
