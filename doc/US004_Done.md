# US-004: タスク一覧表示機能 完了報告

## 完了定義

- [x] 全タスクが一覧表示される
- [x] 各タスクの状態、優先度、タイトルが表示される
- [x] 統計情報（合計、未完了、完了数）が表示される

## 完了のエビデンス

### 一覧表示コマンドの実装、タスク情報のフォーマット処理、統計情報の計算処理、テーブル形式での出力機能

`internal/app/app.go` に `GetTaskStats()` メソッドを追加し、タスクの統計情報を計算できるようにしました。

```go
// internal/app/app.go (抜粋)
// GetTaskStats はタスクの統計情報を返します。
func (a *App) GetTaskStats() (total, completed, incomplete int) {
	total = len(a.Tasks.Tasks)
	for _, t := range a.Tasks.Tasks {
		if t.Status == task.StatusDone {
			completed++
		} else {
			incomplete++
		}
	}
	return
}
```

`cmd/zan/main.go` に `listTasks` 関数を実装し、タスクの一覧表示と統計情報の表示を行えるようにしました。
- `listTasks` 関数は、`appInstance.GetAllTasks()` で全タスクを取得し、`appInstance.GetTaskStats()` で統計情報を取得します。
- 各タスクはID、状態アイコン、優先度（色付き）、タイトルがテーブル形式で表示されます。
- 画面下部には合計、未完了、完了のタスク数が表示されます。
- 状態アイコンと優先度カラーはそれぞれ `getStatusIcon` と `getPriorityColor` 関数で定義しています。

```go
// cmd/zan/main.go (抜粋)
func listTasks(cmd *cobra.Command, args []string) {
	tasks := appInstance.GetAllTasks()
	total, completed, incomplete := appInstance.GetTaskStats()

	if len(tasks) == 0 {
		fmt.Println("No tasks found. Add a new task using 'zan add <title>'")
		return
	}

	fmt.Println("ID         Status    Priority  Title")
	fmt.Println("--------------------------------------------------")
	for _, t := range tasks {
		statusIcon := getStatusIcon(t.Status)
		priorityColor := getPriorityColor(t.Priority)
		fmt.Printf("%-10s %-9s %s%-9s\033[0m %s\n", t.ID[:8], statusIcon, priorityColor, t.Priority, t.Title)
	}
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Total: %d | Incomplete: %d | Completed: %d\n", total, incomplete, completed)
}

func getStatusIcon(status task.Status) string {
	switch status {
	case task.StatusTODO:
		return "●"
	case task.StatusInProgress:
		return "◐"
	case task.StatusDone:
		return "✓"
	case task.StatusPending:
		return "⏸"
	default:
		return "?"
	}
}

func getPriorityColor(priority task.Priority) string {
	// ANSI escape codes for colors
	const (
		Red    = "\033[31m"
		Yellow = "\033[33m"
		Green  = "\033[32m"
		Reset  = "\033[0m"
	)
	switch priority {
	case task.PriorityHigh:
		return Red
	case task.PriorityMedium:
		return Yellow
	case task.PriorityLow:
		return Green
	default:
		return Reset
	}
}
```

### 単体テストの作成

`internal/app/app_test.go` に `TestGetTaskStats` 関数を実装し、統計情報計算の単体テストを作成しました。これにより、タスクの合計数、完了数、未完了数が正しく計算されることを確認しています。

```go
// internal/app/app_test.go (抜粋)
func TestGetTaskStats(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "zan_test_stats_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	app.AddTask("Task 1", "", "", nil)
	app.AddTask("Task 2", "", "", nil)
	task3, _ := app.AddTask("Task 3", "", "", nil)
	app.UpdateTask(task3.ID, "", "", task.StatusDone, "", nil) // 1つ完了

	total, completed, incomplete := app.GetTaskStats()

	if total != 3 {
		t.Errorf("GetTaskStats() expected total 3, got %d", total)
	}
	if completed != 1 {
		t.Errorf("GetTaskStats() expected completed 1, got %d", completed)
	}
	if incomplete != 2 {
		t.Errorf("GetTaskStats() expected incomplete 2, got %d", incomplete)
	}
}