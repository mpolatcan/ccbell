package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name     string
		homeDir  string
		wantPath string
	}{
		{
			name:     "with home dir",
			homeDir:  "/home/user",
			wantPath: "/home/user/.claude/ccbell.state",
		},
		{
			name:     "empty home dir",
			homeDir:  "",
			wantPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.homeDir)
			// Normalize path separators for Windows compatibility
			gotPath := m.filePath
			if runtime.GOOS == "windows" {
				gotPath = filepath.ToSlash(gotPath)
			}
			if gotPath != tt.wantPath {
				t.Errorf("filePath = %v, want %v", m.filePath, tt.wantPath)
			}
		})
	}
}

func TestManager_CheckCooldown(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccbell-state-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("no cooldown when cooldownSecs is 0", func(t *testing.T) {
		m := NewManager(tmpDir)
		inCooldown, err := m.CheckCooldown("stop", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inCooldown {
			t.Error("should not be in cooldown when cooldownSecs is 0")
		}
	})

	t.Run("no cooldown when filePath is empty", func(t *testing.T) {
		m := NewManager("")
		inCooldown, err := m.CheckCooldown("stop", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inCooldown {
			t.Error("should not be in cooldown when filePath is empty")
		}
	})

	t.Run("first trigger not in cooldown", func(t *testing.T) {
		m := NewManager(tmpDir)
		// Clean up any existing state
		m.Clear()

		inCooldown, err := m.CheckCooldown("stop", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inCooldown {
			t.Error("first trigger should not be in cooldown")
		}
	})

	t.Run("second trigger within cooldown", func(t *testing.T) {
		m := NewManager(tmpDir)
		m.Clear()

		// First trigger
		_, err := m.CheckCooldown("stop", 60)
		if err != nil {
			t.Fatalf("first trigger error: %v", err)
		}

		// Immediate second trigger should be in cooldown
		inCooldown, err := m.CheckCooldown("stop", 60)
		if err != nil {
			t.Fatalf("second trigger error: %v", err)
		}
		if !inCooldown {
			t.Error("second immediate trigger should be in cooldown")
		}
	})

	t.Run("different events have separate cooldowns", func(t *testing.T) {
		m := NewManager(tmpDir)
		m.Clear()

		// Trigger stop event
		_, err := m.CheckCooldown("stop", 60)
		if err != nil {
			t.Fatalf("stop trigger error: %v", err)
		}

		// Different event should not be in cooldown
		inCooldown, err := m.CheckCooldown("permission_prompt", 60)
		if err != nil {
			t.Fatalf("permission_prompt trigger error: %v", err)
		}
		if inCooldown {
			t.Error("different event should not be in cooldown")
		}
	})
}

func TestManager_GetLastTrigger(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ccbell-state-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := NewManager(tmpDir)
	m.Clear()

	t.Run("returns zero time for unknown event", func(t *testing.T) {
		lastTrigger, err := m.GetLastTrigger("unknown")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !lastTrigger.IsZero() {
			t.Error("should return zero time for unknown event")
		}
	})

	t.Run("returns correct time after trigger", func(t *testing.T) {
		before := time.Now().Add(-time.Second)

		_, err := m.CheckCooldown("stop", 10)
		if err != nil {
			t.Fatalf("trigger error: %v", err)
		}

		after := time.Now().Add(time.Second)

		lastTrigger, err := m.GetLastTrigger("stop")
		if err != nil {
			t.Fatalf("GetLastTrigger error: %v", err)
		}

		if lastTrigger.Before(before) || lastTrigger.After(after) {
			t.Errorf("lastTrigger %v not between %v and %v", lastTrigger, before, after)
		}
	})
}

func TestManager_Clear(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ccbell-state-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := NewManager(tmpDir)

	// Create state
	_, err = m.CheckCooldown("stop", 10)
	if err != nil {
		t.Fatalf("trigger error: %v", err)
	}

	// Verify state file exists
	if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
		t.Fatal("state file should exist")
	}

	// Clear
	if err := m.Clear(); err != nil {
		t.Fatalf("Clear error: %v", err)
	}

	// Verify state file deleted
	if _, err := os.Stat(m.filePath); !os.IsNotExist(err) {
		t.Error("state file should be deleted after Clear")
	}
}

func TestManager_CorruptedStateFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ccbell-state-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := NewManager(tmpDir)

	// Write corrupted JSON
	if err := os.WriteFile(m.filePath, []byte("not valid json"), FileMode); err != nil {
		t.Fatal(err)
	}

	// Should handle corrupted file gracefully
	inCooldown, err := m.CheckCooldown("stop", 10)
	if err != nil {
		t.Fatalf("should not error on corrupted file: %v", err)
	}
	if inCooldown {
		t.Error("should not be in cooldown with corrupted state")
	}
}

func TestManager_AtomicSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ccbell-state-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := NewManager(tmpDir)

	// Trigger to create state
	_, err = m.CheckCooldown("stop", 10)
	if err != nil {
		t.Fatalf("trigger error: %v", err)
	}

	// Verify file is valid JSON
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("state file should be valid JSON: %v", err)
	}

	if state.LastTrigger["stop"] == 0 {
		t.Error("stop event should have timestamp")
	}

	// Verify file permissions (skip on Windows - different permission model)
	if runtime.GOOS != "windows" {
		info, err := os.Stat(m.filePath)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != FileMode {
			t.Errorf("file mode = %v, want %v", info.Mode().Perm(), FileMode)
		}
	}
}
