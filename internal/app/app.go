package app

import (
	"errors"
	"fmt"
	"time"

	"zan/internal/store"
	"zan/internal/task"

	"github.com/google/uuid"
)

// App はアプリケーションの主要なロジックを管理します。
type App struct {
	Tasks *task.Tasks
}

// NewApp は新しいAppインスタンスを作成し、タスクデータをロードします。
func NewApp() (*App, error) {
	tasks, err := store.LoadTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}
	return &App{Tasks: tasks}, nil
}

// AddTask は新しいタスクを作成し、タスクリストに追加します。
func (a *App) AddTask(title, description string, priority task.Priority, tags []string) (*task.Task, error) {
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	if priority == "" {
		priority = a.Tasks.Settings.DefaultPriority
	}

	newTask := task.Task{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Status:      task.StatusTODO,
		Priority:    priority,
		Tags:        tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := newTask.Validate(); err != nil {
		return nil, fmt.Errorf("invalid task data: %w", err)
	}

	a.Tasks.Tasks = append(a.Tasks.Tasks, newTask)
	if a.Tasks.Settings.AutoSave {
		if err := store.SaveTasks(a.Tasks); err != nil {
			return nil, fmt.Errorf("failed to auto-save tasks: %w", err)
		}
	}
	return &newTask, nil
}

// GetTaskByID は指定されたIDのタスクを返します。
func (a *App) GetTaskByID(id string) (*task.Task, error) {
	for _, t := range a.Tasks.Tasks {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("task with ID %s not found", id)
}

// UpdateTask は既存のタスクを更新します。
func (a *App) UpdateTask(id, title, description string, status task.Status, priority task.Priority, tags []string) (*task.Task, error) {
	for i, t := range a.Tasks.Tasks {
		if t.ID == id {
			if title != "" {
				a.Tasks.Tasks[i].Title = title
			}
			if description != "" {
				a.Tasks.Tasks[i].Description = description
			}
			if status != "" {
				a.Tasks.Tasks[i].Status = status
				if status == task.StatusDone {
					now := time.Now()
					a.Tasks.Tasks[i].CompletedAt = &now
				} else {
					a.Tasks.Tasks[i].CompletedAt = nil
				}
			}
			if priority != "" {
				a.Tasks.Tasks[i].Priority = priority
			}
			if tags != nil {
				a.Tasks.Tasks[i].Tags = tags
			}
			a.Tasks.Tasks[i].UpdatedAt = time.Now()

			if err := a.Tasks.Tasks[i].Validate(); err != nil {
				return nil, fmt.Errorf("invalid task data after update: %w", err)
			}

			if a.Tasks.Settings.AutoSave {
				if err := store.SaveTasks(a.Tasks); err != nil {
					return nil, fmt.Errorf("failed to auto-save tasks: %w", err)
				}
			}
			return &a.Tasks.Tasks[i], nil
		}
	}
	return nil, fmt.Errorf("task with ID %s not found", id)
}

// DeleteTask は指定されたIDのタスクを削除します。
func (a *App) DeleteTask(id string) error {
	for i, t := range a.Tasks.Tasks {
		if t.ID == id {
			a.Tasks.Tasks = append(a.Tasks.Tasks[:i], a.Tasks.Tasks[i+1:]...)
			if a.Tasks.Settings.AutoSave {
				if err := store.SaveTasks(a.Tasks); err != nil {
					return fmt.Errorf("failed to auto-save tasks after deletion: %w", err)
				}
			}
			return nil
		}
	}
	return fmt.Errorf("task with ID %s not found", id)
}

// GetAllTasks は全てのタスクを返します。
func (a *App) GetAllTasks() []task.Task {
	return a.Tasks.Tasks
}
