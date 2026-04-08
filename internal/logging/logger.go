package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
)

var levelNames = map[LogLevel]string{
	LevelDebug:   "DEBUG",
	LevelInfo:    "INFO",
	LevelWarning: "WARNING",
	LevelError:   "ERROR",
	LevelFatal:   "FATAL",
}

// Logger handles game logging
type Logger struct {
	file          *os.File
	logger        *log.Logger
	level         LogLevel
	consoleOutput bool
	fileOutput    bool
}

// Config holds logger configuration
type Config struct {
	LogDirectory   string
	LogFileName    string
	Level          LogLevel
	ConsoleOutput  bool
	FileOutput     bool
	MaxFileSizeMB  int
	MaxBackupFiles int
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *Config {
	return &Config{
		LogDirectory:   "./logs",
		LogFileName:    "game.log",
		Level:          LevelInfo,
		ConsoleOutput:  true,
		FileOutput:     true,
		MaxFileSizeMB:  10,
		MaxBackupFiles: 3,
	}
}

// NewLogger creates a new logger instance
func NewLogger(config *Config) (*Logger, error) {
	logger := &Logger{
		level:         config.Level,
		consoleOutput: config.ConsoleOutput,
		fileOutput:    config.FileOutput,
	}

	if config.FileOutput {
		// Create log directory if it doesn't exist
		if err := os.MkdirAll(config.LogDirectory, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file
		logPath := filepath.Join(config.LogDirectory, config.LogFileName)
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		logger.file = file

		// Create multi-writer for both file and console
		var writers []io.Writer
		if config.ConsoleOutput {
			writers = append(writers, os.Stdout)
		}
		writers = append(writers, file)

		logger.logger = log.New(io.MultiWriter(writers...), "", log.Ldate|log.Ltime)
	} else if config.ConsoleOutput {
		logger.logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	}

	return logger, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// log writes a message at the specified level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	caller := ""
	if ok {
		caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	message := fmt.Sprintf(format, args...)
	logMessage := fmt.Sprintf("[%s] %s - %s", levelNames[level], caller, message)

	if l.logger != nil {
		l.logger.Println(logMessage)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	l.log(LevelWarning, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// ErrorWithStack logs an error with stack trace
func (l *Logger) ErrorWithStack(err error, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log(LevelError, "%s: %v\nStack trace:\n%s", message, err, getStackTrace())
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelFatal, format, args...)
	os.Exit(1)
}

// getStackTrace returns the current stack trace
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// LogGameEvent logs a game event with structured data
func (l *Logger) LogGameEvent(eventType string, data map[string]interface{}) {
	l.Info("Game Event: %s - %v", eventType, data)
}

// LogPlayerAction logs player actions for debugging
func (l *Logger) LogPlayerAction(action string, details map[string]interface{}) {
	l.Debug("Player Action: %s - %v", action, details)
}

// LogCombat logs combat events
func (l *Logger) LogCombat(attacker, target string, damage int, result string) {
	l.Info("Combat: %s attacked %s for %d damage - %s", attacker, target, damage, result)
}

// LogMovement logs player movement
func (l *Logger) LogMovement(from, to string, direction string) {
	l.Debug("Movement: %s -> %s (%s)", from, to, direction)
}

// LogError logs an error with context
func (l *Logger) LogError(err error, context map[string]interface{}) {
	l.Error("Error: %v - Context: %v", err, context)
}
