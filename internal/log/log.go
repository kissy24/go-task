package log

import (
	"log"
	"os"
	"path/filepath"
)

var logger *log.Logger

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home directory: %v", err)
	}
	logDir := filepath.Join(home, ".go-task")
	if err := os.MkdirAll(logDir, 0700); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	logFile, err := os.OpenFile(filepath.Join(logDir, "app.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger = log.New(logFile, "", log.LstdFlags)
}

// Info は情報レベルのログを出力します。
func Info(v ...interface{}) {
	logger.SetPrefix("INFO: ")
	logger.Println(v...)
}

// Error はエラーレベルのログを出力します。
func Error(v ...interface{}) {
	logger.SetPrefix("ERROR: ")
	logger.Println(v...)
}
