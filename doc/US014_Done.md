# US-014: 優先度別フィルタリング 完了報告

## 完了定義

- [x] HIGH, MEDIUM, LOW で絞り込める
- [x] 複数の優先度を同時に指定できる
- [ ] 優先度順でのソートが可能 (US-017で実装予定)

## 完了のエビデンス

### 1. 優先度フィルタリングロジックの実装

`internal/app/app.go`に`GetFilteredTasksByPriority`メソッドを実装し、指定された優先度のリストに基づいてタスクをフィルタリングする機能を追加しました。

### 2. 優先度選択UIの実装

`cmd/main.go`の`model`構造体に`filterPriorityInput`と`filteredPriorities`フィールドを追加し、`initialModel`で初期化しました。
`Update`メソッドに`p`キーでの優先度フィルタリングモードへの遷移と、フィルタリングモードでの入力処理（優先度の選択、確定、キャンセル）を追加しました。

### 3. フィルタ条件表示機能とフィルタリング結果の表示

`cmd/main.go`の`View`メソッドを更新し、優先度フィルタリングモードのUIと、優先度フィルタリングが適用されている場合に現在のフィルタ条件を表示するようにしました。
また、`m.tasks`がフィルタリング結果で更新されるようにしました。

### 4. 単体テストの作成

`internal/app/app_test.go`に`TestGetFilteredTasksByPriority`テスト関数を追加し、様々な優先度のタスクを作成し、異なるフィルタ条件で`GetFilteredTasksByPriority`を呼び出し、期待される結果が返されることを確認しました。

全てのテストがパスしたことを確認済みです。