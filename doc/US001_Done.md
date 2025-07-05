# US-001: タスクデータ構造の定義 完了報告

## 完了定義

- [x] タスクのデータ構造が定義されている
- [x] JSON形式でのシリアライズ/デシリアライズが可能
- [x] 必須フィールドと任意フィールドが明確に分かれている

## 完了のエビデンス

### タスク構造体の定義とJSONタグの設定

`internal/task/task.go` に `Task` 構造体、`Status` および `Priority` 列挙型、`Tasks` および `Settings` 構造体を定義しました。各フィールドにはJSONタグを設定し、JSON形式でのシリアライズ/デシリアライズに対応しています。

```go
// internal/task/task.go
package task

import (
	"errors"
	"fmt"
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
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      Status    `json:"status"`
	Priority    Priority  `json:"priority"`
	Tags        []string  `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
```

### データ検証ロジックの実装

`Task` 構造体に `Validate()` メソッドを実装し、`ID` と `Title` が空でないこと、`Status` と `Priority` が有効な値であることを検証します。

### 単体テストの作成

`internal/task/task_test.go` に `TestTaskValidate` 関数を実装し、`Validate()` メソッドの単体テストを作成しました。これにより、有効なタスクと無効なタスクの両方で検証ロジックが正しく機能することを確認しています。

```go
// internal/task/task_test.go
package task

import (
	"testing"
	"time"
)

func TestTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		wantErr bool
	}{
		{
			name: "Valid Task",
			task: Task{
				ID:        "test-id-1",
				Title:     "Test Task",
				Status:    StatusTODO,
				Priority:  PriorityMedium,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "Empty ID",
			task: Task{
				ID:        "",
				Title:     "Test Task",
				Status:    StatusTODO,
				Priority:  PriorityMedium,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "Empty Title",
			task: Task{
				ID:        "test-id-2",
				Title:     "",
				Status:    StatusTODO,
				Priority:  PriorityMedium,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "Invalid Status",
			task: Task{
				ID:        "test-id-3",
				Title:     "Test Task",
				Status:    "INVALID_STATUS",
				Priority:  PriorityMedium,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "Invalid Priority",
			task: Task{
				ID:        "test-id-4",
				Title:     "Test Task",
				Status:    StatusTODO,
				Priority:  "INVALID_PRIORITY",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Task.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}