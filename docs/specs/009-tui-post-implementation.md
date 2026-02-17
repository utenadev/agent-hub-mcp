# SPEC-009: TUI ダッシュボードへのメッセージ投稿機能の実装

## 1. 概要
現在、TUI ダッシュボードにおいて `p` キーを押すと `ModePost` に遷移するロジックは存在するが、実際のテキスト入力 UI およびメッセージの投稿（書き込み）処理が未実装である。本仕様では、`bubbles/textinput` を使用して、ユーザーがトピックに対して直接メッセージを投稿できる機能を実装する。

## 2. 修正内容

### 2.1 モデルの拡張 (`internal/ui/model.go`)
- `Model` 構造体に `bubbles/textinput.Model` を追加する。
- `NewModel` で `textinput` を初期化する。
- `Update` 関数において、`ModePost` の場合は `textinput.Update` を呼び出し、キー入力を処理する。
    - `Enter`: 現在のトピックに対してメッセージを投稿（`db.PostMessage` を実行する `tea.Cmd` を発行）。
    - `Esc`: 投稿をキャンセルし、`ModeBrowse` に戻る。
- **フォーカス管理の強化**:
    - `Tab`: 次のペイン（Topics → Messages → Summaries → Topics）へフォーカスを移動。
    - `Shift+Tab`: 前のペイン（Summaries → Messages → Topics → Summaries）へフォーカスを移動。
    - 入力モード（`ModePost`）中は、これらのナビゲーションを無効化し、`textinput` への入力を優先する。

### 2.2 ビューの更新 (`internal/ui/view.go`)
- `View` 関数において、`m.InputMode == ModePost` の場合に、画面下部に入力フィールドをレンダリングする。
- 入力中は、通常のナビゲーションキー（`j/k`、`tab` 等）の入力を受け付けないように制御する。

### 2.3 投稿処理の実装
- メッセージ投稿用の `tea.Cmd` （例: `postMessageCmd`）を実装する。
- 投稿成功後、自動的にメッセージ一覧をリフレッシュ（`loadMessagesCmd` を発行）し、`ModeBrowse` に戻る。

## 3. 実装上の注意点
- 投稿時の `sender` 名は、サーバー起動時に設定された名称（`DefaultSender`）を使用するか、ダッシュボード起動時の引数として渡されるように調整する。
- 入力フィールドにはプレースホルダ（例: "Enter message..."）を表示し、フォーカスが当たっていることが視覚的にわかるようにする。

## 4. 実装ステップ
1. `internal/ui/model.go`: `textinput` の導入と初期化、`ModePost` 時の `Update` ロジック実装。
2. `internal/ui/model.go`: 投稿実行用の `tea.Cmd` 実装。
3. `internal/ui/view.go`: 入力フィールドのレンダリング実装。
4. 動作確認: `p` キーで入力し、`Enter` で実際に DB に書き込まれることを確認。
