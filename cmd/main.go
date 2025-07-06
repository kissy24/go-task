package main

import (
	"fmt"
	"os"
	"zan/internal/app"
	"zan/internal/task"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	app         *app.App
	tasks       []task.Task
	cursor      int
	selected    map[string]struct{} // selected task IDs
	currentView string
	err         error
}

func initialModel() model {
	a, err := app.NewApp()
	if err != nil {
		return model{err: err}
	}

	return model{
		app:         a,
		tasks:       a.GetAllTasks(),
		selected:    make(map[string]struct{}),
		currentView: "main", // "main", "add", "edit", "detail"
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil` for now, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		// エラーが発生している場合は、qで終了のみ
		if msg, ok := msg.(tea.KeyMsg); ok && (msg.String() == "q" || msg.String() == "ctrl+c") {
			return m, tea.Quit
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}

		case "enter", " ":
			if len(m.tasks) > 0 {
				taskID := m.tasks[m.cursor].ID
				_, ok := m.selected[taskID]
				if ok {
					delete(m.selected, taskID)
				} else {
					m.selected[taskID] = struct{}{}
				}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress q to quit.", m.err)
	}

	s := "GoTask CLI v1.0.0\n\n"

	if len(m.tasks) == 0 {
		s += "No tasks found. Press 'a' to add a new task.\n\n"
	} else {
		for i, t := range m.tasks {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			checked := " "
			if _, ok := m.selected[t.ID]; ok {
				checked = "x"
			}

			s += fmt.Sprintf("%s [%s] %s [%s]\n", cursor, checked, t.Title, t.Priority)
		}
	}

	total, completed, incomplete := m.app.GetTaskStats()
	s += fmt.Sprintf("\nTotal: %d | Incomplete: %d | Completed: %d\n\n", total, incomplete, completed)
	s += "[a]dd [e]dit [d]elete [v]iew [f]ilter [q]uit [h]elp\n"

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
