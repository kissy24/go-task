package app

import (
	"os"
	"testing"

	"go-task/internal/task"
)

func BenchmarkNewApp(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "go-task_bench_newapp_")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnvForBench(b, tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewApp()
		if err != nil {
			b.Fatalf("NewApp() failed: %v", err)
		}
	}
}

func BenchmarkAddTask(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "go-task_bench_addtask_")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	setupTestEnvForBench(b, tmpDir)

	app, err := NewApp()
	if err != nil {
		b.Fatalf("NewApp() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.AddTask("Benchmark Task", "description", task.PriorityMedium, nil)
	}
}

// setupTestEnv for benchmark
func setupTestEnvForBench(b *testing.B, tempDir string) {
	oldHome := os.Getenv("HOME")
	oldTestEnv := os.Getenv("GO_TASK_TEST_ENV")
	os.Setenv("HOME", tempDir)
	os.Setenv("GO_TASK_TEST_ENV", "true")
	b.Cleanup(func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("GO_TASK_TEST_ENV", oldTestEnv)
	})
}
