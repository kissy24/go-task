# US-005: タスク詳細表示機能 完了報告

## 完了定義

- [x] タスクIDを指定して詳細を表示できる
- [x] 全ての属性が整理されて表示される
- [x] 存在しないIDの場合は適切なエラーメッセージが表示される

## 完了のエビデンス

### 詳細表示コマンドの実装、タスク検索機能、詳細情報のフォーマット処理、エラーハンドリングの実装

`internal/app/app.go` に `GetTaskByID()` メソッドを実装し、指定されたIDのタスクを取得できるようにしました。タスクが見つからない場合はエラーを返します。

```go
// internal/app/app.go (抜粋)
// GetTaskByID は指定されたIDのタスクを返します。
func (a *App) GetTaskByID(id string) (*task.Task, error) {
	for _, t := range a.Tasks.Tasks {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("task with ID %s not found", id)
}
```

`cmd/go-task/main.go` に `showTask` 関数を実装し、タスクの詳細表示を行えるようにしました。
- `showTask` 関数は、引数からタスクIDを取得し、`appInstance.GetTaskByID()` を呼び出してタスクを取得します。
- 取得したタスクの全ての属性（ID, Title, Description, Status, Priority, Tags, CreatedAt, UpdatedAt, CompletedAt）を整理された形式で表示します。
- 存在しないIDが指定された場合は、適切なエラーメッセージを標準エラー出力に表示します。

```go
// cmd/go-task/main.go (抜粋)
func showTask(cmd *cobra.Command, args []string) {
	id := args[0]
	t, err := appInstance.GetTaskByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error showing task: %v\n", err)
		return
	}

	fmt.Println("Task Details:")
	fmt.Printf("  ID:          %s\n", t.ID)
	fmt.Printf("  Title:       %s\n", t.Title)
	fmt.Printf("  Description: %s\n", t.Description)
	fmt.Printf("  Status:      %s %s\n", getStatusIcon(t.Status), t.Status)
	fmt.Printf("  Priority:    %s%s\033[0m %s\n", getPriorityColor(t.Priority), t.Priority, t.Priority)
	fmt.Printf("  Tags:        %s\n", strings.Join(t.Tags, ", "))
	fmt.Printf("  Created At:  %s\n", t.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Updated At:  %s\n", t.UpdatedAt.Format("2006-01-02 15:04:05"))
	if t.CompletedAt != nil {
		fmt.Printf("  Completed At: %s\n", t.CompletedAt.Format("2006-01-02 15:04:05"))
	}
}
```

### 単体テストの作成

`internal/app/app_test.go` に `TestGetTaskByID` 関数を実装し、タスク検索機能の単体テストを作成しました。これにより、存在するIDと存在しないIDの両方でタスクの取得が正しく機能することを確認しています。

```go
// internal/app/app_test.go (抜粋)
func TestGetTaskByID(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "go-task_test_get_")
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

	// 存在するIDで取得
	foundTask, err := app.GetTaskByID(task1.ID)
	if err != nil {
		t.Fatalf("GetTaskByID() failed for existing ID: %v", err)
	}
	if foundTask.ID != task1.ID {
		t.Errorf("GetTaskByID() got ID %s, want %s", foundTask.ID, task1.ID)
	}

	// 存在しないIDで取得
	_, err = app.GetTaskByID("non-existent-id")
	if err == nil {
		t.Errorf("GetTaskByID() expected error for non-existent ID, got nil")
	}
}