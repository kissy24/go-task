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

var (
	// 優先度に応じた色
	priorityColors = map[task.Priority]lipgloss.Color{
		task.PriorityHigh:   lipgloss.Color("9"),  // Red
		task.PriorityMedium: lipgloss.Color("11"), // Yellow
		task.PriorityLow:    lipgloss.Color("10"), // Green
	}

	// 状態に応じたアイコン
	statusIcons = map[task.Status]string{
		task.StatusTODO:       "●",
		task.StatusInProgress: "◐",
		task.StatusDone:       "✓",
		task.StatusPending:    "⏸",
	}
)

type model struct {
	app         *app.App
	tasks       []task.Task
	cursor      int
	selected    map[string]struct{} // selected task IDs
	currentView string
	err         error

	// Filter fields
	filterStatusInput textinput.Model
	filteredStatuses  map[task.Status]struct{}
	isFiltering       bool

	filterPriorityInput textinput.Model
	filteredPriorities  map[task.Priority]struct{}

	filterTagsInput textinput.Model
	filteredTags    map[string]struct{}

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

	fsi := textinput.New()
	fsi.Placeholder = "Filter by status (e.g., TODO,IN_PROGRESS)"
	fsi.CharLimit = 50
	fsi.Width = 50

	return model{
		app:               a,
		tasks:             a.GetAllTasks(),
		selected:          make(map[string]struct{}),
		currentView:       "main", // "main", "add", "edit", "detail", "filter"
		titleInput:        ti,
		descriptionInput:  di,
		priorityInput:     pi,
		tagsInput:         tai,
		focusIndex:        0,
		filterStatusInput: fsi,
		filteredStatuses:  make(map[task.Status]struct{}),
		isFiltering:       false,
	}

	fpi := textinput.New()
	fpi.Placeholder = "Filter by priority (e.g., HIGH,MEDIUM)"
	fpi.CharLimit = 50
	fpi.Width = 50

	fti := textinput.New()
	fti.Placeholder = "Filter by tags (e.g., work,personal)"
	fti.CharLimit = 100
	fti.Width = 50

	return model{
		app:                 a,
		tasks:               a.GetAllTasks(),
		selected:            make(map[string]struct{}),
		currentView:         "main", // "main", "add", "edit", "detail", "filter", "filter_priority", "filter_tags"
		titleInput:          ti,
		descriptionInput:    di,
		priorityInput:       pi,
		tagsInput:           tai,
		focusIndex:          0,
		filterStatusInput:   fsi,
		filteredStatuses:    make(map[task.Status]struct{}),
		isFiltering:         false,
		filterPriorityInput: fpi,
		filteredPriorities:  make(map[task.Priority]struct{}),
		filterTagsInput:     fti,
		filteredTags:        make(map[string]struct{}),
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
			if m.currentView == "main" && len(m.tasks) > 0 { // カーソルがタスクを指している場合
				t := m.tasks[m.cursor]
				m.titleInput.SetValue(t.Title)
				m.descriptionInput.SetValue(t.Description)
				m.priorityInput.SetValue(string(t.Priority))
				m.tagsInput.SetValue(strings.Join(t.Tags, ","))
				m.currentView = "edit"
				m.focusIndex = 0
				m.titleInput.Focus()
				return m, nil
			}

		case "f": // Filter tasks
			if m.currentView == "main" {
				m.currentView = "filter"
				m.filterStatusInput.Focus()
				return m, nil
			}

		case "p": // Filter by priority
			if m.currentView == "main" {
				m.currentView = "filter_priority"
				m.filterPriorityInput.Focus()
				return m, nil
			}

		case "t": // Filter by tags
			if m.currentView == "main" {
				m.currentView = "filter_tags"
				m.filterTagsInput.Focus()
				return m, nil
			}

		case "esc":
			if m.currentView == "add" || m.currentView == "edit" || m.currentView == "filter" || m.currentView == "filter_priority" || m.currentView == "filter_tags" {
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
				m.filterStatusInput.SetValue("") // Clear status filter input
				m.filterStatusInput.Blur()
				m.filteredStatuses = make(map[task.Status]struct{}) // Clear filtered statuses
				m.filterPriorityInput.SetValue("")                  // Clear priority filter input
				m.filterPriorityInput.Blur()
				m.filteredPriorities = make(map[task.Priority]struct{}) // Clear filtered priorities
				m.filterTagsInput.SetValue("")                          // Clear tags filter input
				m.filterTagsInput.Blur()
				m.filteredTags = make(map[string]struct{}) // Clear filtered tags
				m.isFiltering = false
				m.tasks = m.app.GetAllTasks()          // Reset tasks to all tasks
				m.selected = make(map[string]struct{}) // Clear selection
				return m, nil
			}

		case "up", "shift+tab":
			if m.currentView == "add" || m.currentView == "edit" { // editビューも追加
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
			if m.currentView == "add" || m.currentView == "edit" { // editビューも追加
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
				// if len(m.selected) != 1 { // 選択状態は不要になったためコメントアウト
				// 	m.err = fmt.Errorf("Please select exactly one task to edit")
				// 	return m, nil
				// }
				taskID := m.tasks[m.cursor].ID // カーソルが指しているタスクのIDを使用

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
					m.tasks = m.app.GetAllTasks() // Refresh tasks
					// m.selected = make(map[string]struct{}) // 選択状態をクリアしない
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
			} else if m.currentView == "filter" {
				statusStr := m.filterStatusInput.Value()
				if statusStr != "" {
					statuses := strings.Split(strings.ToUpper(statusStr), ",")
					m.filteredStatuses = make(map[task.Status]struct{})
					for _, s := range statuses {
						m.filteredStatuses[task.Status(strings.TrimSpace(s))] = struct{}{}
					}
					m.isFiltering = true
					m.tasks = m.app.GetFilteredTasksByStatus(m.convertStatusMapToList())
				} else {
					m.isFiltering = false
					m.tasks = m.app.GetAllTasks()
				}
				m.currentView = "main"
				m.filterStatusInput.SetValue("")
				m.filterStatusInput.Blur()
				return m, tea.Batch(cmds...)
			} else if m.currentView == "filter_priority" {
				priorityStr := m.filterPriorityInput.Value()
				if priorityStr != "" {
					priorities := strings.Split(strings.ToUpper(priorityStr), ",")
					m.filteredPriorities = make(map[task.Priority]struct{})
					for _, p := range priorities {
						m.filteredPriorities[task.Priority(strings.TrimSpace(p))] = struct{}{}
					}
					m.isFiltering = true
					m.tasks = m.app.GetFilteredTasksByPriority(m.convertPriorityMapToList())
				} else {
					m.isFiltering = false
					m.tasks = m.app.GetAllTasks()
				}
				m.currentView = "main"
				m.filterPriorityInput.SetValue("")
				m.filterPriorityInput.Blur()
				return m, tea.Batch(cmds...)
			} else if m.currentView == "filter_tags" {
				tagsStr := m.filterTagsInput.Value()
				if tagsStr != "" {
					tags := strings.Split(tagsStr, ",")
					m.filteredTags = make(map[string]struct{})
					for _, t := range tags {
						m.filteredTags[strings.TrimSpace(t)] = struct{}{}
					}
					m.isFiltering = true
					m.tasks = m.app.GetFilteredTasksByTags(m.convertTagMapToList())
				} else {
					m.isFiltering = false
					m.tasks = m.app.GetAllTasks()
				}
				m.currentView = "main"
				m.filterTagsInput.SetValue("")
				m.filterTagsInput.Blur()
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
	} else if m.currentView == "filter" {
		m.filterStatusInput, cmd = m.filterStatusInput.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.currentView == "filter_priority" {
		m.filterPriorityInput, cmd = m.filterPriorityInput.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.currentView == "filter_tags" {
		m.filterTagsInput, cmd = m.filterTagsInput.Update(msg)
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
			inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		} else {
			// Remove focused state
			inputs[i].Blur()
			inputs[i].PromptStyle = lipgloss.NewStyle()
			inputs[i].TextStyle = lipgloss.NewStyle()
		}
	}
	return tea.Batch(cmds...)
}

// convertStatusMapToList はmap[task.Status]struct{}を[]task.Statusに変換します。
func (m model) convertStatusMapToList() []task.Status {
	var statuses []task.Status
	for s := range m.filteredStatuses {
		statuses = append(statuses, s)
	}
	return statuses
}

// convertPriorityMapToList はmap[task.Priority]struct{}を[]task.Priorityに変換します。
func (m model) convertPriorityMapToList() []task.Priority {
	var priorities []task.Priority
	for p := range m.filteredPriorities {
		priorities = append(priorities, p)
	}
	return priorities
}

// convertTagMapToList はmap[string]struct{}を[]stringに変換します。
func (m model) convertTagMapToList() []string {
	var tags []string
	for t := range m.filteredTags {
		tags = append(tags, t)
	}
	return tags
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress q to quit.", m.err)
	}

	switch m.currentView {
	case "main":
		s := "GoTask CLI v1.0.0\n\n"

		if m.isFiltering {
			filterConditions := []string{}
			for status := range m.filteredStatuses {
				filterConditions = append(filterConditions, string(status))
			}
			s += fmt.Sprintf("Filtered by Status: %s\n\n", strings.Join(filterConditions, ", "))
		}

		if len(m.tasks) == 0 {
			s += "No tasks found. Press 'a' to add a new task.\n\n"
		} else {
			for i, t := range m.tasks {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}

				// checked := " " // 未使用のためコメントアウト
				if _, ok := m.selected[t.ID]; ok {
					// checked = "x" // 未使用のためコメントアウト
				}

				// 色とアイコンを適用
				statusIcon := statusIcons[t.Status]
				priorityColor := priorityColors[t.Priority]
				styledTitle := lipgloss.NewStyle().Foreground(priorityColor).Render(t.Title)

				s += fmt.Sprintf("%s %s %s %s\n", cursor, statusIcon, styledTitle, lipgloss.NewStyle().Foreground(priorityColor).Render(string(t.Priority)))
			}
		}

		total, completed, incomplete := m.app.GetTaskStats()
		s += fmt.Sprintf("\nTotal: %d | Incomplete: %d | Completed: %d\n\n", total, incomplete, completed)
		s += "[a]dd [e]dit [d]elete [v]iew [f]ilter [p]riority filter [t]ag filter [q]uit [h]elp\n"
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
	case "filter":
		return fmt.Sprintf(
			"Filter Tasks by Status (e.g., TODO,IN_PROGRESS)\n\n%s\n\n%s",
			m.filterStatusInput.View(),
			"[enter] to apply filter, [esc] to cancel",
		)
	case "filter_priority":
		return fmt.Sprintf(
			"Filter Tasks by Priority (e.g., HIGH,MEDIUM)\n\n%s\n\n%s",
			m.filterPriorityInput.View(),
			"[enter] to apply filter, [esc] to cancel",
		)
	case "filter_tags":
		allTags := m.app.GetAllUniqueTags()
		tagsList := ""
		if len(allTags) > 0 {
			tagsList = fmt.Sprintf("Available Tags: %s\n\n", strings.Join(allTags, ", "))
		} else {
			tagsList = "No tags available.\n\n"
		}
		return fmt.Sprintf(
			"Filter Tasks by Tags (e.g., work,personal)\n\n%s%s\n\n%s",
			tagsList,
			m.filterTagsInput.View(),
			"[enter] to apply filter, [esc] to cancel",
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
