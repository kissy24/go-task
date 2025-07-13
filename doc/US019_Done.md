# US-019: データエクスポート機能 完了報告

## 完了定義

- [x] JSON形式でエクスポートできる
- [x] 出力先を指定できる
- [x] エクスポートの成功/失敗が表示される

## 完了のエビデンス

### 実装内容

- `internal/store/store.go` に `MarshalTasks` 関数を追加し、タスクデータをJSON形式にマーシャルできるようにしました。また、`ioutil` の使用箇所を `os` に変更しました。
- `internal/app/app.go` に `ExportTasks` 関数を実装し、指定されたファイルパスにタスクデータをエクスポートできるようにしました。
- `cmd/main.go` の `model` 構造体に `exportInput` フィールドを追加し、メインビューのフッターにエクスポートオプションの表示を追加しました。
- `cmd/main.go` に `x` キーによるエクスポートビューへの遷移、およびエクスポートロジックを実装しました。

### テスト結果

`go test -v ./internal/app` を実行し、`TestExportTasks` を含む全てのテストがパスすることを確認しました。

```
=== RUN   TestNewApp
--- PASS: TestNewApp (0.00s)
=== RUN   TestAddTask
--- PASS: TestAddTask (0.00s)
=== RUN   TestGetTaskByID
--- PASS: TestGetTaskByID (0.00s)
=== RUN   TestUpdateTask
--- PASS: TestUpdateTask (0.00s)
=== RUN   TestDeleteTask
--- PASS: TestDeleteTask (0.00s)
=== RUN   TestGetAllTasks
--- PASS: TestGetAllTasks (0.00s)
=== RUN   TestGetTaskStats
--- PASS: TestGetTaskStats (0.00s)
=== RUN   TestGetFilteredTasksByStatus
--- PASS: TestGetFilteredTasksByStatus (0.00s)
=== RUN   TestGetFilteredTasksByPriority
--- PASS: TestGetFilteredTasksByPriority (0.00s)
=== RUN   TestGetFilteredTasksByTags
--- PASS: TestGetFilteredTasksByTags (0.00s)
=== RUN   TestSearchTasks
--- PASS: TestSearchTasks (0.00s)
=== RUN   TestGetAllUniqueTags
--- PASS: TestGetAllUniqueTags (0.00s)
=== RUN   TestSortTasks
--- PASS: TestSortTasks (0.00s)
=== RUN   TestExportTasks
--- PASS: TestExportTasks (0.00s)
PASS
ok  	zan/internal/app	0.000s