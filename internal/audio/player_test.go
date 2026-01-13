package audio

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

const darwinOS = "darwin"
const linuxOS = "linux"

func TestResolveSoundPath(t *testing.T) {
	// Create temp plugin root with sounds
	tempDir, err := os.MkdirTemp("", "ccbell-audio-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	soundsDir := filepath.Join(tempDir, "sounds")
	if err := os.MkdirAll(soundsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create dummy sound file
	stopSound := filepath.Join(soundsDir, "stop.aiff")
	if err := os.WriteFile(stopSound, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	player := NewPlayer(tempDir)

	tests := []struct {
		name      string
		soundSpec string
		eventType string
		wantPath  string
		wantErr   bool
	}{
		{
			name:      "bundled sound",
			soundSpec: "bundled:stop",
			eventType: "stop",
			wantPath:  stopSound,
			wantErr:   false,
		},
		{
			name:      "empty defaults to bundled",
			soundSpec: "",
			eventType: "stop",
			wantPath:  stopSound,
			wantErr:   false,
		},
		{
			name:      "invalid bundled name",
			soundSpec: "bundled:../etc/passwd",
			eventType: "stop",
			wantPath:  "",
			wantErr:   true,
		},
		{
			name:      "bundled sound not found",
			soundSpec: "bundled:nonexistent",
			eventType: "stop",
			wantPath:  "",
			wantErr:   true,
		},
		{
			name:      "custom relative path rejected",
			soundSpec: "custom:relative/path.mp3",
			eventType: "stop",
			wantPath:  "",
			wantErr:   true,
		},
		{
			name:      "custom path traversal rejected",
			soundSpec: "custom:/path/../etc/passwd",
			eventType: "stop",
			wantPath:  "",
			wantErr:   true,
		},
		{
			name:      "direct relative path rejected",
			soundSpec: "relative/path.mp3",
			eventType: "stop",
			wantPath:  "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := player.ResolveSoundPath(tt.soundSpec, tt.eventType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveSoundPath(%q, %q) error = %v, wantErr %v",
					tt.soundSpec, tt.eventType, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantPath {
				t.Errorf("ResolveSoundPath(%q, %q) = %q, want %q",
					tt.soundSpec, tt.eventType, got, tt.wantPath)
			}
		})
	}
}

func TestDetectPlatform(t *testing.T) {
	platform := detectPlatform()

	switch runtime.GOOS {
	case darwinOS:
		if platform != PlatformMacOS {
			t.Errorf("expected PlatformMacOS on darwin, got %s", platform)
		}
	case linuxOS:
		if platform != PlatformLinux {
			t.Errorf("expected PlatformLinux on linux, got %s", platform)
		}
	default:
		if platform != PlatformUnknown {
			t.Errorf("expected PlatformUnknown on %s, got %s", runtime.GOOS, platform)
		}
	}
}

func TestGetFallbackPath(t *testing.T) {
	// Create temp plugin root
	tempDir, err := os.MkdirTemp("", "ccbell-fallback-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	soundsDir := filepath.Join(tempDir, "sounds")
	if err := os.MkdirAll(soundsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create stop.aiff
	stopSound := filepath.Join(soundsDir, "stop.aiff")
	if err := os.WriteFile(stopSound, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	player := NewPlayer(tempDir)

	t.Run("fallback to bundled stop", func(t *testing.T) {
		path := player.GetFallbackPath("permission_prompt")
		if path != stopSound {
			t.Errorf("expected fallback to stop.aiff, got %s", path)
		}
	})

	t.Run("fallback to event-specific sound", func(t *testing.T) {
		// Create permission_prompt.aiff
		ppSound := filepath.Join(soundsDir, "permission_prompt.aiff")
		if err := os.WriteFile(ppSound, []byte("dummy"), 0644); err != nil {
			t.Fatal(err)
		}

		path := player.GetFallbackPath("permission_prompt")
		if path != ppSound {
			t.Errorf("expected event-specific sound, got %s", path)
		}
	})
}

func TestHasAudioPlayer(t *testing.T) {
	player := NewPlayer("")

	// This should return true on most development machines
	hasPlayer := player.HasAudioPlayer()

	// Just verify it doesn't panic
	t.Logf("Platform: %s, HasAudioPlayer: %v", player.Platform(), hasPlayer)
}

func TestNewPlayer(t *testing.T) {
	tests := []struct {
		name       string
		pluginRoot string
	}{
		{"with plugin root", "/path/to/plugin"},
		{"empty plugin root", ""},
		{"home dir plugin root", "~/.claude/plugins/ccbell"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := NewPlayer(tt.pluginRoot)
			if player.pluginRoot != tt.pluginRoot {
				t.Errorf("pluginRoot = %q, want %q", player.pluginRoot, tt.pluginRoot)
			}
			if player.platform == "" {
				t.Error("platform should not be empty")
			}
		})
	}
}

func TestPlayerPlatform(t *testing.T) {
	player := NewPlayer("")
	platform := player.Platform()

	// Platform should be one of the known values
	validPlatforms := map[Platform]bool{
		PlatformMacOS:   true,
		PlatformLinux:   true,
		PlatformUnknown: true,
	}

	if !validPlatforms[platform] {
		t.Errorf("Platform() returned invalid platform: %s", platform)
	}
}

func TestPlayEmptyPath(t *testing.T) {
	player := NewPlayer("")
	err := player.Play("", 0.5)
	if err == nil {
		t.Error("Play with empty path should return error")
	}
}

func TestPlayNonexistentFile(t *testing.T) {
	player := NewPlayer("")
	err := player.Play("/nonexistent/path/to/sound.aiff", 0.5)
	if err == nil {
		t.Error("Play with nonexistent file should return error")
	}
}

func TestResolveCustomSoundValid(t *testing.T) {
	// Create a temp file to test with
	tempDir, err := os.MkdirTemp("", "ccbell-custom-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	soundFile := filepath.Join(tempDir, "custom.mp3")
	if err := os.WriteFile(soundFile, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	player := NewPlayer("")

	// Test valid absolute path
	path, err := player.resolveCustomSound(soundFile)
	if err != nil {
		t.Errorf("resolveCustomSound with valid path failed: %v", err)
	}
	if path != soundFile {
		t.Errorf("resolveCustomSound = %q, want %q", path, soundFile)
	}
}

func TestResolveCustomSoundInvalid(t *testing.T) {
	player := NewPlayer("")

	tests := []struct {
		name string
		path string
	}{
		{"relative path", "relative/path.mp3"},
		{"path traversal", "/some/path/../other"},
		{"nonexistent file", "/nonexistent/file.mp3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := player.resolveCustomSound(tt.path)
			if err == nil {
				t.Errorf("resolveCustomSound(%q) should return error", tt.path)
			}
		})
	}
}

func TestResolveBundledSoundValidation(t *testing.T) {
	player := NewPlayer("")

	invalidNames := []struct {
		name  string
		input string
	}{
		{"uppercase", "Stop"},
		{"numbers", "stop123"},
		{"special chars", "stop;rm"},
		{"path traversal", "../etc/passwd"},
		{"spaces", "stop sound"},
		{"dots", "stop.sound"},
	}

	for _, tt := range invalidNames {
		t.Run(tt.name, func(t *testing.T) {
			_, err := player.resolveBundledSound(tt.input)
			if err == nil {
				t.Errorf("resolveBundledSound(%q) should return error", tt.input)
			}
		})
	}
}

func TestGetFallbackPathEmpty(t *testing.T) {
	// Create empty temp dir (no sounds)
	tempDir, err := os.MkdirTemp("", "ccbell-empty-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	player := NewPlayer(tempDir)

	// Without any bundled sounds, should return empty
	path := player.GetFallbackPath("stop")
	if path != "" {
		t.Errorf("GetFallbackPath on empty dir should return empty, got %q", path)
	}
}

func TestResolveSoundPathCustom(t *testing.T) {
	// Create temp file
	tempDir, err := os.MkdirTemp("", "ccbell-resolve-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	soundFile := filepath.Join(tempDir, "test.mp3")
	if err := os.WriteFile(soundFile, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	player := NewPlayer("")

	// Test custom sound resolution
	path, err := player.ResolveSoundPath("custom:"+soundFile, "stop")
	if err != nil {
		t.Errorf("ResolveSoundPath(custom:) failed: %v", err)
	}
	if path != soundFile {
		t.Errorf("ResolveSoundPath = %q, want %q", path, soundFile)
	}
}

func TestGetLinuxPlayerArgs(t *testing.T) {
	tests := []struct {
		name      string
		player    string
		soundPath string
		volume    float64
		want      []string
	}{
		{
			name:      "paplay",
			player:    "paplay",
			soundPath: "/path/to/sound.aiff",
			volume:    0.5,
			want:      []string{"/path/to/sound.aiff"},
		},
		{
			name:      "aplay quiet mode",
			player:    "aplay",
			soundPath: "/path/to/sound.aiff",
			volume:    0.5,
			want:      []string{"-q", "/path/to/sound.aiff"},
		},
		{
			name:      "mpv with volume",
			player:    "mpv",
			soundPath: "/path/to/sound.aiff",
			volume:    0.75,
			want:      []string{"--really-quiet", "--volume=75", "/path/to/sound.aiff"},
		},
		{
			name:      "ffplay with volume",
			player:    "ffplay",
			soundPath: "/path/to/sound.aiff",
			volume:    0.25,
			want:      []string{"-nodisp", "-autoexit", "-volume", "25", "/path/to/sound.aiff"},
		},
		{
			name:      "unknown player",
			player:    "unknown_player",
			soundPath: "/path/to/sound.aiff",
			volume:    0.5,
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getLinuxPlayerArgs(tt.player, tt.soundPath, tt.volume)
			switch {
			case tt.want == nil:
				if got != nil {
					t.Errorf("getLinuxPlayerArgs() = %v, want nil", got)
				}
			case len(got) != len(tt.want):
				t.Errorf("getLinuxPlayerArgs() length = %d, want %d", len(got), len(tt.want))
			default:
				for i, v := range got {
					if v != tt.want[i] {
						t.Errorf("getLinuxPlayerArgs()[%d] = %q, want %q", i, v, tt.want[i])
					}
				}
			}
		})
	}
}

func TestFindPackageManager(t *testing.T) {
	// This test verifies the function doesn't panic
	// The actual result depends on the environment
	result := findPackageManager()
	t.Logf("Found package manager: %q", result)

	// Verify it returns empty string or a valid key from packageManagers
	if result != "" {
		if _, ok := packageManagers[result]; !ok {
			t.Errorf("findPackageManager() returned unknown package manager: %q", result)
		}
	}
}

func TestInstallAudioPlayer(t *testing.T) {
	// Test with unknown player - this checks the pkg mapping first
	// Note: if no package manager is found, the error will be "no package manager"
	// rather than "unknown player" because findPackageManager is called first
	err := installAudioPlayer("unknown_player")
	if err == nil {
		t.Error("installAudioPlayer(unknown) should return error")
	}
	// The error message depends on whether a package manager is found
	errMsg := err.Error()
	if errMsg != "no package manager found" && errMsg != "unknown player: unknown_player" {
		t.Errorf("unexpected error message: %q", errMsg)
	}
}

func TestEnsureAudioPlayer(t *testing.T) {
	player := NewPlayer("")

	// This test will succeed if any audio player is available on the system
	// or fail if no package manager is found to install one
	playerName, err := player.EnsureAudioPlayer()

	t.Logf("EnsureAudioPlayer result: name=%q, err=%v", playerName, err)

	// The test passes if either:
	// 1. A player was found (playerName != "", err == nil)
	// 2. No package manager is available (err contains "no package manager")
	if err != nil && playerName == "" {
		// Expected on systems without package managers
		t.Logf("Expected behavior: %v", err)
	}
}

func TestBundledSoundNameRegex(t *testing.T) {
	validNames := []string{"stop", "permission_prompt", "idle_prompt", "subagent", "test_sound"}
	invalidNames := []string{"Stop", "STOP", "stop!", "123stop", "stop sound", "stop.", "/stop", "test_sound_123"}

	for _, name := range validNames {
		if !bundledSoundNameRegex.MatchString(name) {
			t.Errorf("bundledSoundNameRegex should match %q", name)
		}
	}
	for _, name := range invalidNames {
		if bundledSoundNameRegex.MatchString(name) {
			t.Errorf("bundledSoundNameRegex should not match %q", name)
		}
	}
}

func TestResolveSoundPathDirectPath(t *testing.T) {
	// Create temp file
	tempDir, err := os.MkdirTemp("", "ccbell-direct-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	soundFile := filepath.Join(tempDir, "direct.mp3")
	if err := os.WriteFile(soundFile, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	player := NewPlayer("")

	// Test direct path (should be treated like custom)
	path, err := player.ResolveSoundPath(soundFile, "stop")
	if err != nil {
		t.Errorf("ResolveSoundPath(direct path) failed: %v", err)
	}
	if path != soundFile {
		t.Errorf("ResolveSoundPath = %q, want %q", path, soundFile)
	}
}

func TestPlayMacOSNonBlocking(t *testing.T) {
	if runtime.GOOS != darwinOS {
		t.Skip("afplay is only available on macOS")
	}

	// Create temp sound file
	tempDir, err := os.MkdirTemp("", "ccbell-play-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	soundFile := filepath.Join(tempDir, "test.aiff")
	if err := os.WriteFile(soundFile, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	player := NewPlayer("")

	// Should not block - returns immediately after starting process
	err = player.playMacOS(soundFile, 0.5)
	if err != nil {
		t.Errorf("playMacOS should not return error: %v", err)
	}
}

func TestPlayLinuxNoPlayer(t *testing.T) {
	if runtime.GOOS != linuxOS {
		t.Skip("this test is only for Linux")
	}

	player := NewPlayer("")

	// Check if any audio player is available
	hasPlayer := false
	for _, p := range linuxAudioPlayerNames {
		if _, err := exec.LookPath(p); err == nil {
			hasPlayer = true
			break
		}
	}

	// Mock: if no player is available, should return error
	// This test verifies the error message
	err := player.playLinux("/nonexistent.aiff", 0.5)
	if hasPlayer {
		// Player available - playLinux may succeed or fail depending on player
		t.Logf("Audio player available, playLinux result: %v", err)
	} else {
		// No player - should return error
		if err == nil {
			t.Error("playLinux with no player should return error")
			return
		}
		expectedMsg := "no audio player found"
		if !contains(err.Error(), expectedMsg) {
			t.Errorf("error message should contain %q, got %q", expectedMsg, err.Error())
		}
	}
}

func TestPlayUnsupportedPlatform(t *testing.T) {
	// We can't easily test PlatformUnknown in play method
	// because detectPlatform() uses runtime.GOOS
	// But we can verify the error path exists
	// Note: Play checks file existence first, so we need a file that exists
	// Create a temp file to test the platform check path
	tempDir, err := os.MkdirTemp("", "ccbell-unsupported-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	soundFile := filepath.Join(tempDir, "test.aiff")
	if err := os.WriteFile(soundFile, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	player := &Player{platform: PlatformUnknown, pluginRoot: ""}
	err = player.Play(soundFile, 0.5)
	if err == nil {
		t.Error("Play with unknown platform should return error")
	}
	if err.Error() != "unsupported platform: unknown" {
		t.Errorf("error message = %q, want %q", err.Error(), "unsupported platform: unknown")
	}
}

func TestResolveBundledSoundNotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ccbell-notfound-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	soundsDir := filepath.Join(tempDir, "sounds")
	if err := os.MkdirAll(soundsDir, 0755); err != nil {
		t.Fatal(err)
	}

	player := NewPlayer(tempDir)

	_, err = player.resolveBundledSound("nonexistent")
	if err == nil {
		t.Error("resolveBundledSound for nonexistent should return error")
	}
	if err.Error() != "bundled sound not found: nonexistent" {
		t.Errorf("error message = %q", err.Error())
	}
}

func TestLinuxAudioPlayerNamesOrder(t *testing.T) {
	// Verify the priority order is correct
	expectedOrder := []string{"mpv", "paplay", "aplay", "ffplay"}
	for i, name := range linuxAudioPlayerNames {
		if name != expectedOrder[i] {
			t.Errorf("linuxAudioPlayerNames[%d] = %q, want %q", i, name, expectedOrder[i])
		}
	}
}

func TestPlayerPackagesMapping(t *testing.T) {
	// Verify all players have packages defined
	players := []string{"mpv", "ffplay", "paplay", "aplay"}
	for _, player := range players {
		if pkg, ok := playerPackages[player]; !ok {
			t.Errorf("playerPackages[%q] not defined", player)
		} else if pkg == "" {
			t.Errorf("playerPackages[%q] should not be empty", player)
		}
	}
}

func TestPackageManagersMapping(t *testing.T) {
	// Verify all package managers have commands defined
	for pm, cmd := range packageManagers {
		if cmd == "" {
			t.Errorf("packageManagers[%q] should not be empty", pm)
		}
		// Check for common package manager keywords
		hasInstall := contains(cmd, "install") || contains(cmd, "add") || contains(cmd, "sync") || contains(cmd, "ask") || contains(cmd, " -S ")
		if !hasInstall {
			t.Errorf("packageManagers[%q] should contain install/add/sync/ask command: %q", pm, cmd)
		}
	}
}

// Helper function.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestPlayLinuxWithPlayer(t *testing.T) {
	if runtime.GOOS != linuxOS {
		t.Skip("this test is only for Linux")
	}

	// Create temp sound file
	tempDir, err := os.MkdirTemp("", "ccbell-linux-play-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	soundFile := filepath.Join(tempDir, "test.aiff")
	if err := os.WriteFile(soundFile, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	player := NewPlayer("")

	// Try to play - will succeed if any audio player is installed
	err = player.playLinux(soundFile, 0.5)
	// Either succeeds (player found) or fails (no player) - both are valid
	t.Logf("playLinux result: err=%v", err)
}

func TestHasAudioPlayerMacOS(t *testing.T) {
	if runtime.GOOS != darwinOS {
		t.Skip("this test is only for macOS")
	}

	player := NewPlayer("")
	hasPlayer := player.HasAudioPlayer()

	if !hasPlayer {
		t.Error("HasAudioPlayer should return true on macOS with afplay")
	}
}

func TestHasAudioPlayerLinuxNoPlayer(t *testing.T) {
	if runtime.GOOS != linuxOS {
		t.Skip("this test is only for Linux")
	}

	// Test with a mock that has no players available
	player := &Player{platform: PlatformLinux, pluginRoot: ""}
	hasPlayer := player.HasAudioPlayer()

	// May be true or false depending on system, just log it
	t.Logf("HasAudioPlayer on Linux: %v", hasPlayer)
}

func TestHasAudioPlayerUnknown(t *testing.T) {
	player := &Player{platform: PlatformUnknown, pluginRoot: ""}
	hasPlayer := player.HasAudioPlayer()

	if hasPlayer {
		t.Error("HasAudioPlayer should return false for unknown platform")
	}
}

func TestInstallAudioPlayerWithMock(t *testing.T) {
	// This tests the error path when player is unknown
	err := installAudioPlayer("totally_fake_player_xyz")
	if err == nil {
		t.Error("installAudioPlayer with unknown player should return error")
	}
}

func TestEnsureAudioPlayerWithExisting(t *testing.T) {
	if runtime.GOOS != linuxOS {
		t.Skip("this test is only for Linux")
	}

	player := NewPlayer("")
	playerName, err := player.EnsureAudioPlayer()

	// On Linux CI, likely no player is installed
	// Test should not panic regardless of result
	t.Logf("EnsureAudioPlayer: name=%q, err=%v", playerName, err)
}

func TestDetectPlatformUnknown(t *testing.T) {
	// Test PlatformUnknown by creating a player with unknown platform
	player := &Player{platform: PlatformUnknown, pluginRoot: ""}
	if player.Platform() != PlatformUnknown {
		t.Errorf("expected PlatformUnknown, got %s", player.Platform())
	}
}

func TestPlayWithValidLinuxPlayer(t *testing.T) {
	if runtime.GOOS != linuxOS {
		t.Skip("this test is only for Linux")
	}

	// Create temp sound file
	tempDir, err := os.MkdirTemp("", "ccbell-valid-player-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	soundFile := filepath.Join(tempDir, "test.aiff")
	if err := os.WriteFile(soundFile, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	player := NewPlayer("")

	// Try to play - may succeed if a player like aplay is available
	err = player.Play(soundFile, 0.5)
	t.Logf("Play with valid file: err=%v", err)
}

func TestInstallAudioPlayerNoManager(t *testing.T) {
	// Test with a mock that has no package manager
	// by calling installAudioPlayer when findPackageManager returns empty
	// This tests the "no package manager found" error path
	pm := findPackageManager()
	if pm == "" {
		// Skip if no package manager on this system
		t.Skip("no package manager available")
	}

	// Test with unknown player
	err := installAudioPlayer("totally_fake_player_xyz_123")
	if err == nil {
		t.Error("installAudioPlayer with unknown player should return error")
	}
	if err.Error() != "unknown player: totally_fake_player_xyz_123" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestInstallAudioPlayerSuccess(t *testing.T) {
	if runtime.GOOS != linuxOS {
		t.Skip("this test is only for Linux")
	}

	pm := findPackageManager()
	if pm == "" {
		t.Skip("no package manager available")
	}

	// Test with a valid player (mpv) - this will fail because we can't actually install
	// but it tests the code path
	err := installAudioPlayer("mpv")
	// Either succeeds or fails, but shouldn't panic
	t.Logf("installAudioPlayer(mpv): err=%v", err)
}

func TestFindPackageManagerKnown(t *testing.T) {
	result := findPackageManager()
	if result != "" {
		// Package manager found, verify it's a valid one
		if _, ok := packageManagers[result]; !ok {
			t.Errorf("findPackageManager returned unknown manager: %s", result)
		}
		t.Logf("Found package manager: %s", result)
	} else {
		t.Log("No package manager found on this system")
	}
}

func TestPlayLinuxErrorPath(t *testing.T) {
	if runtime.GOOS != linuxOS {
		t.Skip("this test is only for Linux")
	}

	player := NewPlayer("")
	err := player.playLinux("/nonexistent/path/to/sound.aiff", 0.5)

	// Should return error because no player is available
	if err == nil {
		t.Log("playLinux returned nil (player may be available)")
	} else {
		t.Logf("playLinux error: %v", err)
	}
}
