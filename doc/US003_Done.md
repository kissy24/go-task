# US-003: タスク作成機能 完了報告

## 完了定義

- [x] タイトルを必須として新しいタスクを作成できる
- [x] 詳細説明、優先度、タグを任意で設定できる
- [x] 作成時にIDと作成日時が自動設定される

## 完了のエビデンス

### タスク作成コマンドの実装、UUID生成機能、日時自動設定機能、入力値検証の実装

`internal/app/app.go` に `AddTask` メソッドを実装しました。

- `AddTask` メソッドは、タイトルを必須とし、詳細説明、優先度、タグを任意で受け取ります。
- 新しいタスクには `github.com/google/uuid` を使用して一意のIDを自動生成します。
- `CreatedAt` と `UpdatedAt` はタスク作成時に自動的に現在時刻に設定されます。
- `task.Validate()` メソッドを呼び出すことで、入力値の検証を行っています。
- `store.SaveTasks()` を呼び出すことで、タスクの自動保存も行われます。

```go
// internal/app/app.go
package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"zan/internal/store"
	"zan/internal/task"
)

// App はアプリケーションの主要なロジックを管理します。
type App struct {
	Tasks *task.Tasks
}

// NewApp は新しいAppインスタンスを作成し、タスクデータをロードします。
func NewApp() (*App, error) {
	tasks, err := store.LoadTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}
	return &App{Tasks: tasks}, nil
}

// AddTask は新しいタスクを作成し、タスクリストに追加します。
func (a *App) AddTask(title, description string, priority task.Priority, tags []string) (*task.Task, error) {
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	if priority == "" {
		priority = a.Tasks.Settings.DefaultPriority
	}

	newTask := task.Task{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Status:      task.StatusTODO,
		Priority:    priority,
		Tags:        tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := newTask.Validate(); err != nil {
		return nil, fmt.Errorf("invalid task data: %w", err)
	}

	a.Tasks.Tasks = append(a.Tasks.Tasks, newTask)
	if a.Tasks.Settings.AutoSave {
		if err := store.SaveTasks(a.Tasks); err != nil {
			return nil, fmt.Errorf("failed to auto-save tasks: %w", err)
		}
	}
	return &newTask, nil
}

// GetTaskByID は指定されたIDのタスクを返します。
func (a *App) GetTaskByID(id string) (*task.Task, error) {
	for _, t := range a.Tasks.Tasks {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("task with ID %s not found", id)
}

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

// GetAllTasks は全てのタスクを返します。
func (a *App) GetAllTasks() []task.Task {
	return a.Tasks.Tasks
}
```

### 単体テストの作成

`internal/app/app_test.go` に `TestAddTask` 関数を実装し、`AddTask` メソッドの単体テストを作成しました。これにより、有効なタスクの追加、タイトルが空の場合のエラーハンドリング、デフォルト優先度の設定が正しく機能することを確認しています。

```go
// internal/app/app_test.go
package app

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"zan/internal/store"
	"zan/internal/task"
)

// setupTestEnv はテスト用の環境変数を設定し、テスト終了後に元に戻します。
func setupTestEnv(t *testing.T, tempDir string) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})
}

func TestNewApp(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "go-task_test_app_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v, wantErr %v", err, false)
	}
	if app == nil || app.Tasks == nil {
		t.Errorf("NewApp() returned nil app or tasks")
	}
	if len(app.Tasks.Tasks) != 0 {
		t.Errorf("NewApp() expected empty tasks, got %d", len(app.Tasks.Tasks))
	}
}

func TestAddTask(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "go-task_test_add_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	// 有効なタスクの追加
	addedTask, err := app.AddTask("Test Task 1", "Description 1", task.PriorityHigh, []string{"tag1", "tag2"})
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}
	if addedTask.Title != "Test Task 1" {
		t.Errorf("AddTask() got title %s, want %s", addedTask.Title, "Test Task 1")
	}
	if addedTask.ID == "" {
		t.Errorf("AddTask() ID was not generated")
	}
	if addedTask.Status != task.StatusTODO {
		t.Errorf("AddTask() got status %s, want %s", addedTask.Status, task.StatusTODO)
	}
	if addedTask.Priority != task.PriorityHigh {
		t.Errorf("AddTask() got priority %s, want %s", addedTask.Priority, task.PriorityHigh)
	}
	if len(app.Tasks.Tasks) != 1 {
		t.Errorf("Expected 1 task in list, got %d", len(app.Tasks.Tasks))
	}

	// タイトルが空のタスクの追加
	_, err = app.AddTask("", "Description 2", task.PriorityMedium, nil)
	if err == nil {
		t.Errorf("AddTask() expected error for empty title, got nil")
	}

	// デフォルト優先度の確認
	app.Tasks.Settings.DefaultPriority = task.PriorityLow
	addedTask2, err := app.AddTask("Test Task 2", "", "", nil)
	if err != nil {
		t.Fatalf("AddTask() failed for default priority: %v", err)
	}
	if addedTask2.Priority != task.PriorityLow {
		t.Errorf("AddTask() got default priority %s, want %s", addedTask2.Priority, task.PriorityLow)
	}
}

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

func TestUpdateTask(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "go-task_test_update_")
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

func TestDeleteTask(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "go-task_test_delete_")
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

func TestGetAllTasks(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "go-task_test_getall_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	app.AddTask("Task A", "", "", nil)
	app.AddTask("Task B", "", "", nil)

	tasks := app.GetAllTasks()
	if len(tasks) != 2 {
		t.Errorf("GetAllTasks() expected 2 tasks, got %d", len(tasks))
	}
}