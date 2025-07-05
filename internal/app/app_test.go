package app

import (
	"io/ioutil"
	"os"
	"testing"

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
	tmpDir, err := ioutil.TempDir("", "zan_test_app_")
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
	tmpDir, err := ioutil.TempDir("", "zan_test_add_")
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
	tmpDir, err := ioutil.TempDir("", "zan_test_get_")
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

func TestGetAllTasks(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "zan_test_getall_")
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
