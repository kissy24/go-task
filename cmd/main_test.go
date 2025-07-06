package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialModel(t *testing.T) {
	m := initialModel()
	if len(m.choices) != 3 {
		t.Errorf("Expected 3 choices, got %d", len(m.choices))
	}
	if m.cursor != 0 {
		t.Errorf("Expected cursor to be 0, got %d", m.cursor)
	}
	if len(m.selected) != 0 {
		t.Errorf("Expected no selected items, got %d", len(m.selected))
	}
}

func TestUpdate(t *testing.T) {
	m := initialModel()

	// Test moving cursor down
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(model)
	if m.cursor != 1 {
		t.Errorf("Expected cursor to be 1, got %d", m.cursor)
	}

	// Test selecting an item
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(model)
	if _, ok := m.selected[1]; !ok {
		t.Errorf("Expected item at cursor 1 to be selected")
	}

	// Test quitting
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Errorf("Expected a quit command")
	}
}

func TestView(t *testing.T) {
	m := initialModel()
	view := m.View()
	expected := `What should we buy at the market?

> [ ] Buy carrots
  [ ] Buy celery
  [ ] Buy kohlrabi

Press q to quit.
`
	if view != expected {
		t.Errorf("Expected view:\n%s\nGot:\n%s", expected, view)
	}
}
