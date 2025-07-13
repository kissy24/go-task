# US-020: データインポート機能 完了報告

## 完了定義

- [x] JSON形式からインポートできる
- [x] 既存データとの重複チェックが可能
- [x] インポートの成功/失敗が表示される

## 完了のエビデンス

### 実装内容

- `internal/app/app.go` に `ImportTasks` 関数を実装し、指定されたファイルパスからタスクデータをインポートできるようにしました。
- `ImportTasks` 関数内で既存タスクとのID重複チェックを行い、重複しないタスクのみを追加するようにしました。
- `cmd/main.go` の `model` 構造体に `importInput` フィールドを追加し、メインビューのフッターにインポートオプションの表示を追加しました。
- `cmd/main.go` に `i` キーによるインポートビューへの遷移、およびインポートロジックを実装しました。

### テスト結果

`go test -v ./internal/app` を実行し、`TestImportTasks` を含む全てのテストがパスすることを確認しました。

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
=== RUN   TestImportTasks
--- PASS: TestImportTasks (0.00s)
PASS
ok  	zan/internal/app	0.000s