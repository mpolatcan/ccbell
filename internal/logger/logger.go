// Package logger provides debug logging with rotation for ccbell.
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// MaxLogSize is the maximum log file size before rotation (1MB).
	MaxLogSize = 1024 * 1024
	// RotateCount is the number of rotated log files to keep.
	RotateCount = 3
	// FileMode is the permission mode for log files.
	FileMode = 0600
)

// Logger handles debug logging with rotation.
type Logger struct {
	enabled  bool
	filePath string
	pid      int
	mu       sync.Mutex
}

// New creates a new Logger instance.
func New(enabled bool, homeDir string) *Logger {
	logPath := ""
	if homeDir != "" {
		logPath = filepath.Join(homeDir, ".claude", "ccbell.log")
	}

	return &Logger{
		enabled:  enabled,
		filePath: logPath,
		pid:      os.Getpid(),
	}
}

// Debug logs a message if debug mode is enabled.
func (l *Logger) Debug(format string, args ...interface{}) {
	if !l.enabled || l.filePath == "" {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Rotate if needed
	l.rotateIfNeeded()

	// Open file for appending
	f, err := os.OpenFile(l.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, FileMode)
	if err != nil {
		return // Silent failure - logging shouldn't break the app
	}
	defer f.Close()

	// Format and write
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(f, "[%s] [%d] %s\n", timestamp, l.pid, msg)
}

// rotateIfNeeded checks log size and rotates if necessary.
func (l *Logger) rotateIfNeeded() {
	info, err := os.Stat(l.filePath)
	if err != nil {
		return // File doesn't exist yet
	}

	if info.Size() < MaxLogSize {
		return
	}

	// Rotate: .log.2 -> .log.3, .log.1 -> .log.2, .log.0 -> .log.1, .log -> .log.0
	for i := RotateCount - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", l.filePath, i-1)
		newPath := fmt.Sprintf("%s.%d", l.filePath, i)
		// Best effort rotation - ignore errors for old rotated files
		// They may not exist, which is fine
		_ = os.Rename(oldPath, newPath)
	}

	// Main log rotation - if this fails, we'll just keep appending
	// to the existing file (better than losing logs)
	if err := os.Rename(l.filePath, l.filePath+".0"); err != nil {
		// Rotation failed - try to truncate instead to prevent unbounded growth
		if f, truncErr := os.OpenFile(l.filePath, os.O_TRUNC|os.O_WRONLY, FileMode); truncErr == nil {
			f.Close()
		}
	}
}

// SetEnabled enables or disables logging.
func (l *Logger) SetEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = enabled
}

// IsEnabled returns whether logging is enabled.
func (l *Logger) IsEnabled() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.enabled
}
