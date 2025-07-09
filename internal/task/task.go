package task

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Status はタスクの状態を表す列挙型です。
type Status string

const (
	StatusTODO       Status = "TODO"
	StatusInProgress Status = "IN_PROGRESS"
	StatusDone       Status = "DONE"
	StatusPending    Status = "PENDING"
)

// Priority はタスクの優先度を表す列挙型です。
type Priority string

const (
	PriorityHigh   Priority = "HIGH"
	PriorityMedium Priority = "MEDIUM"
	PriorityLow    Priority = "LOW"
)

// Task は単一のタスクのデータ構造を定義します。
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      Status     `json:"status"`
	Priority    Priority   `json:"priority"`
	Tags        []string   `json:"tags,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"` // 完了時のみ設定されるためポインタ
}

// Tasks はタスクのリストと全体データ構造を定義します。
type Tasks struct {
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tasks     []Task    `json:"tasks"`
	Settings  Settings  `json:"settings"`
}

// Settings はアプリケーションの設定を定義します。
type Settings struct {
	DefaultPriority Priority `json:"default_priority"`
	AutoSave        bool     `json:"auto_save"`
	Theme           string   `json:"theme"`
}

// Validate はTask構造体のフィールドが有効な値を持っているか検証します。
func (t *Task) Validate() error {
	if t.ID == "" {
		return errors.New("Task ID cannot be empty")
	}
	if t.Title == "" {
		return errors.New("Task title cannot be empty")
	}
	switch t.Status {
	case StatusTODO, StatusInProgress, StatusDone, StatusPending:
		// 有効なステータス
	default:
		return fmt.Errorf("Invalid task status: %s", t.Status)
	}
	switch t.Priority {
	case PriorityHigh, PriorityMedium, PriorityLow:
		// 有効な優先度
	default:
		return fmt.Errorf("Invalid task priority: %s", t.Priority)
	}
	return nil
}

// SearchTasks はキーワードに基づいてタスクを検索します。
// タイトルと詳細説明に対して大文字小文字を区別しない部分一致検索を行います。
func SearchTasks(tasks []Task, keyword string) []Task {
	if keyword == "" {
		return tasks
	}

	var foundTasks []Task
	lowerKeyword := strings.ToLower(keyword)

	for _, t := range tasks {
		lowerTitle := strings.ToLower(t.Title)
		lowerDescription := strings.ToLower(t.Description)
		// デバッグログ
		// fmt.Printf("Comparing: Title='%s' (lower='%s'), Desc='%s' (lower='%s'), Keyword='%s'\n",
		// 	t.Title, lowerTitle, t.Description, lowerDescription, lowerKeyword)
		if strings.Contains(lowerTitle, lowerKeyword) ||
			strings.Contains(lowerDescription, lowerKeyword) {
			foundTasks = append(foundTasks, t)
		}
	}
	return foundTasks
}
