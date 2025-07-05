package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"zan/internal/task"
)

var (
	dataDir              = ".zan"
	dataFile             = "tasks.json"
	filePerm os.FileMode = 0600
	dirPerm  os.FileMode = 0700
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

	data, err := ioutil.ReadFile(filePath)
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

	if err := ioutil.WriteFile(filePath, data, filePerm); err != nil {
		return fmt.Errorf("failed to write data file %s: %w", filePath, err)
	}

	return nil
}
