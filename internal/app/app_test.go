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
	oldTestEnv := os.Getenv("ZAN_TEST_ENV")
	os.Setenv("HOME", tempDir)
	os.Setenv("ZAN_TEST_ENV", "true") // テスト環境であることを示す
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("ZAN_TEST_ENV", oldTestEnv) // 元に戻す
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

func TestGetFilteredTasksByStatus(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "zan_test_filter_status_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	// テスト用のタスクを追加
	task1, _ := app.AddTask("Task TODO High", "", task.PriorityHigh, nil)
	task2, _ := app.AddTask("Task IN_PROGRESS Medium", "", task.PriorityMedium, nil)
	task3, _ := app.AddTask("Task DONE Low", "", task.PriorityLow, nil)
	task4, _ := app.AddTask("Task PENDING High", "", task.PriorityHigh, nil)
	task5, _ := app.AddTask("Task TODO Medium", "", task.PriorityMedium, nil)

	app.UpdateTask(task1.ID, "", "", task.StatusTODO, "", nil)
	app.UpdateTask(task2.ID, "", "", task.StatusInProgress, "", nil)
	app.UpdateTask(task3.ID, "", "", task.StatusDone, "", nil)
	app.UpdateTask(task4.ID, "", "", task.StatusPending, "", nil)
	app.UpdateTask(task5.ID, "", "", task.StatusTODO, "", nil)

	tests := []struct {
		name     string
		statuses []task.Status
		expected int
	}{
		{
			name:     "Filter by TODO",
			statuses: []task.Status{task.StatusTODO},
			expected: 2,
		},
		{
			name:     "Filter by IN_PROGRESS and PENDING",
			statuses: []task.Status{task.StatusInProgress, task.StatusPending},
			expected: 2,
		},
		{
			name:     "Filter by DONE",
			statuses: []task.Status{task.StatusDone},
			expected: 1,
		},
		{
			name:     "No filter (empty statuses)",
			statuses: []task.Status{},
			expected: 5, // 全てのタスク
		},
		{
			name:     "No matching status",
			statuses: []task.Status{"NON_EXISTENT"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filteredTasks := app.GetFilteredTasksByStatus(tt.statuses)
			if len(filteredTasks) != tt.expected {
				t.Errorf("GetFilteredTasksByStatus() got %d tasks, want %d for statuses %v", len(filteredTasks), tt.expected, tt.statuses)
			}

			// 各タスクが正しいステータスを持っているか確認 (オプション)
			for _, ft := range filteredTasks {
				found := false
				for _, s := range tt.statuses {
					if ft.Status == s {
						found = true
						break
					}
				}
				if !found && len(tt.statuses) > 0 { // 空のstatusesの場合はチェックしない
					t.Errorf("Task %s has unexpected status %s for filter %v", ft.Title, ft.Status, tt.statuses)
				}
			}
		})
	}
}

func TestGetFilteredTasksByPriority(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "zan_test_filter_priority_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	// テスト用のタスクを追加
	app.AddTask("Task High 1", "", task.PriorityHigh, nil)
	app.AddTask("Task Medium 1", "", task.PriorityMedium, nil)
	app.AddTask("Task Low 1", "", task.PriorityLow, nil)
	app.AddTask("Task High 2", "", task.PriorityHigh, nil)
	app.AddTask("Task Medium 2", "", task.PriorityMedium, nil)

	tests := []struct {
		name       string
		priorities []task.Priority
		expected   int
	}{
		{
			name:       "Filter by HIGH",
			priorities: []task.Priority{task.PriorityHigh},
			expected:   2,
		},
		{
			name:       "Filter by MEDIUM and LOW",
			priorities: []task.Priority{task.PriorityMedium, task.PriorityLow},
			expected:   3,
		},
		{
			name:       "No filter (empty priorities)",
			priorities: []task.Priority{},
			expected:   5, // 全てのタスク
		},
		{
			name:       "No matching priority",
			priorities: []task.Priority{"NON_EXISTENT"},
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filteredTasks := app.GetFilteredTasksByPriority(tt.priorities)
			if len(filteredTasks) != tt.expected {
				t.Errorf("GetFilteredTasksByPriority() got %d tasks, want %d for priorities %v", len(filteredTasks), tt.expected, tt.priorities)
			}
			// 各タスクが正しい優先度を持っているか確認 (オプション)
			for _, ft := range filteredTasks {
				found := false
				for _, p := range tt.priorities {
					if ft.Priority == p {
						found = true
						break
					}
				}
				if !found && len(tt.priorities) > 0 { // 空のprioritiesの場合はチェックしない
					t.Errorf("Task %s has unexpected priority %s for filter %v", ft.Title, ft.Priority, tt.priorities)
				}
			}
		})
	}
}
