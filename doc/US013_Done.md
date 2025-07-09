# US-013: 状態別フィルタリング 完了報告

## 完了定義

- [x] TODO, IN_PROGRESS, DONE, PENDING で絞り込める
- [x] 複数の状態を同時に指定できる
- [x] フィルタ条件が表示される

## 完了のエビデンス

### 1. フィルタリングロジックの実装

`internal/app/app.go`に`GetFilteredTasksByStatus`メソッドを実装し、指定されたステータスのリストに基づいてタスクをフィルタリングする機能を追加しました。

### 2. 状態選択UIの実装

`cmd/main.go`の`model`構造体に`filterStatusInput`と`filteredStatuses`フィールドを追加し、`initialModel`で初期化しました。
`Update`メソッドに`f`キーでのフィルタリングモードへの遷移と、フィルタリングモードでの入力処理（状態の選択、確定、キャンセル）を追加しました。

### 3. フィルタ条件表示機能とフィルタリング結果の表示

`cmd/main.go`の`View`メソッドを更新し、フィルタリングモードのUIと、フィルタリングが適用されている場合に現在のフィルタ条件を表示するようにしました。
また、`m.tasks`がフィルタリング結果で更新されるようにしました。

### 4. 単体テストの作成

`internal/app/app_test.go`に`TestGetFilteredTasksByStatus`テスト関数を追加し、様々な状態のタスクを作成し、異なるフィルタ条件で`GetFilteredTasksByStatus`を呼び出し、期待される結果が返されることを確認しました。
また、テスト環境でのダミーデータロードを制御するために、`internal/app/app.go`と`internal/app/app_test.go`を修正しました。

全てのテストがパスしたことを確認済みです。