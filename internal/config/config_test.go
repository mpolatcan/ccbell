package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateEventType(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		wantErr   bool
	}{
		{"valid stop", "stop", false},
		{"valid permission_prompt", "permission_prompt", false},
		{"valid subagent", "subagent", false},
		{"invalid event", "invalid_event", true},
		{"injection attempt", "stop; echo pwned", true},
		{"uppercase", "STOP", true},
		{"empty", "", true},
		{"path traversal", "../etc/passwd", true},
		{"numeric", "123", true},
		{"special chars", "stop$test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEventType(tt.eventType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEventType(%q) error = %v, wantErr %v", tt.eventType, err, tt.wantErr)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "default config is valid",
			config:  Default(),
			wantErr: false,
		},
		{
			name: "invalid quiet hours start format",
			config: &Config{
				QuietHours: &QuietHours{Start: "25:00", End: "07:00"},
			},
			wantErr: true,
		},
		{
			name: "invalid quiet hours end format",
			config: &Config{
				QuietHours: &QuietHours{Start: "22:00", End: "7:00"},
			},
			wantErr: true,
		},
		{
			name: "valid quiet hours",
			config: &Config{
				QuietHours: &QuietHours{Start: "22:00", End: "07:00"},
			},
			wantErr: false,
		},
		{
			name: "volume out of range",
			config: &Config{
				Events: map[string]*Event{
					"stop": {Volume: ptrFloat(1.5)},
				},
			},
			wantErr: true,
		},
		{
			name: "negative volume",
			config: &Config{
				Events: map[string]*Event{
					"stop": {Volume: ptrFloat(-0.5)},
				},
			},
			wantErr: true,
		},
		{
			name: "negative cooldown",
			config: &Config{
				Events: map[string]*Event{
					"stop": {Cooldown: ptrInt(-5)},
				},
			},
			wantErr: true,
		},
		{
			name: "unknown event type",
			config: &Config{
				Events: map[string]*Event{
					"unknown_event": {Volume: ptrFloat(0.5)},
				},
			},
			wantErr: true,
		},
		{
			name: "activeProfile not found",
			config: &Config{
				ActiveProfile: "nonexistent",
				Profiles:      map[string]*Profile{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEventConfig(t *testing.T) {
	cfg := &Config{
		ActiveProfile: "work",
		Events: map[string]*Event{
			"stop": {Enabled: ptrBool(true), Sound: "bundled:stop", Volume: ptrFloat(0.5), Cooldown: ptrInt(5)},
		},
		Profiles: map[string]*Profile{
			"work": {
				Events: map[string]*Event{
					"stop": {Sound: "bundled:subagent", Volume: ptrFloat(0.3)},
				},
			},
			"silent": {
				Events: map[string]*Event{
					"stop": {Enabled: ptrBool(false)},
				},
			},
		},
	}

	t.Run("profile overrides base config", func(t *testing.T) {
		eventCfg := cfg.GetEventConfig("stop")
		if eventCfg.Sound != "bundled:subagent" {
			t.Errorf("expected sound 'bundled:subagent', got '%s'", eventCfg.Sound)
		}
		if *eventCfg.Volume != 0.3 {
			t.Errorf("expected volume 0.3, got %f", *eventCfg.Volume)
		}
		// Cooldown should be inherited from base
		if *eventCfg.Cooldown != 5 {
			t.Errorf("expected cooldown 5, got %d", *eventCfg.Cooldown)
		}
	})

	t.Run("default profile uses base config", func(t *testing.T) {
		cfg.ActiveProfile = "default"
		eventCfg := cfg.GetEventConfig("stop")
		if eventCfg.Sound != "bundled:stop" {
			t.Errorf("expected sound 'bundled:stop', got '%s'", eventCfg.Sound)
		}
	})

	t.Run("undefined event returns defaults", func(t *testing.T) {
		eventCfg := cfg.GetEventConfig("permission_prompt")
		if eventCfg.Sound != "bundled:permission_prompt" {
			t.Errorf("expected default sound, got '%s'", eventCfg.Sound)
		}
	})
}

func TestLoadConfig(t *testing.T) {
	// Create temp directory for test configs
	tempDir, err := os.MkdirTemp("", "ccbell-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create .claude directory
	claudeDir := filepath.Join(tempDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("loads valid config", func(t *testing.T) {
		configPath := filepath.Join(claudeDir, "ccbell.config.json")
		configContent := `{
			"enabled": true,
			"debug": true,
			"activeProfile": "work",
			"events": {
				"stop": {"sound": "bundled:stop", "volume": 0.7}
			},
			"profiles": {
				"work": {
					"events": {
						"stop": {"volume": 0.3}
					}
				}
			}
		}`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, path, err := Load(tempDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if path != configPath {
			t.Errorf("expected path %s, got %s", configPath, path)
		}
		if !cfg.Debug {
			t.Error("expected debug to be true")
		}
		if cfg.ActiveProfile != "work" {
			t.Errorf("expected activeProfile 'work', got '%s'", cfg.ActiveProfile)
		}
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		configPath := filepath.Join(claudeDir, "ccbell.config.json")
		if err := os.WriteFile(configPath, []byte("{invalid json}"), 0644); err != nil {
			t.Fatal(err)
		}

		_, _, err := Load(tempDir)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("returns defaults when no config exists", func(t *testing.T) {
		os.Remove(filepath.Join(claudeDir, "ccbell.config.json"))

		cfg, path, err := Load(tempDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if path != "" {
			t.Errorf("expected empty path, got %s", path)
		}
		if !cfg.Enabled {
			t.Error("expected default enabled to be true")
		}
	})
}
