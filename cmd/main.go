package main

import (
	"fmt"
	"os"
	"strings"

	"zan/internal/app"
	"zan/internal/task"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	app         *app.App
	tasks       []task.Task
	cursor      int
	selected    map[string]struct{} // selected task IDs
	currentView string
	err         error

	// Add task form fields
	titleInput       textinput.Model
	descriptionInput textinput.Model
	priorityInput    textinput.Model
	tagsInput        textinput.Model
	focusIndex       int // Which input field is focused
}

func initialModel() model {
	a, err := app.NewApp()
	if err != nil {
		return model{err: err}
	}

	ti := textinput.New()
	ti.Placeholder = "Task Title"
	ti.Focus()
	ti.CharLimit = 255
	ti.Width = 50

	di := textinput.New()
	di.Placeholder = "Task Description"
	di.CharLimit = 1000
	di.Width = 50

	pi := textinput.New()
	pi.Placeholder = "High, Medium, Low (default: Medium)"
	pi.CharLimit = 10
	pi.Width = 20

	tai := textinput.New()
	tai.Placeholder = "tag1,tag2,tag3"
	tai.CharLimit = 100
	tai.Width = 50

	return model{
		app:              a,
		tasks:            a.GetAllTasks(),
		selected:         make(map[string]struct{}),
		currentView:      "main", // "main", "add", "edit", "detail"
		titleInput:       ti,
		descriptionInput: di,
		priorityInput:    pi,
		tagsInput:        tai,
		focusIndex:       0,
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

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "a": // Add task
			if m.currentView == "main" {
				m.currentView = "add"
				m.titleInput.Focus()
				m.focusIndex = 0
				return m, nil
			}

		case "e": // Edit task
			if m.currentView == "main" && len(m.selected) == 1 {
				for id := range m.selected {
					t, err := m.app.GetTaskByID(id)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.titleInput.SetValue(t.Title)
					m.descriptionInput.SetValue(t.Description)
					m.priorityInput.SetValue(string(t.Priority))
					m.tagsInput.SetValue(strings.Join(t.Tags, ","))
					m.currentView = "edit"
					m.focusIndex = 0
					m.titleInput.Focus()
					return m, nil
				}
			}

		case "esc":
			if m.currentView == "add" || m.currentView == "edit" {
				m.currentView = "main"
				// Clear form fields
				m.titleInput.SetValue("")
				m.descriptionInput.SetValue("")
				m.priorityInput.SetValue("")
				m.tagsInput.SetValue("")
				m.titleInput.Blur()
				m.descriptionInput.Blur()
				m.priorityInput.Blur()
				m.tagsInput.Blur()
				m.selected = make(map[string]struct{}) // Clear selection
				return m, nil
			}

		case "up", "shift+tab":
			if m.currentView == "add" {
				m.focusIndex--
				// Wrap around
				if m.focusIndex < 0 {
					m.focusIndex = 3
				}
				cmds = append(cmds, m.setFocus())
			} else if m.currentView == "main" {
				if m.cursor > 0 {
					m.cursor--
				}
			}

		case "down", "tab":
			if m.currentView == "add" {
				m.focusIndex++
				// Wrap around
				if m.focusIndex > 3 {
					m.focusIndex = 0
				}
				cmds = append(cmds, m.setFocus())
			} else if m.currentView == "main" {
				if m.cursor < len(m.tasks)-1 {
					m.cursor++
				}
			}

		case "enter":
			if m.currentView == "add" {
				title := m.titleInput.Value()
				description := m.descriptionInput.Value()
				priority := task.Priority(strings.ToUpper(m.priorityInput.Value()))
				tagsStr := m.tagsInput.Value()
				var tags []string
				if tagsStr != "" {
					tags = strings.Split(tagsStr, ",")
				}

				_, err := m.app.AddTask(title, description, priority, tags)
				if err != nil {
					m.err = err
				} else {
					m.currentView = "main"
					m.tasks = m.app.GetAllTasks() // Refresh tasks
					// Clear form fields
					m.titleInput.SetValue("")
					m.descriptionInput.SetValue("")
					m.priorityInput.SetValue("")
					m.tagsInput.SetValue("")
					m.titleInput.Blur()
					m.descriptionInput.Blur()
					m.priorityInput.Blur()
					m.tagsInput.Blur()
				}
				return m, tea.Batch(cmds...)
			} else if m.currentView == "edit" {
				if len(m.selected) != 1 {
					m.err = fmt.Errorf("Please select exactly one task to edit")
					return m, nil
				}
				var taskID string
				for id := range m.selected {
					taskID = id
				}

				title := m.titleInput.Value()
				description := m.descriptionInput.Value()
				priority := task.Priority(strings.ToUpper(m.priorityInput.Value()))
				tagsStr := m.tagsInput.Value()
				var tags []string
				if tagsStr != "" {
					tags = strings.Split(tagsStr, ",")
				}

				_, err := m.app.UpdateTask(taskID, title, description, "", priority, tags) // Status is not edited here
				if err != nil {
					m.err = err
				} else {
					m.currentView = "main"
					m.tasks = m.app.GetAllTasks()          // Refresh tasks
					m.selected = make(map[string]struct{}) // Clear selection
					// Clear form fields
					m.titleInput.SetValue("")
					m.descriptionInput.SetValue("")
					m.priorityInput.SetValue("")
					m.tagsInput.SetValue("")
					m.titleInput.Blur()
					m.descriptionInput.Blur()
					m.priorityInput.Blur()
					m.tagsInput.Blur()
				}
				return m, tea.Batch(cmds...)
			} else if m.currentView == "main" {
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
		case " ": // Spacebar for main view selection
			if m.currentView == "main" {
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
	}

	// Handle text input updates
	if m.currentView == "add" || m.currentView == "edit" {
		switch m.focusIndex {
		case 0:
			m.titleInput, cmd = m.titleInput.Update(msg)
		case 1:
			m.descriptionInput, cmd = m.descriptionInput.Update(msg)
		case 2:
			m.priorityInput, cmd = m.priorityInput.Update(msg)
		case 3:
			m.tagsInput, cmd = m.tagsInput.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	return m, tea.Batch(cmds...)
}

func (m *model) setFocus() tea.Cmd {
	cmds := make([]tea.Cmd, 4)
	inputs := []*textinput.Model{&m.titleInput, &m.descriptionInput, &m.priorityInput, &m.tagsInput}
	for i := 0; i <= len(inputs)-1; i++ {
		if i == m.focusIndex {
			// Set focused state
			cmds[i] = inputs[i].Focus()
			inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")) // lipglossのスタイル設定を削除
			inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))   // lipglossのスタイル設定を削除
		} else {
			// Remove focused state
			inputs[i].Blur()
			inputs[i].PromptStyle = lipgloss.NewStyle() // lipglossのスタイル設定を削除
			inputs[i].TextStyle = lipgloss.NewStyle()   // lipglossのスタイル設定を削除
		}
	}
	return tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress q to quit.", m.err)
	}

	switch m.currentView {
	case "main":
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

	case "add":
		return fmt.Sprintf(
			"Add New Task\n\n%s\n%s\n%s\n%s\n\n%s",
			m.titleInput.View(),
			m.descriptionInput.View(),
			m.priorityInput.View(),
			m.tagsInput.View(),
			"[enter] to submit, [esc] to cancel",
		)
	case "edit":
		return fmt.Sprintf(
			"Edit Task\n\n%s\n%s\n%s\n%s\n\n%s",
			m.titleInput.View(),
			m.descriptionInput.View(),
			m.priorityInput.View(),
			m.tagsInput.View(),
			"[enter] to save, [esc] to cancel",
		)
	}

	return "Unknown view"
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
