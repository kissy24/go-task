package config

import (
	"os"
	"path/filepath"
	"testing"

	"zan/internal/task"
)

func TestGetConfigFilePath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}
	expectedPath := filepath.Join(homeDir, ConfigDir, ConfigFileName)

	path, err := GetConfigFilePath()
	if err != nil {
		t.Fatalf("GetConfigFilePath returned an error: %v", err)
	}
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}

	// Clean up created directory if it was created
	os.RemoveAll(filepath.Join(homeDir, ConfigDir))
}

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	if cfg.Settings.DefaultPriority != task.PriorityMedium {
		t.Errorf("Expected default priority to be Medium, got %s", cfg.Settings.DefaultPriority)
	}
	if !cfg.Settings.AutoSave {
		t.Errorf("Expected auto save to be true, got %t", cfg.Settings.AutoSave)
	}
	if cfg.Settings.Theme != "default" {
		t.Errorf("Expected theme to be 'default', got %s", cfg.Settings.Theme)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Ensure a clean state
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}
	configDirPath := filepath.Join(homeDir, ConfigDir)
	os.RemoveAll(configDirPath)
	defer os.RemoveAll(configDirPath)

	// Test saving default config
	defaultCfg := NewDefaultConfig()
	err = defaultCfg.SaveConfig()
	if err != nil {
		t.Fatalf("Failed to save default config: %v", err)
	}

	// Test loading config
	loadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.Settings.DefaultPriority != defaultCfg.Settings.DefaultPriority {
		t.Errorf("Loaded default priority mismatch: expected %s, got %s", defaultCfg.Settings.DefaultPriority, loadedCfg.Settings.DefaultPriority)
	}
	if loadedCfg.Settings.AutoSave != defaultCfg.Settings.AutoSave {
		t.Errorf("Loaded auto save mismatch: expected %t, got %t", defaultCfg.Settings.AutoSave, loadedCfg.Settings.AutoSave)
	}
	if loadedCfg.Settings.Theme != defaultCfg.Settings.Theme {
		t.Errorf("Loaded theme mismatch: expected %s, got %s", defaultCfg.Settings.Theme, loadedCfg.Settings.Theme)
	}

	// Test modifying and saving config
	loadedCfg.Settings.DefaultPriority = task.PriorityHigh
	loadedCfg.Settings.AutoSave = false
	loadedCfg.Settings.Theme = "dark"
	err = loadedCfg.SaveConfig()
	if err != nil {
		t.Fatalf("Failed to save modified config: %v", err)
	}

	// Test loading modified config
	reloadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to reload modified config: %v", err)
	}

	if reloadedCfg.Settings.DefaultPriority != task.PriorityHigh {
		t.Errorf("Reloaded default priority mismatch: expected %s, got %s", task.PriorityHigh, reloadedCfg.Settings.DefaultPriority)
	}
	if reloadedCfg.Settings.AutoSave != false {
		t.Errorf("Reloaded auto save mismatch: expected %t, got %t", false, reloadedCfg.Settings.AutoSave)
	}
	if reloadedCfg.Settings.Theme != "dark" {
		t.Errorf("Reloaded theme mismatch: expected %s, got %s", "dark", reloadedCfg.Settings.Theme)
	}

	// Test loading non-existent config (should create default)
	os.RemoveAll(configDirPath) // Remove config file to simulate non-existence
	newLoadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load non-existent config: %v", err)
	}
	if newLoadedCfg.Settings.DefaultPriority != task.PriorityMedium {
		t.Errorf("Expected default priority after non-existent load, got %s", newLoadedCfg.Settings.DefaultPriority)
	}
}
