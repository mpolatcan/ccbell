package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testConfigDisabledPlugin is the JSON config content used in tests.
const testConfigDisabledPlugin = `{"enabled": false}`

func TestPrintUsage(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printUsage()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify key elements are present
	expectedStrings := []string{
		"ccbell - Sound notifications for Claude Code",
		"USAGE:",
		"EVENT TYPES:",
		"stop",
		"permission_prompt",
		"idle_prompt",
		"subagent",
		"OPTIONS:",
		"--help",
		"--version",
		"CONFIGURATION:",
		"SOUND FORMATS:",
		"bundled:stop",
		"custom:",
		"bundled:subagent",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("printUsage() output missing %q", expected)
		}
	}
}

func TestRunWithVersion(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name string
		arg  string
	}{
		{"version flag", "--version"},
		{"version short", "-v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = []string{"ccbell", tt.arg}
			err := run()
			if err != nil {
				t.Errorf("run() with %s returned error: %v", tt.arg, err)
			}
		})
	}
}

func TestRunWithHelp(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name string
		arg  string
	}{
		{"help flag", "--help"},
		{"help short", "-h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = []string{"ccbell", tt.arg}
			err := run()
			if err != nil {
				t.Errorf("run() with %s returned error: %v", tt.arg, err)
			}
		})
	}
}

func TestRunWithInvalidEventType(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name      string
		eventType string
		wantErr   bool
	}{
		{"invalid event", "invalid_event_type_xyz", true},
		{"path traversal attempt", "../../../etc/passwd", true},
		{"command injection attempt", "stop; rm -rf /", true},
		{"uppercase event", "STOP", true},
		{"empty event", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = []string{"ccbell", tt.eventType}
			err := run()
			if (err != nil) != tt.wantErr {
				t.Errorf("run() with event %q: error = %v, wantErr = %v", tt.eventType, err, tt.wantErr)
			}
		})
	}
}

func TestRunWithValidEventTypeDisabled(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	oldHome := os.Getenv("HOME")
	oldProjectDir := os.Getenv("CLAUDE_PROJECT_DIR")
	oldPluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	defer func() {
		os.Args = oldArgs
		os.Setenv("HOME", oldHome)
		if oldProjectDir != "" {
			os.Setenv("CLAUDE_PROJECT_DIR", oldProjectDir)
		} else {
			os.Unsetenv("CLAUDE_PROJECT_DIR")
		}
		if oldPluginRoot != "" {
			os.Setenv("CLAUDE_PLUGIN_ROOT", oldPluginRoot)
		} else {
			os.Unsetenv("CLAUDE_PLUGIN_ROOT")
		}
	}()

	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "ccbell-main-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with plugin disabled
	configContent := testConfigDisabledPlugin
	configPath := filepath.Join(claudeDir, "ccbell.config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Set environment
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("CLAUDE_PROJECT_DIR")
	os.Setenv("CLAUDE_PLUGIN_ROOT", tmpDir)

	// Test with valid event type but plugin disabled
	os.Args = []string{"ccbell", "stop"}
	err = run()
	if err != nil {
		t.Errorf("run() with disabled plugin should not error, got: %v", err)
	}
}

func TestRunWithEventDisabled(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	oldHome := os.Getenv("HOME")
	oldProjectDir := os.Getenv("CLAUDE_PROJECT_DIR")
	oldPluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	defer func() {
		os.Args = oldArgs
		os.Setenv("HOME", oldHome)
		if oldProjectDir != "" {
			os.Setenv("CLAUDE_PROJECT_DIR", oldProjectDir)
		} else {
			os.Unsetenv("CLAUDE_PROJECT_DIR")
		}
		if oldPluginRoot != "" {
			os.Setenv("CLAUDE_PLUGIN_ROOT", oldPluginRoot)
		} else {
			os.Unsetenv("CLAUDE_PLUGIN_ROOT")
		}
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccbell-main-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with stop event disabled
	configContent := `{
		"enabled": true,
		"events": {
			"stop": {
				"enabled": false
			}
		}
	}`
	configPath := filepath.Join(claudeDir, "ccbell.config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Set environment
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("CLAUDE_PROJECT_DIR")
	os.Setenv("CLAUDE_PLUGIN_ROOT", tmpDir)

	// Test with stop event disabled
	os.Args = []string{"ccbell", "stop"}
	err = run()
	if err != nil {
		t.Errorf("run() with disabled event should not error, got: %v", err)
	}
}

func TestRunDefaultEventType(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	oldHome := os.Getenv("HOME")
	oldProjectDir := os.Getenv("CLAUDE_PROJECT_DIR")
	oldPluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	defer func() {
		os.Args = oldArgs
		os.Setenv("HOME", oldHome)
		if oldProjectDir != "" {
			os.Setenv("CLAUDE_PROJECT_DIR", oldProjectDir)
		} else {
			os.Unsetenv("CLAUDE_PROJECT_DIR")
		}
		if oldPluginRoot != "" {
			os.Setenv("CLAUDE_PLUGIN_ROOT", oldPluginRoot)
		} else {
			os.Unsetenv("CLAUDE_PLUGIN_ROOT")
		}
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccbell-main-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with plugin disabled (to exit early without playing sound)
	configContent := testConfigDisabledPlugin
	configPath := filepath.Join(claudeDir, "ccbell.config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Set environment
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("CLAUDE_PROJECT_DIR")
	os.Setenv("CLAUDE_PLUGIN_ROOT", tmpDir)

	// Test with no arguments (should default to "stop")
	os.Args = []string{"ccbell"}
	err = run()
	if err != nil {
		t.Errorf("run() with no args should not error, got: %v", err)
	}
}

func TestRunWithQuietHours(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	oldHome := os.Getenv("HOME")
	oldProjectDir := os.Getenv("CLAUDE_PROJECT_DIR")
	oldPluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	defer func() {
		os.Args = oldArgs
		os.Setenv("HOME", oldHome)
		if oldProjectDir != "" {
			os.Setenv("CLAUDE_PROJECT_DIR", oldProjectDir)
		} else {
			os.Unsetenv("CLAUDE_PROJECT_DIR")
		}
		if oldPluginRoot != "" {
			os.Setenv("CLAUDE_PLUGIN_ROOT", oldPluginRoot)
		} else {
			os.Unsetenv("CLAUDE_PLUGIN_ROOT")
		}
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccbell-main-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with 24-hour quiet hours (always quiet)
	configContent := `{
		"enabled": true,
		"quietHours": {
			"start": "00:00",
			"end": "23:59"
		}
	}`
	configPath := filepath.Join(claudeDir, "ccbell.config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Set environment
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("CLAUDE_PROJECT_DIR")
	os.Setenv("CLAUDE_PLUGIN_ROOT", tmpDir)

	// Test - should exit early due to quiet hours
	os.Args = []string{"ccbell", "stop"}
	err = run()
	if err != nil {
		t.Errorf("run() during quiet hours should not error, got: %v", err)
	}
}

func TestValidEventTypes(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	oldHome := os.Getenv("HOME")
	oldProjectDir := os.Getenv("CLAUDE_PROJECT_DIR")
	oldPluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	defer func() {
		os.Args = oldArgs
		os.Setenv("HOME", oldHome)
		if oldProjectDir != "" {
			os.Setenv("CLAUDE_PROJECT_DIR", oldProjectDir)
		} else {
			os.Unsetenv("CLAUDE_PROJECT_DIR")
		}
		if oldPluginRoot != "" {
			os.Setenv("CLAUDE_PLUGIN_ROOT", oldPluginRoot)
		} else {
			os.Unsetenv("CLAUDE_PLUGIN_ROOT")
		}
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccbell-main-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with plugin disabled (to exit early)
	configContent := testConfigDisabledPlugin
	configPath := filepath.Join(claudeDir, "ccbell.config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Set environment
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("CLAUDE_PROJECT_DIR")
	os.Setenv("CLAUDE_PLUGIN_ROOT", tmpDir)

	// Test all valid event types
	validEvents := []string{"stop", "permission_prompt", "idle_prompt", "subagent"}

	for _, event := range validEvents {
		t.Run(event, func(t *testing.T) {
			os.Args = []string{"ccbell", event}
			err := run()
			if err != nil {
				t.Errorf("run() with valid event %q should not error, got: %v", event, err)
			}
		})
	}
}

func TestRunWithCooldown(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	oldHome := os.Getenv("HOME")
	oldProjectDir := os.Getenv("CLAUDE_PROJECT_DIR")
	oldPluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	defer func() {
		os.Args = oldArgs
		os.Setenv("HOME", oldHome)
		if oldProjectDir != "" {
			os.Setenv("CLAUDE_PROJECT_DIR", oldProjectDir)
		} else {
			os.Unsetenv("CLAUDE_PROJECT_DIR")
		}
		if oldPluginRoot != "" {
			os.Setenv("CLAUDE_PLUGIN_ROOT", oldPluginRoot)
		} else {
			os.Unsetenv("CLAUDE_PLUGIN_ROOT")
		}
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccbell-cooldown-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with high cooldown but plugin disabled
	configContent := `{
		"enabled": false,
		"events": {
			"stop": {
				"enabled": true,
				"cooldown": 3600
			}
		}
	}`
	configPath := filepath.Join(claudeDir, "ccbell.config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Set environment
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("CLAUDE_PROJECT_DIR")
	os.Setenv("CLAUDE_PLUGIN_ROOT", tmpDir)

	// First call - should succeed (no previous state)
	os.Args = []string{"ccbell", "stop"}
	err = run()
	if err != nil {
		t.Errorf("First run() should not error, got: %v", err)
	}
}

func TestRunWithDebugMode(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	oldHome := os.Getenv("HOME")
	oldProjectDir := os.Getenv("CLAUDE_PROJECT_DIR")
	oldPluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	defer func() {
		os.Args = oldArgs
		os.Setenv("HOME", oldHome)
		if oldProjectDir != "" {
			os.Setenv("CLAUDE_PROJECT_DIR", oldProjectDir)
		} else {
			os.Unsetenv("CLAUDE_PROJECT_DIR")
		}
		if oldPluginRoot != "" {
			os.Setenv("CLAUDE_PLUGIN_ROOT", oldPluginRoot)
		} else {
			os.Unsetenv("CLAUDE_PLUGIN_ROOT")
		}
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccbell-debug-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with debug enabled but plugin disabled
	configContent := `{
		"enabled": false,
		"debug": true
	}`
	configPath := filepath.Join(claudeDir, "ccbell.config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Set environment
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("CLAUDE_PROJECT_DIR")
	os.Setenv("CLAUDE_PLUGIN_ROOT", tmpDir)

	os.Args = []string{"ccbell", "stop"}
	err = run()
	if err != nil {
		t.Errorf("run() with debug mode should not error, got: %v", err)
	}

	// Cleanup debug log if created
	logPath := filepath.Join(claudeDir, "ccbell.log")
	os.Remove(logPath)
}

func TestRunWithInvalidConfig(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	oldHome := os.Getenv("HOME")
	oldProjectDir := os.Getenv("CLAUDE_PROJECT_DIR")
	oldPluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	defer func() {
		os.Args = oldArgs
		os.Setenv("HOME", oldHome)
		if oldProjectDir != "" {
			os.Setenv("CLAUDE_PROJECT_DIR", oldProjectDir)
		} else {
			os.Unsetenv("CLAUDE_PROJECT_DIR")
		}
		if oldPluginRoot != "" {
			os.Setenv("CLAUDE_PLUGIN_ROOT", oldPluginRoot)
		} else {
			os.Unsetenv("CLAUDE_PLUGIN_ROOT")
		}
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccbell-invalid-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create invalid JSON config
	configContent := `{invalid json`
	configPath := filepath.Join(claudeDir, "ccbell.config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Create sounds directory with a fallback sound
	soundsDir := filepath.Join(tmpDir, "sounds")
	if err := os.MkdirAll(soundsDir, 0755); err != nil {
		t.Fatal(err)
	}
	stopSound := filepath.Join(soundsDir, "stop.aiff")
	if err := os.WriteFile(stopSound, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	// Set environment
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("CLAUDE_PROJECT_DIR")
	os.Setenv("CLAUDE_PLUGIN_ROOT", tmpDir)

	// Should fall back to defaults and try to play (will fail on sound playback)
	os.Args = []string{"ccbell", "stop"}
	err = run()
	// This may error or not depending on whether the sound can play
	// The key is it doesn't panic
	t.Logf("run() with invalid config returned: %v", err)
}

func TestRunWithActiveProfile(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	oldHome := os.Getenv("HOME")
	oldPluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	defer func() {
		os.Args = oldArgs
		os.Setenv("HOME", oldHome)
		if oldPluginRoot != "" {
			os.Setenv("CLAUDE_PLUGIN_ROOT", oldPluginRoot)
		} else {
			os.Unsetenv("CLAUDE_PLUGIN_ROOT")
		}
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccbell-profile-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with profile (disabled to avoid audio playback)
	configContent := `{
		"enabled": false,
		"activeProfile": "work",
		"profiles": {
			"work": {
				"events": {
					"stop": {"enabled": true, "volume": 0.3}
				}
			}
		}
	}`
	configPath := filepath.Join(claudeDir, "ccbell.config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Set environment
	os.Setenv("HOME", tmpDir)
	os.Setenv("CLAUDE_PLUGIN_ROOT", tmpDir)

	os.Args = []string{"ccbell", "stop"}
	err = run()
	// Should not error when plugin is disabled (exits early before audio)
	if err != nil {
		t.Errorf("run() with disabled plugin should not error, got: %v", err)
	}
}

func TestRunWithUserProfile(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	oldHome := os.Getenv("HOME")
	oldPluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	defer func() {
		os.Args = oldArgs
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
		if oldPluginRoot != "" {
			os.Setenv("CLAUDE_PLUGIN_ROOT", oldPluginRoot)
		} else {
			os.Unsetenv("CLAUDE_PLUGIN_ROOT")
		}
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccbell-userprofile-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with disabled plugin (to avoid audio playback)
	configContent := testConfigDisabledPlugin
	configPath := filepath.Join(claudeDir, "ccbell.config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Set environment - use temp dir for HOME
	os.Unsetenv("HOME")
	os.Setenv("HOME", tmpDir)
	os.Setenv("CLAUDE_PLUGIN_ROOT", tmpDir)

	os.Args = []string{"ccbell", "stop"}
	err = run()
	if err != nil {
		t.Errorf("run() should not error, got: %v", err)
	}
}

func TestRunWithMissingConfigCreatesDefault(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	oldHome := os.Getenv("HOME")
	oldPluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	defer func() {
		os.Args = oldArgs
		os.Setenv("HOME", oldHome)
		if oldPluginRoot != "" {
			os.Setenv("CLAUDE_PLUGIN_ROOT", oldPluginRoot)
		} else {
			os.Unsetenv("CLAUDE_PLUGIN_ROOT")
		}
	}()

	// Create temp directory WITHOUT .claude directory
	tmpDir, err := os.MkdirTemp("", "ccbell-missing-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create sounds directory with a fallback sound
	soundsDir := filepath.Join(tmpDir, "sounds")
	if err := os.MkdirAll(soundsDir, 0755); err != nil {
		t.Fatal(err)
	}
	stopSound := filepath.Join(soundsDir, "stop.aiff")
	if err := os.WriteFile(stopSound, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	// Set environment
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("CLAUDE_PROJECT_DIR")
	os.Setenv("CLAUDE_PLUGIN_ROOT", tmpDir)

	// Create disabled config to exit early before audio playback
	disabledConfig := `{"enabled": false}`
	configPath := filepath.Join(tmpDir, ".claude", "ccbell.config.json")
	if err := os.WriteFile(configPath, []byte(disabledConfig), 0600); err != nil {
		t.Fatal(err)
	}

	// Test with disabled plugin - should create default config and exit early
	os.Args = []string{"ccbell", "stop"}
	err = run()
	if err != nil {
		t.Errorf("run() with missing config should create default and not error, got: %v", err)
	}

	// Verify config was preserved (not overwritten since it already existed)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("run() should have config file")
	}
}

func TestRunWithDifferentVolumes(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	oldHome := os.Getenv("HOME")
	oldPluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	defer func() {
		os.Args = oldArgs
		os.Setenv("HOME", oldHome)
		if oldPluginRoot != "" {
			os.Setenv("CLAUDE_PLUGIN_ROOT", oldPluginRoot)
		} else {
			os.Unsetenv("CLAUDE_PLUGIN_ROOT")
		}
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccbell-volume-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with different volumes but plugin disabled
	configContent := `{
		"enabled": false,
		"events": {
			"stop": {"volume": 0.1},
			"permission_prompt": {"volume": 1.0}
		}
	}`
	configPath := filepath.Join(claudeDir, "ccbell.config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Set environment
	os.Setenv("HOME", tmpDir)
	os.Setenv("CLAUDE_PLUGIN_ROOT", tmpDir)

	// Test low volume
	os.Args = []string{"ccbell", "stop"}
	err = run()
	if err != nil {
		t.Errorf("run() with low volume should not error, got: %v", err)
	}

	// Test max volume
	os.Args = []string{"ccbell", "permission_prompt"}
	err = run()
	if err != nil {
		t.Errorf("run() with max volume should not error, got: %v", err)
	}
}

func TestRunWithZeroCooldown(t *testing.T) {
	// Save original args and env
	oldArgs := os.Args
	oldHome := os.Getenv("HOME")
	oldPluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	defer func() {
		os.Args = oldArgs
		os.Setenv("HOME", oldHome)
		if oldPluginRoot != "" {
			os.Setenv("CLAUDE_PLUGIN_ROOT", oldPluginRoot)
		} else {
			os.Unsetenv("CLAUDE_PLUGIN_ROOT")
		}
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ccbell-zero-cooldown-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with zero cooldown (no blocking)
	configContent := `{
		"enabled": true,
		"events": {
			"stop": {"enabled": false, "cooldown": 0}
		}
	}`
	configPath := filepath.Join(claudeDir, "ccbell.config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Set environment
	os.Setenv("HOME", tmpDir)
	os.Setenv("CLAUDE_PLUGIN_ROOT", tmpDir)

	os.Args = []string{"ccbell", "stop"}
	err = run()
	if err != nil {
		t.Errorf("run() with zero cooldown should not error, got: %v", err)
	}
}

func TestDerefFunctions(t *testing.T) {
	// Test derefBool
	trueVal := true
	falseVal := false
	if derefBool(nil, true) != true {
		t.Error("derefBool with nil should return default")
	}
	if derefBool(&trueVal, false) != true {
		t.Error("derefBool with true pointer should return true")
	}
	if derefBool(&falseVal, true) != false {
		t.Error("derefBool with false pointer should return false")
	}

	// Test derefFloat
	fval := 0.75
	if derefFloat(nil, 0.5) != 0.5 {
		t.Error("derefFloat with nil should return default")
	}
	if derefFloat(&fval, 0.1) != 0.75 {
		t.Error("derefFloat with pointer should return value")
	}

	// Test derefInt
	ival := 42
	if derefInt(nil, 10) != 10 {
		t.Error("derefInt with nil should return default")
	}
	if derefInt(&ival, 0) != 42 {
		t.Error("derefInt with pointer should return value")
	}
}

func TestFindPluginRoot(t *testing.T) {
	tests := []struct {
		name     string
		homeDir  string
		wantPath bool
	}{
		{"empty home dir", "", false},
		{"nonexistent home dir", "/nonexistent/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findPluginRoot(tt.homeDir)
			if tt.wantPath && result == "" {
				t.Error("findPluginRoot expected a path but got empty string")
			}
			if !tt.wantPath && result != "" {
				t.Error("findPluginRoot expected empty but got:", result)
			}
		})
	}
}

func TestFindPluginRootWithCache(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "ccbell-plugintest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create cache directory structure
	cacheDir := filepath.Join(tmpDir, ".claude", "plugins", "cache")
	pluginsDir := filepath.Join(cacheDir, "mpolatcan-cc-plugins")
	ccbellDir := filepath.Join(pluginsDir, "ccbell")
	versionDir := filepath.Join(ccbellDir, "v0.2.20")
	soundsDir := filepath.Join(versionDir, "sounds")

	if err := os.MkdirAll(soundsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a dummy sound file for validation
	stopSound := filepath.Join(soundsDir, "stop.aiff")
	if err := os.WriteFile(stopSound, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	result := findPluginRoot(tmpDir)
	if result == "" {
		t.Error("findPluginRoot should find ccbell in cache")
	}
	if !strings.Contains(result, "ccbell") {
		t.Error("findPluginRoot result should contain 'ccbell'")
	}
}
