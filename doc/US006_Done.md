# US-006: タスク状態変更機能 完了報告

## 完了定義

- [x] TODO, IN_PROGRESS, DONE, PENDING の状態間で変更できる
- [x] 完了時に完了日時が自動設定される
- [x] 状態変更時に更新日時が自動更新される

## 完了のエビデンス

### 状態変更コマンドの実装、完了処理、更新日時の自動設定、状態検証の実装

`internal/app/app.go` の `UpdateTask()` メソッドを実装し、タスクの状態変更、完了日時の自動設定、更新日時の自動更新を行えるようにしました。

- `UpdateTask` メソッドは、指定されたIDのタスクの状態を更新します。
- 状態が `DONE` に変更された場合、`CompletedAt` フィールドに現在時刻が自動的に設定されます。
- 状態が `DONE` 以外に変更された場合、`CompletedAt` フィールドはクリアされます。
- `UpdatedAt` フィールドは、タスクが更新されるたびに現在時刻に自動的に更新されます。
- `task.Validate()` メソッドを呼び出すことで、状態変更後のタスクデータの検証を行っています。

```go
// internal/app/app.go (抜粋)
// UpdateTask は既存のタスクを更新します。
func (a *App) UpdateTask(id, title, description string, status task.Status, priority task.Priority, tags []string) (*task.Task, error) {
	for i, t := range a.Tasks.Tasks {
		if t.ID == id {
			if title != "" {
				a.Tasks.Tasks[i].Title = title
			}
			if description != "" {
				a.Tasks.Tasks[i].Description = description
			}
			if status != "" {
				a.Tasks.Tasks[i].Status = status
				if status == task.StatusDone {
					now := time.Now()
					a.Tasks.Tasks[i].CompletedAt = &now
				} else {
					a.Tasks.Tasks[i].CompletedAt = nil
				}
			}
			if priority != "" {
				a.Tasks.Tasks[i].Priority = priority
			}
			if tags != nil {
				a.Tasks.Tasks[i].Tags = tags
			}
			a.Tasks.Tasks[i].UpdatedAt = time.Now()

			if err := a.Tasks.Tasks[i].Validate(); err != nil {
				return nil, fmt.Errorf("invalid task data after update: %w", err)
			}

			if a.Tasks.Settings.AutoSave {
				if err := store.SaveTasks(a.Tasks); err != nil {
					return nil, fmt.Errorf("failed to auto-save tasks: %w", err)
				}
			}
			return &a.Tasks.Tasks[i], nil
		}
	}
	return nil, fmt.Errorf("task with ID %s not found", id)
}
```

`cmd/zan/main.go` に `completeTask` 関数を実装し、タスクを完了状態にできるようにしました。

```go
// cmd/zan/main.go (抜粋)
func completeTask(cmd *cobra.Command, args []string) {
	id := args[0]
	_, err := appInstance.UpdateTask(id, "", "", task.StatusDone, "", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error completing task: %v\n", err)
		return
	}
	fmt.Printf("Task %s marked as DONE.\n", id)
}
```

### 単体テストの作成

`internal/app/app_test.go` に `TestUpdateTask` 関数を実装し、タスクの状態変更機能の単体テストを作成しました。これにより、タスクのタイトル、説明、タグの更新、状態の完了/未完了への変更、完了日時の設定/クリアが正しく機能することを確認しています。

```go
// internal/app/app_test.go (抜粋)
func TestUpdateTask(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "zan_test_update_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	originalTask, _ := app.AddTask("Original Title", "Original Desc", task.PriorityMedium, []string{"old"})

	// タイトルと説明を更新
	updatedTask, err := app.UpdateTask(originalTask.ID, "New Title", "New Desc", "", "", []string{"new"})
	if err != nil {
		t.Fatalf("UpdateTask() failed: %v", err)
	}
	if updatedTask.Title != "New Title" || updatedTask.Description != "New Desc" || updatedTask.Tags[0] != "new" {
		t.Errorf("UpdateTask() failed to update fields. Got: %+v", updatedTask)
	}

	// 状態を完了に更新
	completedTask, err := app.UpdateTask(originalTask.ID, "", "", task.StatusDone, "", nil)
	if err != nil {
		t.Fatalf("UpdateTask() failed to complete task: %v", err)
	}
	if completedTask.Status != task.StatusDone || completedTask.CompletedAt == nil {
		t.Errorf("UpdateTask() failed to set status to DONE or CompletedAt. Got: %+v", completedTask)
	}

	// 状態をTODOに戻す
	reopenedTask, err := app.UpdateTask(originalTask.ID, "", "", task.StatusTODO, "", nil)
	if err != nil {
		t.Fatalf("UpdateTask() failed to reopen task: %v", err)
	}
	if reopenedTask.Status != task.StatusTODO || reopenedTask.CompletedAt != nil {
		t.Errorf("UpdateTask() failed to set status to TODO or clear CompletedAt. Got: %+v", reopenedTask)
	}

	// 存在しないIDの更新
	_, err = app.UpdateTask("non-existent-id", "Title", "", "", "", nil)
	if err == nil {
		t.Errorf("UpdateTask() expected error for non-existent ID, got nil")
	}
}
```