package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"zan/internal/task"
)

const (
	ConfigFileName = "config.json"
	ConfigDir      = ".zan"
)

type Settings struct {
	DefaultPriority task.Priority `json:"default_priority"`
	AutoSave        bool          `json:"auto_save"`
	Theme           string        `json:"theme"`
}

type Config struct {
	Settings Settings `json:"settings"`
	mu       sync.RWMutex
}

func NewDefaultConfig() *Config {
	return &Config{
		Settings: Settings{
			DefaultPriority: task.PriorityMedium,
			AutoSave:        true,
			Theme:           "default",
		},
	}
}

func GetConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ConfigDir)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return "", err
		}
	}
	return filepath.Join(configDir, ConfigFileName), nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := NewDefaultConfig()
			if err := cfg.SaveConfig(); err != nil {
				return nil, err
			}
			return cfg, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) SaveConfig() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	configPath, err := GetConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}
