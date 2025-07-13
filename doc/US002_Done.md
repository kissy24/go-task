# US-002: データ永続化機能 完了報告

## 完了定義

- [x] `~/.go-task/tasks.json` にデータが保存される
- [x] ディレクトリが存在しない場合は自動作成される
- [x] ファイル権限が適切に設定される (600)

## 完了のエビデンス

### ファイル入出力機能、ディレクトリ作成処理、ファイル権限設定の実装

`internal/store/store.go` に以下の関数を実装しました。

- `GetConfigDirPath()`: 設定ディレクトリのパス (`~/.go-task`) を取得します。
- `GetDataFilePath()`: データファイルのパス (`~/.go-task/tasks.json`) を取得します。
- `EnsureDataDirExists()`: データディレクトリが存在しない場合に `0700` の権限で作成します。
- `LoadTasks()`: データファイルからタスクデータを読み込みます。ファイルが存在しない場合は初期データ構造を返します。
- `SaveTasks()`: タスクデータをデータファイルに `0600` の権限で保存します。

```go
// internal/store/store.go
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
	dataDir  = ".go-task"
	dataFile = "tasks.json"
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
```

### 単体テストの作成

`internal/store/store_test.go` に以下のテスト関数を実装し、データ永続化機能の単体テストを行いました。

- `setupTestEnv()`: テスト用のHOMEディレクトリを設定し、テスト終了後に元に戻します。
- `TestGetConfigDirPath()`: 設定ディレクトリパスの取得をテストします。
- `TestGetDataFilePath()`: データファイルパスの取得をテストします。
- `TestEnsureDataDirExists()`: データディレクトリの作成と存在チェックをテストします。
- `TestLoadAndSaveTasks()`: タスクの読み込みと保存、およびファイル権限のテストを行います。

これらのテストにより、データが正しく保存され、ディレクトリが適切に作成され、ファイル権限が設定されることを確認しました。

```go
// internal/store/store_test.go
package store

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"zan/internal/task"
)

// setupTestEnv はテスト用の環境変数を設定し、テスト終了後に元に戻します。
func setupTestEnv(t *testing.T, tempDir string) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})
}

func TestGetConfigDirPath(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "go-task_test_config_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	expectedPath := filepath.Join(tmpDir, dataDir)
	path, err := GetConfigDirPath()
	if err != nil {
		t.Errorf("GetConfigDirPath() error = %v, wantErr %v", err, false)
	}
	if path != expectedPath {
		t.Errorf("GetConfigDirPath() got = %s, want %s", path, expectedPath)
	}
}

func TestGetDataFilePath(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "go-task_test_data_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	expectedPath := filepath.Join(tmpDir, dataDir, dataFile)
	path, err := GetDataFilePath()
	if err != nil {
		t.Errorf("GetDataFilePath() error = %v, wantErr %v", err, false)
	}
	if path != expectedPath {
		t.Errorf("GetDataFilePath() got = %s, want %s", path, expectedPath)
	}
}

func TestEnsureDataDirExists(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "go-task_test_ensure_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	configDir := filepath.Join(tmpDir, dataDir)
	os.RemoveAll(configDir) // 確実に存在しない状態にする

	err = EnsureDataDirExists()
	if err != nil {
		t.Errorf("EnsureDataDirExists() error = %v, wantErr %v", err, false)
	}
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Data directory %s was not created", configDir)
	}

	// 既に存在するディレクトリの場合
	err = EnsureDataDirExists()
	if err != nil {
		t.Errorf("EnsureDataDirExists() error = %v, wantErr %v", err, false)
	}
}

func TestLoadAndSaveTasks(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "go-task_test_loadsave_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnv(t, tmpDir)

	// 初期状態（ファイルが存在しない場合）
	loadedTasks, err := LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v, wantErr %v", err, false)
	}
	if loadedTasks == nil || len(loadedTasks.Tasks) != 0 {
		t.Errorf("Expected empty tasks on initial load, got %v", loadedTasks)
	}

	// タスクを追加して保存
	testTask := task.Task{
		ID:        "test-id-1",
		Title:     "Test Task",
		Status:    task.StatusTODO,
		Priority:  task.PriorityHigh,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	loadedTasks.Tasks = append(loadedTasks.Tasks, testTask)
	loadedTasks.Version = "1.0.0"
	loadedTasks.Settings = task.Settings{
		DefaultPriority: task.PriorityMedium,
		AutoSave:        true,
		Theme:           "default",
	}

	err = SaveTasks(loadedTasks)
	if err != nil {
		t.Fatalf("SaveTasks() error = %v, wantErr %v", err, false)
	}

	// 保存したファイルを再度読み込み
	reloadedTasks, err := LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v, wantErr %v", err, false)
	}

	if len(reloadedTasks.Tasks) != 1 {
		t.Errorf("Expected 1 task after reload, got %d", len(reloadedTasks.Tasks))
	}
	if reloadedTasks.Tasks[0].ID != testTask.ID {
		t.Errorf("Expected task ID %s, got %s", testTask.ID, reloadedTasks.Tasks[0].ID)
	}

	// ファイル権限の確認
	filePath, _ := GetDataFilePath()
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file %s: %v", filePath, err)
	}
	if info.Mode().Perm() != filePerm {
		t.Errorf("Expected file permission %o, got %o", filePerm, info.Mode().Perm())
	}
}