// Package audio provides cross-platform audio playback for ccbell.
package audio

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Platform represents the detected operating system.
type Platform string

// Platform constants represent the supported operating systems.
const (
	PlatformMacOS   Platform = "macos" // Apple macOS
	PlatformLinux   Platform = "linux" // Linux
	PlatformUnknown Platform = "unknown"
)

// linuxAudioPlayerNames is the list of audio players checked on Linux.
var linuxAudioPlayerNames = []string{"paplay", "aplay", "mpv", "ffplay"}

// getLinuxPlayerArgs returns arguments for a Linux audio player.
func getLinuxPlayerArgs(playerName, soundPath string, volume float64) []string {
	volPercent := int(volume * 100)
	switch playerName {
	case "paplay":
		return []string{soundPath}
	case "aplay":
		return []string{"-q", soundPath}
	case "mpv":
		return []string{"--really-quiet", fmt.Sprintf("--volume=%d", volPercent), soundPath}
	case "ffplay":
		return []string{"-nodisp", "-autoexit", "-volume", fmt.Sprintf("%d", volPercent), soundPath}
	default:
		return nil
	}
}

// bundledSoundNameRegex validates bundled sound names.
var bundledSoundNameRegex = regexp.MustCompile(`^[a-z_]+$`)

// Player handles audio playback.
type Player struct {
	platform   Platform
	pluginRoot string
}

// NewPlayer creates a new audio player.
func NewPlayer(pluginRoot string) *Player {
	return &Player{
		platform:   detectPlatform(),
		pluginRoot: pluginRoot,
	}
}

// detectPlatform determines the current platform.
func detectPlatform() Platform {
	switch runtime.GOOS {
	case "darwin":
		return PlatformMacOS
	case "linux":
		return PlatformLinux
	default:
		return PlatformUnknown
	}
}

// Play plays a sound file at the specified volume (0.0-1.0).
func (p *Player) Play(soundPath string, volume float64) error {
	if soundPath == "" {
		return errors.New("no sound path specified")
	}

	if _, err := os.Stat(soundPath); os.IsNotExist(err) {
		return fmt.Errorf("sound file not found: %s", soundPath)
	}

	switch p.platform {
	case PlatformMacOS:
		return p.playMacOS(soundPath, volume)
	case PlatformLinux:
		return p.playLinux(soundPath, volume)
	case PlatformUnknown:
		return fmt.Errorf("unsupported platform: %s", p.platform)
	default:
		return fmt.Errorf("unknown platform: %s", p.platform)
	}
}

// playMacOS uses afplay on macOS.
func (p *Player) playMacOS(soundPath string, volume float64) error {
	cmd := exec.Command("afplay", "-v", fmt.Sprintf("%.2f", volume), soundPath)
	return cmd.Start() // Non-blocking
}

// playLinux tries available audio players on Linux.
func (p *Player) playLinux(soundPath string, volume float64) error {
	for _, playerName := range linuxAudioPlayerNames {
		if _, err := exec.LookPath(playerName); err == nil {
			args := getLinuxPlayerArgs(playerName, soundPath, volume)
			cmd := exec.Command(playerName, args...)
			return cmd.Start() // Non-blocking
		}
	}

	return errors.New("no audio player found; install pulseaudio, alsa-utils, mpv, or ffmpeg")
}

// ResolveSoundPath resolves a sound specification to an absolute file path.
// Supported formats:
//   - bundled:stop (bundled with plugin)
//   - custom:/path/to/file.mp3
//   - /absolute/path/to/file.mp3
func (p *Player) ResolveSoundPath(soundSpec, eventType string) (string, error) {
	if soundSpec == "" {
		soundSpec = fmt.Sprintf("bundled:%s", eventType)
	}

	switch {
	case strings.HasPrefix(soundSpec, "bundled:"):
		return p.resolveBundledSound(strings.TrimPrefix(soundSpec, "bundled:"))

	case strings.HasPrefix(soundSpec, "custom:"):
		return p.resolveCustomSound(strings.TrimPrefix(soundSpec, "custom:"))

	default:
		// Direct path - apply same security checks as custom
		return p.resolveCustomSound(soundSpec)
	}
}

// resolveCustomSound resolves a custom sound path with security validation.
func (p *Player) resolveCustomSound(path string) (string, error) {
	// Security: must be absolute path
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("custom sound must be absolute path: %s", path)
	}

	// Security: no path traversal
	if strings.Contains(path, "..") {
		return "", errors.New("path traversal not allowed")
	}

	// Check file exists and is readable
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("custom sound not accessible: %s", path)
	}

	return path, nil
}

// resolveBundledSound resolves a bundled sound name.
// Uses os.Lstat to prevent symlink attacks.
func (p *Player) resolveBundledSound(name string) (string, error) {
	// Validate name (lowercase letters and underscores only)
	if !bundledSoundNameRegex.MatchString(name) {
		return "", fmt.Errorf("invalid bundled sound name: %s", name)
	}

	path := filepath.Join(p.pluginRoot, "sounds", name+".aiff")
	// Use Lstat to detect symlinks and prevent path traversal via symlinks
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("bundled sound not found: %s", name)
	}

	return path, nil
}

// GetFallbackPath returns a fallback sound path for the event type.
// Uses Lstat to prevent symlink attacks.
func (p *Player) GetFallbackPath(eventType string) string {
	// Try bundled sound for this event
	path := filepath.Join(p.pluginRoot, "sounds", eventType+".aiff")
	if _, err := os.Lstat(path); err == nil {
		return path
	}

	// Try bundled stop sound (always present)
	path = filepath.Join(p.pluginRoot, "sounds", "stop.aiff")
	if _, err := os.Lstat(path); err == nil {
		return path
	}

	return ""
}

// Platform returns the detected platform.
func (p *Player) Platform() Platform {
	return p.platform
}

// HasAudioPlayer checks if an audio player is available.
func (p *Player) HasAudioPlayer() bool {
	switch p.platform {
	case PlatformMacOS:
		_, err := exec.LookPath("afplay")
		return err == nil
	case PlatformLinux:
		for _, player := range linuxAudioPlayerNames {
			if _, err := exec.LookPath(player); err == nil {
				return true
			}
		}
		return false
	case PlatformUnknown:
		return false
	default:
		return false
	}
}
