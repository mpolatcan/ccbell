package audio

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

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
	case "darwin":
		if platform != PlatformMacOS {
			t.Errorf("expected PlatformMacOS on darwin, got %s", platform)
		}
	case "windows":
		if platform != PlatformWindows {
			t.Errorf("expected PlatformWindows on windows, got %s", platform)
		}
	case "linux":
		// Could be Linux or WSL
		if platform != PlatformLinux && platform != PlatformWSL {
			t.Errorf("expected PlatformLinux or PlatformWSL on linux, got %s", platform)
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
			if player == nil {
				t.Error("NewPlayer returned nil")
			}
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
		PlatformWindows: true,
		PlatformWSL:     true,
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
