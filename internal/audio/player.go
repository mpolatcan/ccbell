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

// Package managers and their install commands.
var packageManagers = map[string]string{
	"apt-get": "sudo apt-get update && sudo apt-get install -y",
	"dnf":     "sudo dnf install -y",
	"yum":     "sudo yum install -y",
	"pacman":  "sudo pacman -S --noconfirm",
	"zypper":  "sudo zypper install -y",
	"apk":     "sudo apk add --no-cache",
	"emerge":  "sudo emerge --ask",
}

// Packages to install for each audio player.
var playerPackages = map[string]string{
	"mpv":     "mpv",
	"ffplay":  "ffmpeg",
	"paplay":  "pulseaudio-utils",
	"aplay":   "alsa-utils",
}

// Platform represents the detected operating system.
type Platform string

// Platform constants represent the supported operating systems.
const (
	PlatformMacOS   Platform = "macos" // Apple macOS
	PlatformLinux   Platform = "linux" // Linux
	PlatformUnknown Platform = "unknown"
)

// linuxAudioPlayerNames is the list of audio players checked on Linux (priority order).
var linuxAudioPlayerNames = []string{"mpv", "paplay", "aplay", "ffplay"}

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
	homeDir    string
}

// NewPlayer creates a new audio player.
func NewPlayer(pluginRoot string) *Player {
	return &Player{
		platform:   detectPlatform(),
		pluginRoot: pluginRoot,
	}
}

// NewPlayerWithHome creates a new audio player with home directory for pack support.
func NewPlayerWithHome(pluginRoot, homeDir string) *Player {
	return &Player{
		platform:   detectPlatform(),
		pluginRoot: pluginRoot,
		homeDir:    homeDir,
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
//   - pack:pack_id:sound_file (sound from a pack)
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

	case strings.HasPrefix(soundSpec, "pack:"):
		return p.resolvePackSound(strings.TrimPrefix(soundSpec, "pack:"))

	default:
		// Direct path - apply same security checks as custom
		return p.resolveCustomSound(soundSpec)
	}
}

// resolvePackSound resolves a pack sound (format: pack_id:sound_file).
func (p *Player) resolvePackSound(spec string) (string, error) {
	// Parse pack_id:sound_file format
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid pack sound format: %s (expected pack_id:sound_file)", spec)
	}

	packID := parts[0]
	soundFile := parts[1]

	// Validate pack ID
	if !bundledSoundNameRegex.MatchString(packID) {
		return "", fmt.Errorf("invalid pack ID: %s", packID)
	}

	// Validate sound file name (basic check for path traversal)
	if strings.Contains(soundFile, "..") || strings.Contains(soundFile, "/") {
		return "", fmt.Errorf("invalid sound file name: %s", soundFile)
	}

	// Check home directory is set
	if p.homeDir == "" {
		return "", fmt.Errorf("home directory not set for pack sounds")
	}

	// Resolve pack directory
	packDir := filepath.Join(p.homeDir, ".claude", "ccbell", "packs", packID)
	path := filepath.Join(packDir, soundFile)

	// Check file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("pack sound not found: pack=%s, sound=%s", packID, soundFile)
	}

	return path, nil
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

// findPackageManager detects available package manager.
func findPackageManager() string {
	for pm := range packageManagers {
		if _, err := exec.LookPath(pm); err == nil {
			return pm
		}
	}
	return ""
}

// installAudioPlayer attempts to install the specified audio player.
func installAudioPlayer(player string) error {
	pm := findPackageManager()
	if pm == "" {
		return errors.New("no package manager found")
	}

	cmdStr := packageManagers[pm]
	pkg := playerPackages[player]
	if pkg == "" {
		return fmt.Errorf("unknown player: %s", player)
	}

	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", cmdStr, pkg))
	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Run()
}

// EnsureAudioPlayer finds or installs an audio player. Returns the player name and error.
func (p *Player) EnsureAudioPlayer() (string, error) {
	// Already have a player?
	for _, player := range linuxAudioPlayerNames {
		if _, err := exec.LookPath(player); err == nil {
			return player, nil
		}
	}

	// Try to install
	for _, player := range linuxAudioPlayerNames {
		if err := installAudioPlayer(player); err == nil {
			if _, err := exec.LookPath(player); err == nil {
				return player, nil
			}
		}
	}

	return "", errors.New("no audio player found; install mpv, ffmpeg, pulseaudio-utils, or alsa-utils")
}
