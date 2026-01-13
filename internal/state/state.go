// Package state manages cooldown state with atomic file operations.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// FileMode is the permission mode for state files.
	FileMode = 0600
)

// State represents the cooldown state.
type State struct {
	LastTrigger map[string]int64 `json:"lastTrigger"`
}

// Manager handles state file operations.
type Manager struct {
	filePath string
	mu       sync.Mutex
}

// NewManager creates a new state manager.
func NewManager(homeDir string) *Manager {
	statePath := ""
	if homeDir != "" {
		statePath = filepath.Join(homeDir, ".claude", "ccbell.state")
	}

	return &Manager{
		filePath: statePath,
	}
}

// CheckCooldown checks if an event is in cooldown period.
// Returns true if in cooldown (should skip notification), false otherwise.
// Also updates the last trigger time if not in cooldown.
func (m *Manager) CheckCooldown(eventType string, cooldownSecs int) (bool, error) {
	if m.filePath == "" || cooldownSecs <= 0 {
		return false, nil // No cooldown configured
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.load()
	if err != nil {
		// If we can't load state, assume not in cooldown but log the error
		state = &State{LastTrigger: make(map[string]int64)}
	}

	currentTime := time.Now().Unix()
	lastTrigger := state.LastTrigger[eventType]
	elapsed := currentTime - lastTrigger

	if elapsed < int64(cooldownSecs) {
		return true, nil // In cooldown
	}

	// Update last trigger time
	state.LastTrigger[eventType] = currentTime
	if err := m.save(state); err != nil {
		return false, fmt.Errorf("failed to save state: %w", err)
	}

	return false, nil
}

// load reads the state file.
func (m *Manager) load() (*State, error) {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{LastTrigger: make(map[string]int64)}, nil
		}
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		// Corrupted state file - start fresh
		return &State{LastTrigger: make(map[string]int64)}, nil
	}

	if state.LastTrigger == nil {
		state.LastTrigger = make(map[string]int64)
	}

	return &state, nil
}

// save writes the state file atomically.
func (m *Manager) save(state *State) error {
	// Ensure directory exists
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n') // Trailing newline

	// Write to temp file first (atomic)
	tempFile, err := os.CreateTemp(dir, "ccbell.state.*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Clean up temp file on any error
	defer func() {
		if tempPath != "" {
			os.Remove(tempPath)
		}
	}()

	// Set permissions before writing content
	if err := tempFile.Chmod(FileMode); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to set temp file permissions: %w", err)
	}

	// Write content
	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Close before rename
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, m.filePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	tempPath = "" // Prevent cleanup of renamed file
	return nil
}

// Clear removes the state file.
func (m *Manager) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.Remove(m.filePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
