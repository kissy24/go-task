package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"zan/internal/task"
)

const (
	dataDir                          = ".zan"
	dataFile                         = "tasks.json"
	backupDir                        = "backup"
	backupFileTimeLayout             = "20060102150405" // YYYYMMDDhhmmss
	maxBackups                       = 5                // Keep last 5 backups
	filePerm             os.FileMode = 0600
	dirPerm              os.FileMode = 0700
)

// GetConfigDirPath は設定ディレクトリのパスを返します。
func GetConfigDirPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, dataDir), nil
}

// GetDataFilePath はデータファイルのパスを返します。
func GetDataFilePath() (string, error) {
	configDir, err := GetConfigDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, dataFile), nil
}

// EnsureDataDirExists はデータディレクトリが存在しない場合に作成します。
func EnsureDataDirExists() error {
	configDir, err := GetConfigDirPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, dirPerm); err != nil {
			return fmt.Errorf("failed to create data directory %s: %w", configDir, err)
		}
	}
	return nil
}

// LoadTasks はデータファイルからタスクを読み込みます。
func LoadTasks() (*task.Tasks, error) {
	filePath, err := GetDataFilePath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// ファイルが存在しない場合は新しいTasks構造体を返す
		return &task.Tasks{
			Version:   "1.0.0",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tasks:     []task.Task{},
			Settings: task.Settings{
				DefaultPriority: task.PriorityMedium,
				AutoSave:        true,
				Theme:           "default",
			},
		}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file %s: %w", filePath, err)
	}

	var tasks task.Tasks
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks data: %w", err)
	}

	return &tasks, nil
}

// SaveTasks はタスクをデータファイルに保存します。
func SaveTasks(tasks *task.Tasks) error {
	if err := EnsureDataDirExists(); err != nil {
		return err
	}

	filePath, err := GetDataFilePath()
	if err != nil {
		return err
	}

	tasks.UpdatedAt = time.Now() // 更新日時を自動更新

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks data: %w", err)
	}

	if err := os.WriteFile(filePath, data, filePerm); err != nil {
		return fmt.Errorf("failed to write data file %s: %w", filePath, err)
	}

	return nil
}

// MarshalTasks はタスクデータをJSON形式にマーシャルします。
func MarshalTasks(tasks *task.Tasks) ([]byte, error) {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tasks data: %w", err)
	}
	return data, nil
}

// GetBackupDirPath はバックアップディレクトリのパスを返します。
func GetBackupDirPath() (string, error) {
	configDir, err := GetConfigDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, backupDir), nil
}

// EnsureBackupDirExists はバックアップディレクトリが存在しない場合に作成します。
func EnsureBackupDirExists() error {
	backupPath, err := GetBackupDirPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		if err := os.MkdirAll(backupPath, dirPerm); err != nil {
			return fmt.Errorf("failed to create backup directory %s: %w", backupPath, err)
		}
	}
	return nil
}

// CreateBackup は現在のタスクデータのバックアップを作成します。
func CreateBackup(tasks *task.Tasks) error {
	if err := EnsureBackupDirExists(); err != nil {
		return err
	}

	backupPath, err := GetBackupDirPath()
	if err != nil {
		return err
	}

	timestamp := time.Now().Format(backupFileTimeLayout)
	backupFileName := fmt.Sprintf("tasks_backup_%s.json", timestamp)
	backupFilePath := filepath.Join(backupPath, backupFileName)

	data, err := MarshalTasks(tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks for backup: %w", err)
	}

	if err := os.WriteFile(backupFilePath, data, filePerm); err != nil {
		return fmt.Errorf("failed to write backup file %s: %w", backupFilePath, err)
	}

	return nil
}

// CleanOldBackups は古いバックアップファイルを削除します。
func CleanOldBackups() error {
	backupPath, err := GetBackupDirPath()
	if err != nil {
		return err
	}

	files, err := os.ReadDir(backupPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // バックアップディレクトリが存在しない場合は何もしない
		}
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	// バックアップファイルをフィルタリングし、作成日時でソート
	var backupFiles []os.FileInfo
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "tasks_backup_") && strings.HasSuffix(file.Name(), ".json") {
			info, err := file.Info()
			if err != nil {
				continue
			}
			backupFiles = append(backupFiles, info)
		}
	}

	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[i].ModTime().Before(backupFiles[j].ModTime())
	})

	// 古いバックアップを削除
	if len(backupFiles) > maxBackups {
		for i := 0; i < len(backupFiles)-maxBackups; i++ {
			filePath := filepath.Join(backupPath, backupFiles[i].Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to remove old backup file %s: %w", filePath, err)
			}
		}
	}
	return nil
}
