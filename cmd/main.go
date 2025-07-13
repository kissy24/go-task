package main

import (
	"fmt"
	"os"
	"strings"

	"go-task/internal/app"
	"go-task/internal/config"
	"go-task/internal/log"
	"go-task/internal/task"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/profile"
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
	app            *app.App
	tasks          []task.Task
	cursor         int
	selected       map[string]struct{} // selected task IDs
	currentView    string
	err            *app.AppError // Use custom error type
	detailViewTask *task.Task     // Currently viewed task in detail view
	cfg            *config.Config // Application configuration

	// Filter fields
	filterStatusInput textinput.Model
	filteredStatuses  map[task.Status]struct{}
	isFiltering       bool

	filterPriorityInput textinput.Model
	filteredPriorities  map[task.Priority]struct{}

	filterTagsInput textinput.Model
	filteredTags    map[string]struct{}

	searchInput   textinput.Model // Search input field
	searchKeyword string          // Current search keyword

	sortInput textinput.Model // Sort input field
	sortBy    string          // Current sort by field
	sortAsc   bool            // Current sort order (ascending/descending)

	// Add task form fields
	titleInput       textinput.Model
	descriptionInput textinput.Model
	priorityInput    textinput.Model
	tagsInput        textinput.Model
	focusIndex       int // Which input field is focused

	// Settings form fields
	defaultPriorityInput textinput.Model
	autoSaveInput        textinput.Model
	themeInput           textinput.Model
	settingsFocusIndex   int // Which input field is focused in settings view

	// Export form fields
	exportInput textinput.Model

	// Import form fields
	importInput textinput.Model
}

func initialModel() model {
	a, err := app.NewApp()
	if err != nil {
		appErr, _ := err.(*app.AppError)
		return model{err: appErr}
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return model{err: app.NewAppError(app.ErrTypeInternal, "Failed to load config", err)}
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

	fpi := textinput.New()
	fpi.Placeholder = "Filter by priority (e.g., HIGH,MEDIUM)"
	fpi.CharLimit = 50
	fpi.Width = 50

	fti := textinput.New()
	fti.Placeholder = "Filter by tags (e.g., work,personal)"
	fti.CharLimit = 100
	fti.Width = 50

	si := textinput.New()
	si.Placeholder = "Search keyword (title or description)"
	si.CharLimit = 255
	si.Width = 50

	sortInput := textinput.New()
	sortInput.Placeholder = "Sort by: created_at, updated_at, priority (e.g., created_at asc)"
	sortInput.CharLimit = 50
	sortInput.Width = 50

	// Settings input fields
	dpi := textinput.New()
	dpi.Placeholder = fmt.Sprintf("Default Priority (e.g., %s, %s, %s)", task.PriorityHigh, task.PriorityMedium, task.PriorityLow)
	dpi.CharLimit = 10
	dpi.Width = 50

	asi := textinput.New()
	asi.Placeholder = "Auto Save (true/false)"
	asi.CharLimit = 5
	asi.Width = 50

	themeInput := textinput.New()
	themeInput.Placeholder = "Theme (e.g., default, dark)"
	themeInput.CharLimit = 20
	themeInput.Width = 50

	return model{
		app:                  a,
		tasks:                a.GetAllTasks(),
		selected:             make(map[string]struct{}),
		currentView:          "main", // "main", "add", "edit", "detail", "filter", "filter_priority", "filter_tags", "search", "sort"
		titleInput:           ti,
		descriptionInput:     di,
		priorityInput:        pi,
		tagsInput:            tai,
		focusIndex:           0,
		filterStatusInput:    fsi,
		filteredStatuses:     make(map[task.Status]struct{}),
		isFiltering:          false,
		filterPriorityInput:  fpi,
		filteredPriorities:   make(map[task.Priority]struct{}),
		filterTagsInput:      fti,
		filteredTags:         make(map[string]struct{}),
		searchInput:          si,
		sortInput:            sortInput,
		sortBy:               "created_at", // Default sort by created_at
		sortAsc:              false,        // Default sort descending
		cfg:                  cfg,
		defaultPriorityInput: dpi,
		autoSaveInput:        asi,
		themeInput:           themeInput,
		exportInput:          textinput.New(),
		importInput:          textinput.New(),
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

		case "i": // Import tasks
			if m.currentView == "main" {
				m.currentView = "import"
				m.importInput.Focus()
				return m, nil
			}

		case "a": // Add task
			if m.currentView == "main" {
				m.currentView = "add"
				m.titleInput.Focus()
				m.focusIndex = 0
				return m, nil
			}

		case "g": // Go to settings
			if m.currentView == "main" {
				m.currentView = "settings"
				m.defaultPriorityInput.SetValue(string(m.cfg.Settings.DefaultPriority))
				m.autoSaveInput.SetValue(fmt.Sprintf("%t", m.cfg.Settings.AutoSave))
				m.themeInput.SetValue(m.cfg.Settings.Theme)
				m.settingsFocusIndex = 0
				m.defaultPriorityInput.Focus()
				return m, nil
			}

		case "o": // Sort tasks
			if m.currentView == "main" {
				m.currentView = "sort"
				m.sortInput.Focus()
				return m, nil
			}

		case "v": // View task details
			if m.currentView == "main" && len(m.tasks) > 0 {
				m.detailViewTask = &m.tasks[m.cursor]
				m.currentView = "detail"
				return m, nil
			}

		case "c": // Change task status (cycle through TODO, IN_PROGRESS, DONE, PENDING)
			if m.currentView == "main" && len(m.tasks) > 0 {
				taskID := m.tasks[m.cursor].ID
				currentTask, err := m.app.GetTaskByID(taskID)
				if err != nil {
					m.err, _ = err.(*app.AppError)
					return m, nil
				}

				nextStatus := currentTask.Status
				switch currentTask.Status {
				case task.StatusTODO:
					nextStatus = task.StatusInProgress
				case task.StatusInProgress:
					nextStatus = task.StatusDone
				case task.StatusDone:
					nextStatus = task.StatusPending
				case task.StatusPending:
					nextStatus = task.StatusTODO
				}

				_, err = m.app.UpdateTask(taskID, "", "", nextStatus, "", nil)
				if err != nil {
					m.err, _ = err.(*app.AppError)
				} else {
					m.tasks = m.app.GetAllTasks() // Refresh tasks
				}
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

		case "s": // Search tasks
			if m.currentView == "main" {
				m.currentView = "search"
				m.searchInput.Focus()
				return m, nil
			}

		case "esc":
			if m.currentView == "add" || m.currentView == "edit" || m.currentView == "filter" || m.currentView == "filter_priority" || m.currentView == "filter_tags" || m.currentView == "search" || m.currentView == "sort" || m.currentView == "detail" || m.currentView == "settings" || m.currentView == "export" || m.currentView == "import" {
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
				m.searchInput.SetValue("")                 // Clear search input
				m.searchInput.Blur()
				m.searchKeyword = ""
				m.sortInput.SetValue("") // Clear sort input
				m.sortInput.Blur()
				m.sortBy = "created_at" // Reset sort by
				m.sortAsc = false       // Reset sort order
				m.isFiltering = false
				m.detailViewTask = nil                 // Clear selected task for detail view
				m.tasks = m.app.GetAllTasks()          // Reset tasks to all tasks
				m.selected = make(map[string]struct{}) // Clear selection

				// Clear settings form fields
				m.defaultPriorityInput.SetValue("")
				m.autoSaveInput.SetValue("")
				m.themeInput.SetValue("")
				m.defaultPriorityInput.Blur()
				m.autoSaveInput.Blur()
				m.themeInput.Blur()
				return m, nil
			}

		case "up", "shift+tab":
			if m.currentView == "add" || m.currentView == "edit" {
				m.focusIndex--
				if m.focusIndex < 0 {
					m.focusIndex = 3
				}
				cmds = append(cmds, m.setFocus())
			} else if m.currentView == "settings" {
				m.settingsFocusIndex--
				if m.settingsFocusIndex < 0 {
					m.settingsFocusIndex = 2
				}
				cmds = append(cmds, m.setSettingsFocus())
			} else if m.currentView == "main" {
				if m.cursor > 0 {
					m.cursor--
				}
			}

		case "down", "tab":
			if m.currentView == "add" || m.currentView == "edit" {
				m.focusIndex++
				if m.focusIndex > 3 {
					m.focusIndex = 0
				}
				cmds = append(cmds, m.setFocus())
			} else if m.currentView == "settings" {
				m.settingsFocusIndex++
				if m.settingsFocusIndex > 2 {
					m.settingsFocusIndex = 0
				}
				cmds = append(cmds, m.setSettingsFocus())
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
					m.err, _ = err.(*app.AppError)
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
					m.err, _ = err.(*app.AppError)
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
			} else if m.currentView == "search" {
				m.searchKeyword = m.searchInput.Value()
				if m.searchKeyword != "" {
					m.isFiltering = true
					m.tasks = m.app.Search(m.searchKeyword)
				} else {
					m.isFiltering = false
					m.tasks = m.app.GetAllTasks()
				}
				m.currentView = "main"
				m.searchInput.SetValue("")
				m.searchInput.Blur()
				return m, tea.Batch(cmds...)
			} else if m.currentView == "sort" {
				sortStr := m.sortInput.Value()
				if sortStr != "" {
					parts := strings.Fields(sortStr)
					if len(parts) > 0 {
						m.sortBy = parts[0]
						if len(parts) > 1 && strings.ToLower(parts[1]) == "asc" {
							m.sortAsc = true
						} else {
							m.sortAsc = false
						}
					}
					m.tasks = m.app.SortTasks(m.tasks, m.sortBy, m.sortAsc)
				} else {
					// Clear sort if input is empty
					m.sortBy = "created_at"
					m.sortAsc = false
					m.tasks = m.app.SortTasks(m.app.GetAllTasks(), m.sortBy, m.sortAsc) // Reset to default sort
				}
				m.currentView = "main"
				m.sortInput.SetValue("")
				m.sortInput.Blur()
				return m, tea.Batch(cmds...)
			} else if m.currentView == "settings" {
				// Save settings
				defaultPriority := task.Priority(strings.ToUpper(m.defaultPriorityInput.Value()))
				autoSaveStr := strings.ToLower(m.autoSaveInput.Value())
				theme := m.themeInput.Value()

				autoSave := false
				if autoSaveStr == "true" {
					autoSave = true
				}

				m.cfg.Settings.DefaultPriority = defaultPriority
				m.cfg.Settings.AutoSave = autoSave
				m.cfg.Settings.Theme = theme

				err := m.cfg.SaveConfig()
				if err != nil {
					m.err = app.NewAppError(app.ErrTypeIO, "Failed to save config", err)
				} else {
					m.currentView = "main"
					// Clear settings form fields
					m.defaultPriorityInput.SetValue("")
					m.autoSaveInput.SetValue("")
					m.themeInput.SetValue("")
					m.defaultPriorityInput.Blur()
					m.autoSaveInput.Blur()
					m.themeInput.Blur()
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
		case " ": // Space bar for main view selection
			if m.currentView == "main" {
				if len(m.tasks) > 0 {
					taskID := m.tasks[m.cursor].ID
					_, ok := m.selected[taskID]
					if ok {
						delete(m.selected, taskID)
					} else {
						m.selected[taskID] = struct{}{}
					}
				} else if m.currentView == "settings" {
					switch m.settingsFocusIndex {
					case 0:
						m.defaultPriorityInput, cmd = m.defaultPriorityInput.Update(msg)
					case 1:
						m.autoSaveInput, cmd = m.autoSaveInput.Update(msg)
					case 2:
						m.themeInput, cmd = m.themeInput.Update(msg)
					}
					cmds = append(cmds, cmd)
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
	} else if m.currentView == "search" {
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.currentView == "sort" {
		m.sortInput, cmd = m.sortInput.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.currentView == "import" {
		m.importInput, cmd = m.importInput.Update(msg)
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

func (m *model) setSettingsFocus() tea.Cmd {
	cmds := make([]tea.Cmd, 3)
	inputs := []*textinput.Model{&m.defaultPriorityInput, &m.autoSaveInput, &m.themeInput}
	for i := 0; i <= len(inputs)-1; i++ {
		if i == m.settingsFocusIndex {
			cmds[i] = inputs[i].Focus()
			inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		} else {
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
		var b strings.Builder
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9")).Render("An error occurred"))
		b.WriteString(fmt.Sprintf("\n\nType: %s\n", m.err.Type))
		b.WriteString(fmt.Sprintf("Message: %s\n\n", m.err.Message))

		switch m.err.Type {
		case app.ErrTypeIO:
			b.WriteString("Suggestion: Please check file permissions and ensure the file exists at the expected location.")
		case app.ErrTypeValidation:
			b.WriteString("Suggestion: Please check your input and try again. Ensure all required fields are filled correctly.")
		case app.ErrTypeNotFound:
			b.WriteString("Suggestion: The requested item could not be found. Please check the ID and try again.")
		case app.ErrTypeInternal:
			b.WriteString("Suggestion: An unexpected internal error occurred. Please check the logs for more details.")
		}

		b.WriteString("\n\nPress 'q' to quit.")
		return b.String()
	}



	switch m.currentView {
	case "main":
		s := "GoTask CLI v1.0.0\n\n"

		if m.isFiltering {
			filterConditions := []string{}
			if len(m.filteredStatuses) > 0 {
				statuses := []string{}
				for status := range m.filteredStatuses {
					statuses = append(statuses, string(status))
				}
				filterConditions = append(filterConditions, fmt.Sprintf("Status: %s", strings.Join(statuses, ", ")))
			}
			if len(m.filteredPriorities) > 0 {
				priorities := []string{}
				for priority := range m.filteredPriorities {
					priorities = append(priorities, string(priority))
				}
				filterConditions = append(filterConditions, fmt.Sprintf("Priority: %s", strings.Join(priorities, ", ")))
			}
			if len(m.filteredTags) > 0 {
				tags := []string{}
				for tag := range m.filteredTags {
					tags = append(tags, tag)
				}
				filterConditions = append(filterConditions, fmt.Sprintf("Tags: %s", strings.Join(tags, ", ")))
			}
			if m.searchKeyword != "" {
				filterConditions = append(filterConditions, fmt.Sprintf("Search: \"%s\"", m.searchKeyword))
			}

			if len(filterConditions) > 0 {
				s += fmt.Sprintf("Active Filters: %s\n\n", strings.Join(filterConditions, "; "))
			}
		}

		if len(m.tasks) == 0 {
			s += "No tasks found. Press 'a' to add a new task.\n\n"
		} else {
			for i, t := range m.tasks {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}

				// 色とアイコンを適用
				statusIcon := statusIcons[t.Status]
				priorityColor := priorityColors[t.Priority]
				// 検索キーワードのハイライト
				displayTitle := t.Title
				if m.searchKeyword != "" {
					// タイトル内のキーワードをハイライト
					displayTitle = strings.ReplaceAll(displayTitle, m.searchKeyword, lipgloss.NewStyle().Background(lipgloss.Color("205")).Render(m.searchKeyword))
				}

				styledTitle := lipgloss.NewStyle().Foreground(priorityColor).Render(displayTitle)

				s += fmt.Sprintf("%s %s %s %s\n", cursor, statusIcon, styledTitle, lipgloss.NewStyle().Foreground(priorityColor).Render(string(t.Priority)))
			}
		}

		total, completed, incomplete := m.app.GetTaskStats()
		s += fmt.Sprintf("\nTotal: %d | Incomplete: %d | Completed: %d\n", total, incomplete, completed)
		s += fmt.Sprintf("Sorted by: %s %s\n\n", m.sortBy, func() string {
			if m.sortAsc {
				return "asc"
			}
			return "desc"
		}())
		s += "[a]dd [e]dit [d]elete [v]iew [c]omplete [f]ilter [p]riority filter [t]ag filter [s]earch [o]sort [g]settings [x]export [i]import [q]uit [h]elp\n"
		return s

	case "detail":
		if m.detailViewTask == nil {
			return "Error: No task selected for detail view. Press [esc] to return to main."
		}
		t := m.detailViewTask
		s := "Task Details\n\n"
		s += fmt.Sprintf("ID: %s\n", t.ID)
		s += fmt.Sprintf("Title: %s\n", t.Title)
		s += fmt.Sprintf("Description: %s\n", t.Description)
		s += fmt.Sprintf("Status: %s\n", t.Status)
		s += fmt.Sprintf("Priority: %s\n", t.Priority)
		s += fmt.Sprintf("Tags: %s\n", strings.Join(t.Tags, ", "))
		s += fmt.Sprintf("Created At: %s\n", t.CreatedAt.Format("2006-01-02 15:04:05"))
		s += fmt.Sprintf("Updated At: %s\n", t.UpdatedAt.Format("2006-01-02 15:04:05"))
		if t.CompletedAt != nil {
			s += fmt.Sprintf("Completed At: %s\n", t.CompletedAt.Format("2006-01-02 15:04:05"))
		}
		s += "\n[esc] to back\n"
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
	case "search":
		return fmt.Sprintf(
			"Search Tasks by Keyword (Title or Description)\n\n%s\n\n%s",
			m.searchInput.View(),
			"[enter] to search, [esc] to cancel",
		)
	case "sort":
		return fmt.Sprintf(
			"Sort Tasks (e.g., created_at asc, priority desc)\n\n%s\n\n%s",
			m.sortInput.View(),
			"[enter] to apply sort, [esc] to cancel",
		)
	case "settings":
		return fmt.Sprintf(
			"Settings\n\n%s\n%s\n%s\n\n%s",
			m.defaultPriorityInput.View(),
			m.autoSaveInput.View(),
			m.themeInput.View(),
			"[enter] to save, [esc] to cancel",
		)
	case "export":
		return fmt.Sprintf(
			"Export Tasks to JSON (e.g., /path/to/tasks.json)\n\n%s\n\n%s",
			m.exportInput.View(),
			"[enter] to export, [esc] to cancel",
		)
	case "import":
		return fmt.Sprintf(
			"Import Tasks from JSON (e.g., /path/to/tasks.json)\n\n%s\n\n%s",
			m.importInput.View(),
			"[enter] to import, [esc] to cancel",
		)
	}

	return "Unknown view"
}

func main() {
	// プロファイリングを有効にするには、環境変数 GO_TASK_PROFILE を設定します。
	if os.Getenv("GO_TASK_PROFILE") == "true" {
		defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	}

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Error("Application failed:", err)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
