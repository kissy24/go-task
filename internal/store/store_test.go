package store

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-task/internal/task"
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
