package main

import (
	"fmt"
	"strings"
	"testing"

	"os" // osパッケージを追加
	"zan/internal/app"
	"zan/internal/store"
	"zan/internal/task"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMain(m *testing.M) {
	// テスト前にtasks.jsonを削除
	configDir, _ := store.GetConfigDirPath()
	os.RemoveAll(configDir) // ディレクトリごと削除

	// テスト実行
	code := m.Run()

	// テスト後にディレクトリごと削除
	os.RemoveAll(configDir)
	os.Exit(code)
}

func TestInitialModel(t *testing.T) {
	m := initialModel()
	if m.err != nil {
		t.Fatalf("initialModel returned an error: %v", m.err)
	}
	if len(m.tasks) != 0 { // initialModel should load no tasks if tasks.json is empty
		t.Errorf("Expected 0 tasks, got %d", len(m.tasks))
	}
	if m.cursor != 0 {
		t.Errorf("Expected cursor to be 0, got %d", m.cursor)
	}
	if len(m.selected) != 0 {
		t.Errorf("Expected no selected items, got %d", len(m.selected))
	}
	if m.currentView != "main" {
		t.Errorf("Expected currentView to be 'main', got %s", m.currentView)
	}
	if m.app == nil {
		t.Errorf("Expected app to be initialized")
	}
}

func TestUpdate(t *testing.T) {
	// Mock app for testing
	mockApp, _ := app.NewApp()
	mockApp.Tasks.Tasks = []task.Task{
		{ID: "test-1", Title: "Test Task 1", Priority: task.PriorityMedium, Status: task.StatusTODO},
		{ID: "test-2", Title: "Test Task 2", Priority: task.PriorityLow, Status: task.StatusTODO},
	}

	m := model{
		app:         mockApp,
		tasks:       mockApp.GetAllTasks(),
		selected:    make(map[string]struct{}),
		currentView: "main",
	}

	// Test moving cursor down
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(model)
	if m.cursor != 1 {
		t.Errorf("Expected cursor to be 1, got %d", m.cursor)
	}

	// Test selecting an item
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if _, ok := m.selected[m.tasks[1].ID]; !ok {
		t.Errorf("Expected item at cursor 1 to be selected")
	}

	// Test quitting
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Errorf("Expected a quit command")
	}
}

func TestView(t *testing.T) {
	// Mock app for testing
	mockApp, _ := app.NewApp()
	mockApp.Tasks.Tasks = []task.Task{
		{ID: "test-a", Title: "Test Task A", Priority: task.PriorityHigh, Status: task.StatusTODO},
		{ID: "test-b", Title: "Test Task B", Priority: task.PriorityMedium, Status: task.StatusTODO},
		{ID: "test-c", Title: "Test Task C", Priority: task.PriorityLow, Status: task.StatusTODO},
	}

	m := model{
		app:         mockApp,
		tasks:       mockApp.GetAllTasks(),
		selected:    make(map[string]struct{}),
		currentView: "main",
	}

	view := m.View()
	// Expected output will depend on generated UUIDs, so we'll check for key elements
	if !strings.Contains(view, "GoTask CLI v1.0.0") {
		t.Errorf("View missing header")
	}
	if !strings.Contains(view, "Test Task A [HIGH]") {
		t.Errorf("View missing Test Task A")
	}
	if !strings.Contains(view, "Test Task B [MEDIUM]") {
		t.Errorf("View missing Test Task B")
	}
	if !strings.Contains(view, "Test Task C [LOW]") {
		t.Errorf("View missing Test Task C")
	}
	expectedStats := fmt.Sprintf("Total: %d | Incomplete: %d | Completed: %d", 3, 3, 0)
	if !strings.Contains(view, expectedStats) {
		t.Errorf("View missing stats. Expected: %s, Got: %s", expectedStats, view)
	}
	if !strings.Contains(view, "[a]dd [e]dit [d]elete [v]iew [f]ilter [q]uit [h]elp") {
		t.Errorf("View missing footer menu")
	}
}
