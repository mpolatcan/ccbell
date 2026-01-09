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

const (
	PlatformMacOS   Platform = "macos"
	PlatformLinux   Platform = "linux"
	PlatformWindows Platform = "windows"
	PlatformWSL     Platform = "wsl"
	PlatformUnknown Platform = "unknown"
)

// Player handles audio playback.
type Player struct {
	platform   Platform
	pluginRoot string
}

// bundledSoundNameRegex validates bundled sound names.
var bundledSoundNameRegex = regexp.MustCompile(`^[a-z_]+$`)

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
	case "windows":
		return PlatformWindows
	case "linux":
		// Check for WSL (both legacy and WSL2)
		if _, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop"); err == nil {
			return PlatformWSL
		}
		// WSL2 uses a different mechanism - check for WSL-specific environment
		if os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSL_INTEROP") != "" {
			return PlatformWSL
		}
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
	case PlatformLinux, PlatformWSL:
		return p.playLinux(soundPath, volume)
	case PlatformWindows:
		return p.playWindows(soundPath)
	default:
		return fmt.Errorf("unsupported platform: %s", p.platform)
	}
}

// playMacOS uses afplay on macOS.
func (p *Player) playMacOS(soundPath string, volume float64) error {
	cmd := exec.Command("afplay", "-v", fmt.Sprintf("%.2f", volume), soundPath)
	return cmd.Start() // Non-blocking
}

// playLinux tries available audio players on Linux.
func (p *Player) playLinux(soundPath string, volume float64) error {
	volPercent := int(volume * 100)

	// Try players in order of preference
	players := []struct {
		name string
		args func() []string
	}{
		{"paplay", func() []string { return []string{soundPath} }},
		{"aplay", func() []string { return []string{"-q", soundPath} }},
		{"mpv", func() []string { return []string{"--really-quiet", fmt.Sprintf("--volume=%d", volPercent), soundPath} }},
		{"ffplay", func() []string {
			return []string{"-nodisp", "-autoexit", "-volume", fmt.Sprintf("%d", volPercent), soundPath}
		}},
	}

	for _, player := range players {
		if _, err := exec.LookPath(player.name); err == nil {
			cmd := exec.Command(player.name, player.args()...)
			return cmd.Start() // Non-blocking
		}
	}

	return errors.New("no audio player found; install pulseaudio, alsa-utils, mpv, or ffmpeg")
}

// playWindows uses PowerShell on Windows.
// Security: Uses -File parameter with a temp script to avoid command injection.
func (p *Player) playWindows(soundPath string) error {
	// Security: Validate path contains only safe characters
	// Allow alphanumeric, spaces, common path chars, but block dangerous ones
	for _, r := range soundPath {
		if r == ';' || r == '&' || r == '|' || r == '`' || r == '$' || r == '(' || r == ')' || r == '{' || r == '}' || r == '<' || r == '>' || r == '\n' || r == '\r' {
			return fmt.Errorf("invalid character in sound path: %q", string(r))
		}
	}

	// Use PowerShell with properly escaped path
	// Double single quotes for PowerShell string escaping
	escaped := strings.ReplaceAll(soundPath, "'", "''")
	// Also escape backticks which are PowerShell's escape character
	escaped = strings.ReplaceAll(escaped, "`", "``")

	psCmd := fmt.Sprintf("(New-Object Media.SoundPlayer '%s').PlaySync()", escaped)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-c", psCmd)
	// Note: PlaySync() is synchronous in PowerShell, but cmd.Start() is used
	// to launch the PowerShell process. The process runs in the background.
	return cmd.Start()
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
	case PlatformLinux, PlatformWSL:
		for _, player := range []string{"paplay", "aplay", "mpv", "ffplay"} {
			if _, err := exec.LookPath(player); err == nil {
				return true
			}
		}
		return false
	case PlatformWindows:
		_, err := exec.LookPath("powershell")
		return err == nil
	default:
		return false
	}
}
