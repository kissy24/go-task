package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

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

func TestAddTask(t *testing.T) {
	m := initialModel()

	// Go to add view
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = updatedModel.(model)
	if m.currentView != "add" {
		t.Fatalf("Expected view to be 'add', got %s", m.currentView)
	}

	// Type title
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("New Task Title")})
	m = updatedModel.(model)
	if m.titleInput.Value() != "New Task Title" {
		t.Errorf("Expected title input to be 'New Task Title', got %s", m.titleInput.Value())
	}

	// Move to next field (description)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focusIndex != 1 {
		t.Errorf("Expected focusIndex to be 1, got %d", m.focusIndex)
	}

	// Type description
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Task Description")})
	m = updatedModel.(model)
	if m.descriptionInput.Value() != "Task Description" {
		t.Errorf("Expected description input to be 'Task Description', got %s", m.descriptionInput.Value())
	}

	// Move to next field (priority)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focusIndex != 2 {
		t.Errorf("Expected focusIndex to be 2, got %d", m.focusIndex)
	}

	// Type priority
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("high")})
	m = updatedModel.(model)
	if m.priorityInput.Value() != "high" {
		t.Errorf("Expected priority input to be 'high', got %s", m.priorityInput.Value())
	}

	// Move to next field (tags)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedModel.(model)
	if m.focusIndex != 3 {
		t.Errorf("Expected focusIndex to be 3, got %d", m.focusIndex)
	}

	// Type tags
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("dev,urgent")})
	m = updatedModel.(model)
	if m.tagsInput.Value() != "dev,urgent" {
		t.Errorf("Expected tags input to be 'dev,urgent', got %s", m.tagsInput.Value())
	}

	// Submit form
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)

	if m.currentView != "main" {
		t.Errorf("Expected view to be 'main' after submission, got %s", m.currentView)
	}
	if len(m.tasks) != 1 {
		t.Errorf("Expected 1 task after submission, got %d", len(m.tasks))
	}
	if m.tasks[0].Title != "New Task Title" {
		t.Errorf("Expected new task title to be 'New Task Title', got %s", m.tasks[0].Title)
	}
	if m.tasks[0].Description != "Task Description" {
		t.Errorf("Expected new task description to be 'Task Description', got %s", m.tasks[0].Description)
	}
	if m.tasks[0].Priority != task.PriorityHigh {
		t.Errorf("Expected new task priority to be 'HIGH', got %s", m.tasks[0].Priority)
	}
	if !strings.Contains(m.tasks[0].Tags[0], "dev") || !strings.Contains(m.tasks[0].Tags[1], "urgent") {
		t.Errorf("Expected new task tags to be 'dev,urgent', got %v", m.tasks[0].Tags)
	}

	// Test canceling add task
	m = initialModel()
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = updatedModel.(model)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updatedModel.(model)
	if m.currentView != "main" {
		t.Errorf("Expected view to be 'main' after cancel, got %s", m.currentView)
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
