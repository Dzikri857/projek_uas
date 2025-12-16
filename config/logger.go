package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

// InitLogger initializes the logging system with rotating log files
func InitLogger() error {
	// Create logs directory if it doesn't exist
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Setup rotating log files
	infoLogFile := &lumberjack.Logger{
		Filename:   filepath.Join(logsDir, "info.log"),
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   // days
		Compress:   true, // compress rotated files
	}

	errorLogFile := &lumberjack.Logger{
		Filename:   filepath.Join(logsDir, "error.log"),
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	debugLogFile := &lumberjack.Logger{
		Filename:   filepath.Join(logsDir, "debug.log"),
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	// Create multi-writers (console + file)
	infoWriter := io.MultiWriter(os.Stdout, infoLogFile)
	errorWriter := io.MultiWriter(os.Stderr, errorLogFile)
	debugWriter := io.MultiWriter(os.Stdout, debugLogFile)

	// Initialize loggers
	InfoLogger = log.New(infoWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(debugWriter, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

	InfoLogger.Println("Logger initialized successfully")
	return nil
}

// LogInfo logs informational messages
func LogInfo(format string, v ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Printf(format, v...)
	}
}

// LogError logs error messages
func LogError(format string, v ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Printf(format, v...)
	}
}

// LogDebug logs debug messages
func LogDebug(format string, v ...interface{}) {
	if DebugLogger != nil {
		DebugLogger.Printf(format, v...)
	}
}

// GetLogFilePath returns the path for a specific log file
func GetLogFilePath(logType string) string {
	return filepath.Join("logs", fmt.Sprintf("%s_%s.log", logType, time.Now().Format("2006-01-02")))
}
