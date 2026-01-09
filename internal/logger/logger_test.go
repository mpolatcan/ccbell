package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		homeDir  string
		wantPath string
	}{
		{
			name:     "enabled with home dir",
			enabled:  true,
			homeDir:  "/home/user",
			wantPath: "/home/user/.claude/ccbell.log",
		},
		{
			name:     "disabled with home dir",
			enabled:  false,
			homeDir:  "/home/user",
			wantPath: "/home/user/.claude/ccbell.log",
		},
		{
			name:     "empty home dir",
			enabled:  true,
			homeDir:  "",
			wantPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.enabled, tt.homeDir)
			if l.enabled != tt.enabled {
				t.Errorf("enabled = %v, want %v", l.enabled, tt.enabled)
			}
			if l.filePath != tt.wantPath {
				t.Errorf("filePath = %v, want %v", l.filePath, tt.wantPath)
			}
			if l.pid == 0 {
				t.Error("pid should not be 0")
			}
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccbell-logger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("logs when enabled", func(t *testing.T) {
		l := New(true, tmpDir)
		l.Debug("test message %s", "arg1")

		content, err := os.ReadFile(l.filePath)
		if err != nil {
			t.Fatalf("failed to read log file: %v", err)
		}

		if !strings.Contains(string(content), "test message arg1") {
			t.Errorf("log content = %q, want to contain 'test message arg1'", content)
		}
	})

	t.Run("does not log when disabled", func(t *testing.T) {
		l := New(false, tmpDir)
		logPath := filepath.Join(claudeDir, "disabled.log")
		l.filePath = logPath

		l.Debug("should not appear")

		_, err := os.Stat(logPath)
		if !os.IsNotExist(err) {
			t.Error("log file should not exist when disabled")
		}
	})

	t.Run("does not log with empty path", func(_ *testing.T) {
		l := New(true, "")
		l.Debug("should not crash")
		// Should not panic
	})
}

func TestLogger_SetEnabled(t *testing.T) {
	l := New(false, "/tmp")

	if l.IsEnabled() {
		t.Error("should start disabled")
	}

	l.SetEnabled(true)
	if !l.IsEnabled() {
		t.Error("should be enabled after SetEnabled(true)")
	}

	l.SetEnabled(false)
	if l.IsEnabled() {
		t.Error("should be disabled after SetEnabled(false)")
	}
}

func TestLogger_RotateIfNeeded(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccbell-rotate-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	l := New(true, tmpDir)

	// Create a file larger than MaxLogSize
	largeContent := strings.Repeat("x", MaxLogSize+100)
	if err := os.WriteFile(l.filePath, []byte(largeContent), FileMode); err != nil {
		t.Fatal(err)
	}

	// Log something to trigger rotation
	l.Debug("trigger rotation")

	// Check that rotation happened
	rotatedPath := l.filePath + ".0"
	if _, err := os.Stat(rotatedPath); os.IsNotExist(err) {
		t.Error("rotated file should exist")
	}

	// Original log should now be small (just the new message)
	info, err := os.Stat(l.filePath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() >= MaxLogSize {
		t.Errorf("log file should be smaller after rotation, got %d bytes", info.Size())
	}
}
