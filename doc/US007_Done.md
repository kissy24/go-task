# US-007: タスク削除機能 完了報告

## 完了定義

- [x] タスクIDを指定してタスクを削除できる
- [x] 削除前に確認メッセージが表示される
- [x] 削除後は元に戻せない旨が明示される

## 完了のエビデンス

### 削除コマンドの実装、確認プロンプトの実装、タスク除去処理の実装

`internal/app/app.go` に `DeleteTask()` メソッドを実装し、指定されたIDのタスクを削除できるようにしました。

- `DeleteTask` メソッドは、指定されたIDのタスクをタスクリストから削除します。
- タスク削除後、`store.SaveTasks()` を呼び出すことで自動保存されます。

```go
// internal/app/app.go (抜粋)
// DeleteTask は指定されたIDのタスクを削除します。
func (a *App) DeleteTask(id string) error {
	for i, t := range a.Tasks.Tasks {
		if t.ID == id {
			a.Tasks.Tasks = append(a.Tasks.Tasks[:i], a.Tasks.Tasks[i+1:]...)
			if a.Tasks.Settings.AutoSave {
				if err := store.SaveTasks(a.Tasks); err != nil {
					return fmt.Errorf("failed to auto-save tasks after deletion: %w", err)
				}
			}
			return nil
		}
	}
	return fmt.Errorf("task with ID %s not found", id)
}
```

`cmd/zan/main.go` に `deleteTask` 関数を実装し、タスク削除機能を提供します。

- `deleteTask` 関数は、タスクIDを受け取り、削除前にユーザーに確認メッセージを表示します。
- ユーザーが「yes」と入力した場合のみ、`appInstance.DeleteTask()` を呼び出してタスクを削除します。
- 削除後は元に戻せない旨をメッセージで明示しています。

```go
// cmd/zan/main.go (抜粋)
func deleteTask(cmd *cobra.Command, args []string) {
	id := args[0]
	fmt.Printf("Are you sure you want to delete task %s? This action cannot be undone. (yes/no): ", id)
	var confirmation string
	fmt.Scanln(&confirmation)

	if strings.ToLower(confirmation) != "yes" {
		fmt.Println("Task deletion cancelled.")
		return
	}

	err := appInstance.DeleteTask(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting task: %v\n", err)
		return
	}
	fmt.Printf("Task %s deleted.\n", id)
}
```

### 単体テストの作成

`internal/app/app_test.go` に `TestDeleteTask` 関数を実装し、タスク削除機能の単体テストを作成しました。これにより、存在するIDのタスクが正しく削除されること、および存在しないIDのタスクを削除しようとした場合にエラーが返されることを確認しています。

```go
// internal/app/app_test.go (抜粋)
func TestDeleteTask(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "zan_test_delete_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	task1, _ := app.AddTask("Task 1", "", "", nil)
	app.AddTask("Task 2", "", "", nil)

	// 存在するIDを削除
	err = app.DeleteTask(task1.ID)
	if err != nil {
		t.Fatalf("DeleteTask() failed: %v", err)
	}
	if len(app.Tasks.Tasks) != 1 {
		t.Errorf("Expected 1 task after deletion, got %d", len(app.Tasks.Tasks))
	}
	_, err = app.GetTaskByID(task1.ID)
	if err == nil {
		t.Errorf("Deleted task found by ID")
	}

	// 存在しないIDを削除
	err = app.DeleteTask("non-existent-id")
	if err == nil {
		t.Errorf("DeleteTask() expected error for non-existent ID, got nil")
	}
}