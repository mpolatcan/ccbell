// ccbell - Sound notification hook for Claude Code
//
// Usage: ccbell <event_type>
// Event types: stop, permission_prompt, idle_prompt, subagent
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mpolatcan/ccbell/internal/audio"
	"github.com/mpolatcan/ccbell/internal/config"
	"github.com/mpolatcan/ccbell/internal/logger"
	"github.com/mpolatcan/ccbell/internal/state"
)

func derefBool(ptr *bool, defaultVal bool) bool {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func derefFloat(ptr *float64, defaultVal float64) float64 {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func derefInt(ptr *int, defaultVal int) int {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Build-time variables (set via -ldflags).
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

// findPluginRoot searches for the ccbell plugin in the plugins cache directory.
// It supports any marketplace path by scanning for directories named "ccbell".
func findPluginRoot(homeDir string) string {
	cacheDir := filepath.Join(homeDir, ".claude", "plugins", "cache")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return ""
	}

	// Find the ccbell plugin directory in any marketplace subdirectory
	var ccbellPath string
	filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip entries with errors
		}
		if info.IsDir() && path != cacheDir {
			// Check if this is a ccbell directory
			if info.Name() == "ccbell" {
				ccbellPath = path
				return filepath.SkipDir // Found it, stop walking
			}
		}
		return nil
	})

	if ccbellPath == "" {
		return ""
	}

	// Find the latest version subdirectory
	var latestVersion string
	filepath.Walk(ccbellPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && path != ccbellPath {
			// Check if it's a version directory (semver format: vX.Y.Z or X.Y.Z)
			name := info.Name()
			if strings.HasPrefix(name, "v") || (len(name) > 0 && name[0] >= '0' && name[0] <= '9') {
				// This is likely a version directory
				if latestVersion == "" || name > latestVersion {
					latestVersion = name
				}
			}
		}
		return nil
	})

	if latestVersion != "" {
		return filepath.Join(ccbellPath, latestVersion)
	}
	return ccbellPath
}

func main() {
	var exitCode int
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "PANIC: %v\n", r)
			exitCode = 2
		}
		os.Exit(exitCode)
	}()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		exitCode = 1
	}
}

func run() error {
	// === Get event type from args ===
	eventType := "stop"
	if len(os.Args) > 1 {
		eventType = os.Args[1]
	}

	// Handle special commands
	if eventType == "--version" || eventType == "-v" {
		fmt.Printf("ccbell %s (commit: %s, built: %s)\n", version, commit, buildDate)
		return nil
	}
	if eventType == "--help" || eventType == "-h" {
		printUsage()
		return nil
	}

	// === Validate event type ===
	if err := config.ValidateEventType(eventType); err != nil {
		return err
	}

	// === Drain stdin (hooks may send data) ===
	// Non-blocking read to prevent hanging. The stdin is drained in a separate
	// goroutine since this is a short-lived process.
	go func() {
		_, _ = io.Copy(io.Discard, os.Stdin)
	}()

	// === Environment setup ===
	homeDir := os.Getenv("HOME")
	pluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	if pluginRoot == "" {
		pluginRoot = findPluginRoot(homeDir)
	}

	// === Ensure config exists ===
	if err := config.EnsureConfig(homeDir); err != nil {
		fmt.Fprintf(os.Stderr, "ccbell: Warning: could not create config: %v\n", err)
	}

	// === Load configuration ===
	cfg, configPath, configErr := config.Load(homeDir)
	if configErr != nil {
		// Config error shouldn't be fatal - use defaults
		cfg = config.Default()
		configPath = "(default - config load failed)"
	}

	// === Initialize logger ===
	log := logger.New(cfg.Debug, homeDir)
	log.Debug("=== ccbell triggered: event=%s ===", eventType)
	log.Debug("Version: %s, Config: %s", version, configPath)

	// Log config error if any (after logger is initialized)
	if configErr != nil {
		log.Debug("Config load error (using defaults): %v", configErr)
		// Also warn to stderr so user knows their config is broken
		fmt.Fprintf(os.Stderr, "ccbell: config error, using defaults: %v\n", configErr)
	}
	log.Debug("Plugin root: %s", pluginRoot)

	// === Check global enable ===
	if !cfg.Enabled {
		log.Debug("Plugin disabled globally, exiting")
		return nil
	}

	// === Get event configuration ===
	eventCfg := cfg.GetEventConfig(eventType)
	log.Debug("Active profile: %s", cfg.ActiveProfile)
	log.Debug("Event config: enabled=%v, sound=%s, volume=%.2f, cooldown=%d",
		derefBool(eventCfg.Enabled, true), eventCfg.Sound, derefFloat(eventCfg.Volume, 0.5), derefInt(eventCfg.Cooldown, 0))

	// === Check event enable ===
	if !derefBool(eventCfg.Enabled, true) {
		log.Debug("Event '%s' is disabled, exiting", eventType)
		return nil
	}

	// === Check quiet hours ===
	if cfg.IsInQuietHours() {
		log.Debug("In quiet hours (%s-%s), suppressing notification",
			cfg.QuietHours.Start, cfg.QuietHours.End)
		return nil
	}

	// === Check cooldown ===
	stateManager := state.NewManager(homeDir)
	inCooldown, err := stateManager.CheckCooldown(eventType, derefInt(eventCfg.Cooldown, 0))
	if err != nil {
		log.Debug("Cooldown check error: %v, proceeding with notification", err)
	} else if inCooldown {
		log.Debug("In cooldown period (%ds), suppressing notification", derefInt(eventCfg.Cooldown, 0))
		return nil
	}

	log.Debug("All checks passed, proceeding to play sound")

	// === Resolve sound path ===
	player := audio.NewPlayer(pluginRoot)
	log.Debug("Detected platform: %s", player.Platform())

	// === Ensure audio player is available ===
	if player.Platform() == audio.PlatformLinux {
		audioPlayer, err := player.EnsureAudioPlayer()
		if err != nil {
			log.Debug("Audio player check failed: %v", err)
			return fmt.Errorf("no audio player available: %w", err)
		}
		log.Debug("Using audio player: %s", audioPlayer)
	}

	soundPath, err := player.ResolveSoundPath(eventCfg.Sound, eventType)
	if err != nil {
		log.Debug("Sound resolution failed: %v, trying fallbacks", err)
		soundPath = player.GetFallbackPath(eventType)
		if soundPath == "" {
			return fmt.Errorf("no playable sound found")
		}
	}
	log.Debug("Final sound path: %s", soundPath)

	// === Play sound ===
	if err := player.Play(soundPath, derefFloat(eventCfg.Volume, 0.5)); err != nil {
		log.Debug("Sound playback failed: %v", err)
		return fmt.Errorf("sound playback failed: %w", err)
	}

	log.Debug("Sound playback initiated successfully")
	log.Debug("=== ccbell completed ===")

	return nil
}

func printUsage() {
	fmt.Println(`ccbell - Sound notifications for Claude Code

USAGE:
    ccbell <event_type>
    ccbell [OPTIONS]

EVENT TYPES:
    stop              Claude finished responding
    permission_prompt Claude needs your permission
    idle_prompt       Claude is waiting for input
    subagent          A background agent completed

OPTIONS:
    -h, --help        Show this help message
    -v, --version     Show version information

CONFIGURATION:
    Global config:  ~/.claude/ccbell.config.json

SOUND FORMATS:
    bundled:stop         Bundled with plugin
    bundled:permission_prompt
    bundled:idle_prompt
    bundled:subagent
    custom:/path/to.mp3  Custom audio file

ENVIRONMENT:
    CLAUDE_PLUGIN_ROOT   Plugin installation directory

For more information, visit: https://github.com/mpolatcan/ccbell`)
}
