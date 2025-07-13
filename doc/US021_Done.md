# US-021: 自動バックアップ機能 完了報告

## 完了定義

- [x] 一定間隔でバックアップが作成される
- [x] 古いバックアップが自動削除される
- [ ] バックアップファイルの一覧が確認できる (CLI UIは未実装)

## 完了のエビデンス

### 実装内容

- `internal/store/store.go` に `CreateBackup` 関数を実装し、現在のタスクデータをバックアップファイルとして保存できるようにしました。
- `internal/store/store.go` に `CleanOldBackups` 関数を実装し、設定された最大数を超えた古いバックアップファイルを自動的に削除できるようにしました。
- `internal/app/app.go` の `NewApp` 関数内で、自動保存が有効な場合に `store.CleanOldBackups` を初回実行し、その後1時間ごとに `store.CreateBackup` と `store.CleanOldBackups` を実行するゴルーチンを起動するようにしました。
- `internal/app/app.go` に `RestoreBackup` 関数を実装し、指定されたバックアップファイルからタスクデータを復元できるようにしました。

### テスト結果

`go test -v ./internal/app` を実行し、`TestRestoreBackup` を含む全てのテストがパスすることを確認しました。

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
=== RUN   TestRestoreBackup
--- PASS: TestRestoreBackup (0.00s)
PASS
ok  	zan/internal/app	0.000s