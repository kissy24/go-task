package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

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
	tmpDir, err := os.MkdirTemp("", "zan_test_app_")
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
	tmpDir, err := os.MkdirTemp("", "zan_test_add_")
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
	tmpDir, err := os.MkdirTemp("", "zan_test_get_")
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
	tmpDir, err := os.MkdirTemp("", "zan_test_update_")
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
	tmpDir, err := os.MkdirTemp("", "zan_test_delete_")
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
	tmpDir, err := os.MkdirTemp("", "zan_test_getall_")
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
	tmpDir, err := os.MkdirTemp("", "zan_test_stats_")
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
	tmpDir, err := os.MkdirTemp("", "zan_test_filter_status_")
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
	tmpDir, err := os.MkdirTemp("", "zan_test_filter_priority_")
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

func TestGetFilteredTasksByTags(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zan_test_filter_tags_")
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
	app.AddTask("Task with tag A", "", task.PriorityHigh, []string{"work", "urgent"})
	app.AddTask("Task with tag B", "", task.PriorityMedium, []string{"personal"})
	app.AddTask("Task with tag C", "", task.PriorityLow, []string{"work"})
	app.AddTask("Task with tag D", "", task.PriorityHigh, []string{"personal", "urgent"})
	app.AddTask("Task with no tags", "", task.PriorityMedium, []string{})

	tests := []struct {
		name     string
		tags     []string
		expected int
	}{
		{
			name:     "Filter by single tag 'work'",
			tags:     []string{"work"},
			expected: 2,
		},
		{
			name:     "Filter by single tag 'personal'",
			tags:     []string{"personal"},
			expected: 2,
		},
		{
			name:     "Filter by multiple tags 'work' AND 'urgent'",
			tags:     []string{"work", "urgent"},
			expected: 1,
		},
		{
			name:     "Filter by multiple tags 'personal' AND 'urgent'",
			tags:     []string{"personal", "urgent"},
			expected: 1,
		},
		{
			name:     "No filter (empty tags)",
			tags:     []string{},
			expected: 5, // 全てのタスク
		},
		{
			name:     "No matching tag",
			tags:     []string{"nonexistent"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filteredTasks := app.GetFilteredTasksByTags(tt.tags)
			if len(filteredTasks) != tt.expected {
				t.Errorf("GetFilteredTasksByTags() got %d tasks, want %d for tags %v", len(filteredTasks), tt.expected, tt.tags)
			}
			// 各タスクが正しいタグを持っているか確認 (オプション)
			for _, ft := range filteredTasks {
				for _, filterTag := range tt.tags {
					found := false
					for _, taskTag := range ft.Tags {
						if strings.EqualFold(strings.TrimSpace(filterTag), strings.TrimSpace(taskTag)) {
							found = true
							break
						}
					}
					if !found && len(tt.tags) > 0 {
						t.Errorf("Task %s does not have expected tag %s for filter %v", ft.Title, filterTag, tt.tags)
					}
				}
			}
		})
	}
}

func TestSearchTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zan_test_search_")
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
	app.AddTask("Buy groceries", "Milk, eggs, bread", task.PriorityHigh, nil)
	app.AddTask("Finish report", "Complete Q3 financial report", task.PriorityMedium, nil)
	app.AddTask("Call John", "Discuss project updates", task.PriorityLow, nil)
	app.AddTask("Prepare presentation", "Review slides for meeting", task.PriorityHigh, nil)
	app.AddTask("Grocery shopping list", "Fruits and vegetables", task.PriorityMedium, nil)

	// デバッグログ
	for i, taskItem := range app.Tasks.Tasks { // 変数名を変更して衝突を避ける
		t.Logf("Task %d: ID=%s, Title='%s', Description='%s'", i, taskItem.ID, taskItem.Title, taskItem.Description)
	}

	tests := []struct {
		name     string
		keyword  string
		expected []string // 期待されるタスクのタイトル
	}{
		{
			name:     "Search by title keyword 'report'",
			keyword:  "report",
			expected: []string{"Finish report"},
		},
		{
			name:     "Search by description keyword 'milk'",
			keyword:  "milk",
			expected: []string{"Buy groceries"},
		},
		{
			name:     "Case-insensitive search 'grocery'",
			keyword:  "grocery",
			expected: []string{"Grocery shopping list"}, // "Buy groceries"は"grocery"を含まない
		},
		{
			name:     "Search by partial keyword 'proj'",
			keyword:  "proj",
			expected: []string{"Call John"},
		},
		{
			name:     "No matching keyword",
			keyword:  "nonexistent",
			expected: []string{},
		},
		{
			name:     "Empty keyword returns all tasks",
			keyword:  "",
			expected: []string{"Buy groceries", "Finish report", "Call John", "Prepare presentation", "Grocery shopping list"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundTasks := app.Search(tt.keyword)
			t.Logf("Search() for keyword '%s' returned %d tasks:", tt.keyword, len(foundTasks))
			for _, ft := range foundTasks {
				t.Logf("  - Found Task: Title='%s'", ft.Title)
			}

			if len(foundTasks) != len(tt.expected) {
				t.Errorf("Search() got %d tasks, want %d for keyword '%s'", len(foundTasks), len(tt.expected), tt.keyword)
			}

			foundTitles := make([]string, len(foundTasks))
			for i, t := range foundTasks {
				foundTitles[i] = t.Title
			}
			// 順序は保証されないため、ソートして比較
			sort.Strings(foundTitles)
			sort.Strings(tt.expected) // 期待値もソート

			for i, expectedTitle := range tt.expected {
				if foundTitles[i] != expectedTitle {
					t.Errorf("Search() for keyword '%s', at index %d got '%s', want '%s'", tt.keyword, i, foundTitles[i], expectedTitle)
				}
			}
		})
	}
}

func TestGetAllUniqueTags(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zan_test_unique_tags_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	app.AddTask("Task 1", "", task.PriorityHigh, []string{"work", "urgent", "personal"})
	app.AddTask("Task 2", "", task.PriorityMedium, []string{"personal", "home"})
	app.AddTask("Task 3", "", task.PriorityLow, []string{"work"})
	app.AddTask("Task 4", "", task.PriorityHigh, []string{"urgent"})
	app.AddTask("Task 5", "", task.PriorityMedium, []string{}) // タグなし

	expectedTags := []string{"home", "personal", "urgent", "work"}
	actualTags := app.GetAllUniqueTags()

	if len(actualTags) != len(expectedTags) {
		t.Fatalf("GetAllUniqueTags() got %d tags, want %d", len(actualTags), len(expectedTags))
	}

	for i, tag := range actualTags {
		if tag != expectedTags[i] {
			t.Errorf("GetAllUniqueTags() at index %d got %s, want %s", i, tag, expectedTags[i])
		}
	}
}

func TestSortTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zan_test_sort_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	// テスト用のタスクを追加 (順序が重要)
	task1, _ := app.AddTask("Task C", "", task.PriorityLow, nil)
	task2, _ := app.AddTask("Task A", "", task.PriorityHigh, nil)
	task3, _ := app.AddTask("Task B", "", task.PriorityMedium, nil)

	// 作成日時を調整してソート順を明確にする
	task1.CreatedAt = task1.CreatedAt.Add(3 * time.Hour)
	task2.CreatedAt = task2.CreatedAt.Add(1 * time.Hour)
	task3.CreatedAt = task3.CreatedAt.Add(2 * time.Hour)

	// 更新日時を調整
	task1.UpdatedAt = task1.UpdatedAt.Add(1 * time.Hour)
	task2.UpdatedAt = task2.UpdatedAt.Add(3 * time.Hour)
	task3.UpdatedAt = task3.UpdatedAt.Add(2 * time.Hour)

	app.Tasks.Tasks = []task.Task{*task1, *task2, *task3} // 順序をリセット

	tests := []struct {
		name      string
		sortBy    string
		ascending bool
		expected  []string // 期待されるタスクのタイトル順
	}{
		{
			name:      "Sort by CreatedAt Ascending",
			sortBy:    "created_at",
			ascending: true,
			expected:  []string{"Task A", "Task B", "Task C"},
		},
		{
			name:      "Sort by CreatedAt Descending",
			sortBy:    "created_at",
			ascending: false,
			expected:  []string{"Task C", "Task B", "Task A"},
		},
		{
			name:      "Sort by UpdatedAt Ascending",
			sortBy:    "updated_at",
			ascending: true,
			expected:  []string{"Task C", "Task B", "Task A"},
		},
		{
			name:      "Sort by UpdatedAt Descending",
			sortBy:    "updated_at",
			ascending: false,
			expected:  []string{"Task A", "Task B", "Task C"},
		},
		{
			name:      "Sort by Priority Ascending (Low to High)",
			sortBy:    "priority",
			ascending: true,
			expected:  []string{"Task C", "Task B", "Task A"}, // Low, Medium, High
		},
		{
			name:      "Sort by Priority Descending (High to Low)",
			sortBy:    "priority",
			ascending: false,
			expected:  []string{"Task A", "Task B", "Task C"}, // High, Medium, Low
		},
		{
			name:      "Default sort (unknown sortBy, CreatedAt Descending)",
			sortBy:    "unknown",
			ascending: true, // ascendingは無視され、CreatedAt Descendingになる
			expected:  []string{"Task C", "Task B", "Task A"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ソートは元のスライスを変更するため、毎回コピーを作成
			tasksCopy := make([]task.Task, len(app.Tasks.Tasks))
			copy(tasksCopy, app.Tasks.Tasks)

			sortedTasks := app.SortTasks(tasksCopy, tt.sortBy, tt.ascending)

			if len(sortedTasks) != len(tt.expected) {
				t.Fatalf("SortTasks() got %d tasks, want %d", len(sortedTasks), len(tt.expected))
			}

			for i, expectedTitle := range tt.expected {
				if sortedTasks[i].Title != expectedTitle {
					t.Errorf("SortTasks() for %s %s, at index %d got %s, want %s", tt.sortBy, func() string {
						if tt.ascending {
							return "Asc"
						}
						return "Desc"
					}(), i, sortedTasks[i].Title, expectedTitle)
				}
			}
		})
	}
}

func TestExportTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zan_test_export_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	// Add some tasks to export
	app.AddTask("Task 1", "Description 1", task.PriorityHigh, []string{"tag1"})
	app.AddTask("Task 2", "Description 2", task.PriorityMedium, []string{"tag2"})

	exportFilePath := filepath.Join(tmpDir, "exported_tasks.json")

	// Test successful export
	err = app.ExportTasks(exportFilePath)
	if err != nil {
		t.Fatalf("ExportTasks() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		t.Errorf("Exported file does not exist at %s", exportFilePath)
	}

	// Verify file content (basic check)
	data, err := os.ReadFile(exportFilePath)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	var exportedTasks task.Tasks
	err = json.Unmarshal(data, &exportedTasks)
	if err != nil {
		t.Fatalf("Failed to unmarshal exported data: %v", err)
	}

	if len(exportedTasks.Tasks) != 2 {
		t.Errorf("Expected 2 tasks in exported file, got %d", len(exportedTasks.Tasks))
	}
	if exportedTasks.Tasks[0].Title != "Task 1" || exportedTasks.Tasks[1].Title != "Task 2" {
		t.Errorf("Exported tasks content mismatch")
	}

	// Test export with empty file path
	err = app.ExportTasks("")
	if err == nil {
		t.Errorf("ExportTasks() expected error for empty file path, got nil")
	}
}

func TestImportTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zan_test_import_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	// Add some initial tasks to the app
	initialTask1, _ := app.AddTask("Initial Task 1", "", task.PriorityMedium, nil)

	// Create a dummy import file
	importFilePath := filepath.Join(tmpDir, "import_tasks.json")
	dummyTasks := &task.Tasks{
		Version:   "1.0.0",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tasks: []task.Task{
			{
				ID:          initialTask1.ID, // Duplicate ID
				Title:       "Duplicate Task 1",
				Description: "This should not be imported",
				Status:      task.StatusTODO,
				Priority:    task.PriorityHigh,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "new-task-id-3",
				Title:       "Imported Task 3",
				Description: "New task from import",
				Status:      task.StatusTODO,
				Priority:    task.PriorityMedium,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "new-task-id-4",
				Title:       "Imported Task 4",
				Description: "Another new task",
				Status:      task.StatusDone,
				Priority:    task.PriorityLow,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
		Settings: task.Settings{
			DefaultPriority: task.PriorityMedium,
			AutoSave:        true,
			Theme:           "default",
		},
	}

	data, err := json.MarshalIndent(dummyTasks, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal dummy tasks: %v", err)
	}
	err = os.WriteFile(importFilePath, data, 0600)
	if err != nil {
		t.Fatalf("Failed to write dummy import file: %v", err)
	}

	// Test successful import
	err = app.ImportTasks(importFilePath)
	if err != nil {
		t.Fatalf("ImportTasks() failed: %v", err)
	}

	// Verify tasks after import
	allTasks := app.GetAllTasks()
	if len(allTasks) != 4 { // 2 initial + 2 new (1 duplicate skipped)
		t.Errorf("Expected 4 tasks after import, got %d", len(allTasks))
	}

	// Check for imported tasks
	foundImported3 := false
	foundImported4 := false
	foundDuplicate1 := false
	for _, t := range allTasks {
		if t.ID == "new-task-id-3" && t.Title == "Imported Task 3" {
			foundImported3 = true
		}
		if t.ID == "new-task-id-4" && t.Title == "Imported Task 4" {
			foundImported4 = true
		}
		if t.ID == initialTask1.ID && t.Title == initialTask1.Title { // Ensure original task is still there, not overwritten by duplicate
			foundDuplicate1 = true
		}
	}

	if !foundImported3 {
		t.Errorf("Imported Task 3 not found")
	}
	if !foundImported4 {
		t.Errorf("Imported Task 4 not found")
	}
	if !foundDuplicate1 {
		t.Errorf("Original Task 1 was overwritten or not found")
	}

	// Test import with empty file path
	err = app.ImportTasks("")
	if err == nil {
		t.Errorf("ImportTasks() expected error for empty file path, got nil")
	}

	// Test import with non-existent file
	err = app.ImportTasks(filepath.Join(tmpDir, "non_existent.json"))
	if err == nil {
		t.Errorf("ImportTasks() expected error for non-existent file, got nil")
	}

	// Test import with invalid JSON
	invalidJsonPath := filepath.Join(tmpDir, "invalid.json")
	err = os.WriteFile(invalidJsonPath, []byte("{invalid json"), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid json file: %v", err)
	}
	err = app.ImportTasks(invalidJsonPath)
	if err == nil {
		t.Errorf("ImportTasks() expected error for invalid JSON, got nil")
	}
}

func TestRestoreBackup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zan_test_restore_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	// Create an initial app with some tasks
	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	app.AddTask("Original Task 1", "Desc 1", task.PriorityHigh, nil)
	app.AddTask("Original Task 2", "Desc 2", task.PriorityMedium, nil)
	initialTaskCount := len(app.GetAllTasks())

	// Create a dummy backup file with different tasks and settings
	backupFilePath := filepath.Join(tmpDir, "backup_tasks.json")
	backupTasksData := &task.Tasks{
		Version:   "1.1.0",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
		Tasks: []task.Task{
			{
				ID:          "backup-task-id-1",
				Title:       "Backup Task A",
				Description: "From backup file",
				Status:      task.StatusTODO,
				Priority:    task.PriorityLow,
				CreatedAt:   time.Now().Add(-48 * time.Hour),
				UpdatedAt:   time.Now().Add(-48 * time.Hour),
			},
			{
				ID:          "backup-task-id-2",
				Title:       "Backup Task B",
				Description: "Another backup task",
				Status:      task.StatusDone,
				Priority:    task.PriorityHigh,
				CreatedAt:   time.Now().Add(-47 * time.Hour),
				UpdatedAt:   time.Now().Add(-47 * time.Hour),
			},
		},
		Settings: task.Settings{
			DefaultPriority: task.PriorityHigh,
			AutoSave:        false,
			Theme:           "dark",
		},
	}

	data, err := json.MarshalIndent(backupTasksData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal backup tasks: %v", err)
	}
	err = os.WriteFile(backupFilePath, data, 0600)
	if err != nil {
		t.Fatalf("Failed to write dummy backup file: %v", err)
	}

	// Test successful restore
	err = app.RestoreBackup(backupFilePath)
	if err != nil {
		t.Fatalf("RestoreBackup() failed: %v", err)
	}

	// Verify tasks and settings after restore
	restoredTasks := app.GetAllTasks()
	if len(restoredTasks) != 2 { // Should be replaced by backup tasks
		t.Errorf("Expected 2 tasks after restore, got %d", len(restoredTasks))
	}
	if restoredTasks[0].Title != "Backup Task A" || restoredTasks[1].Title != "Backup Task B" {
		t.Errorf("Restored tasks content mismatch")
	}
	if app.Tasks.Version != "1.1.0" {
		t.Errorf("Restored version mismatch: expected 1.1.0, got %s", app.Tasks.Version)
	}
	if app.Tasks.Settings.DefaultPriority != task.PriorityHigh {
		t.Errorf("Restored default priority mismatch: expected %s, got %s", task.PriorityHigh, app.Tasks.Settings.DefaultPriority)
	}
	if app.Tasks.Settings.AutoSave != false {
		t.Errorf("Restored auto save mismatch: expected false, got %t", app.Tasks.Settings.AutoSave)
	}
	if app.Tasks.Settings.Theme != "dark" {
		t.Errorf("Restored theme mismatch: expected dark, got %s", app.Tasks.Settings.Theme)
	}

	// Test restore with empty file path
	err = app.RestoreBackup("")
	if err == nil {
		t.Errorf("RestoreBackup() expected error for empty file path, got nil")
	}

	// Test restore with non-existent file
	err = app.RestoreBackup(filepath.Join(tmpDir, "non_existent_backup.json"))
	if err == nil {
		t.Errorf("RestoreBackup() expected error for non-existent file, got nil")
	}

	// Test restore with invalid JSON
	invalidJsonPath := filepath.Join(tmpDir, "invalid_backup.json")
	err = os.WriteFile(invalidJsonPath, []byte("{invalid json"), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid json file: %v", err)
	}
	err = app.RestoreBackup(invalidJsonPath)
	if err == nil {
		t.Errorf("RestoreBackup() expected error for invalid JSON, got nil")
	}
	_ = initialTaskCount // Suppress unused variable warning
}
