// Package config handles ccbell configuration loading and validation.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Config represents the full ccbell configuration.
type Config struct {
	Enabled       bool                `json:"enabled"`
	Debug         bool                `json:"debug"`
	ActiveProfile string              `json:"activeProfile"`
	QuietHours    *QuietHours         `json:"quietHours,omitempty"`
	Events        map[string]*Event   `json:"events,omitempty"`
	Profiles      map[string]*Profile `json:"profiles,omitempty"`
}

// defaultProfileName is the name of the default profile.
const defaultProfileName = "default"

// QuietHours represents do-not-disturb time window.
type QuietHours struct {
	Start string `json:"start"` // HH:MM format
	End   string `json:"end"`   // HH:MM format
}

// Event represents configuration for a single event type.
type Event struct {
	Enabled  *bool    `json:"enabled,omitempty"`
	Sound    string   `json:"sound,omitempty"`
	Volume   *float64 `json:"volume,omitempty"`
	Cooldown *int     `json:"cooldown,omitempty"`
}

// Profile represents a named configuration preset.
type Profile struct {
	Events map[string]*Event `json:"events,omitempty"`
}

// ValidEvents is the whitelist of allowed event types.
var ValidEvents = map[string]bool{
	"stop":              true,
	"permission_prompt": true,
	"idle_prompt":       true,
	"subagent":          true,
}

// timeFormatRegex validates HH:MM format.
var timeFormatRegex = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

// eventTypeRegex validates event type format (lowercase letters and underscores).
var eventTypeRegex = regexp.MustCompile(`^[a-z_]+$`)

// Helper functions for creating pointers to primitives.
func ptrBool(v bool) *bool        { return &v }
func ptrFloat(v float64) *float64 { return &v }
func ptrInt(v int) *int           { return &v }

// Default returns a Config with default values.
func Default() *Config {
	return &Config{
		Enabled:       true,
		Debug:         false,
		ActiveProfile: "default",
		Events: map[string]*Event{
			"stop":              {Enabled: ptrBool(true), Sound: "bundled:stop", Volume: ptrFloat(0.5), Cooldown: ptrInt(0)},
			"permission_prompt": {Enabled: ptrBool(true), Sound: "bundled:permission_prompt", Volume: ptrFloat(0.7), Cooldown: ptrInt(0)},
			"idle_prompt":       {Enabled: ptrBool(true), Sound: "bundled:idle_prompt", Volume: ptrFloat(0.5), Cooldown: ptrInt(0)},
			"subagent":          {Enabled: ptrBool(true), Sound: "bundled:subagent", Volume: ptrFloat(0.5), Cooldown: ptrInt(0)},
		},
	}
}

// Load reads configuration from file, falling back to defaults.
// It only checks the global config at ~/.claude/ccbell.config.json.
func Load(homeDir string) (*Config, string, error) {
	cfg := Default()
	configPath := ""

	// Load global config
	if homeDir != "" {
		globalConfig := filepath.Join(homeDir, ".claude", "ccbell.config.json")
		if data, err := os.ReadFile(globalConfig); err == nil {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, "", fmt.Errorf("invalid JSON in %s: %w", globalConfig, err)
			}
			configPath = globalConfig
		}
	}

	// Validate after loading
	if err := cfg.Validate(); err != nil {
		return nil, configPath, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, configPath, nil
}

// EnsureConfig creates default config file if it doesn't exist.
func EnsureConfig(homeDir string) error {
	configPath := filepath.Join(homeDir, ".claude", "ccbell.config.json")
	if _, err := os.Stat(configPath); err == nil {
		return nil // Already exists
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(Default(), "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	// Validate quiet hours format
	if c.QuietHours != nil {
		if c.QuietHours.Start != "" && !timeFormatRegex.MatchString(c.QuietHours.Start) {
			return fmt.Errorf("invalid quietHours.start format: %s (expected HH:MM)", c.QuietHours.Start)
		}
		if c.QuietHours.End != "" && !timeFormatRegex.MatchString(c.QuietHours.End) {
			return fmt.Errorf("invalid quietHours.end format: %s (expected HH:MM)", c.QuietHours.End)
		}
	}

	// Validate activeProfile exists in Profiles (if not default)
	if c.ActiveProfile != "" && c.ActiveProfile != defaultProfileName {
		if _, ok := c.Profiles[c.ActiveProfile]; !ok {
			return fmt.Errorf("activeProfile %q not found in profiles", c.ActiveProfile)
		}
	}

	// Validate event configs
	for name, event := range c.Events {
		if !ValidEvents[name] {
			return fmt.Errorf("unknown event type: %s", name)
		}
		if event.Volume != nil && (*event.Volume < 0 || *event.Volume > 1) {
			return fmt.Errorf("event %s: volume must be 0.0-1.0, got %f", name, *event.Volume)
		}
		if event.Cooldown != nil && *event.Cooldown < 0 {
			return fmt.Errorf("event %s: cooldown cannot be negative", name)
		}
	}

	// Validate profile event configs
	for profileName, profile := range c.Profiles {
		for eventName, event := range profile.Events {
			if !ValidEvents[eventName] {
				return fmt.Errorf("profile %s: unknown event type: %s", profileName, eventName)
			}
			if event.Volume != nil && (*event.Volume < 0 || *event.Volume > 1) {
				return fmt.Errorf("profile %s, event %s: volume must be 0.0-1.0", profileName, eventName)
			}
			if event.Cooldown != nil && *event.Cooldown < 0 {
				return fmt.Errorf("profile %s, event %s: cooldown cannot be negative", profileName, eventName)
			}
		}
	}

	return nil
}

// GetEventConfig returns the effective configuration for an event,
// considering the active profile.
func (c *Config) GetEventConfig(eventType string) *Event {
	// Start with defaults
	result := &Event{
		Enabled:  ptrBool(true),
		Sound:    fmt.Sprintf("bundled:%s", eventType),
		Volume:   ptrFloat(0.5),
		Cooldown: ptrInt(0),
	}

	// Apply base event config
	if baseEvent, ok := c.Events[eventType]; ok {
		mergeEvent(result, baseEvent)
	}

	// Apply profile overrides (if not default profile)
	if c.ActiveProfile != "" && c.ActiveProfile != "default" {
		if profile, ok := c.Profiles[c.ActiveProfile]; ok {
			if profileEvent, ok := profile.Events[eventType]; ok {
				mergeEvent(result, profileEvent)
			}
		}
	}

	return result
}

// mergeEvent applies set values from src to dst.
// Nil values in src are treated as "not set" and don't override dst.
func mergeEvent(dst, src *Event) {
	if src.Enabled != nil {
		dst.Enabled = src.Enabled
	}
	if src.Sound != "" {
		dst.Sound = src.Sound
	}
	if src.Volume != nil {
		dst.Volume = src.Volume
	}
	if src.Cooldown != nil {
		dst.Cooldown = src.Cooldown
	}
}

// ValidateEventType returns an error if the event type is invalid.
func ValidateEventType(eventType string) error {
	// Check format (alphanumeric and underscore only)
	if !eventTypeRegex.MatchString(eventType) {
		return errors.New("invalid event type format: must be lowercase letters and underscores only")
	}

	// Check whitelist
	if !ValidEvents[eventType] {
		valid := make([]string, 0, len(ValidEvents))
		for k := range ValidEvents {
			valid = append(valid, k)
		}
		return fmt.Errorf("unknown event type: %s (valid: %v)", eventType, valid)
	}

	return nil
}
