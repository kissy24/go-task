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
