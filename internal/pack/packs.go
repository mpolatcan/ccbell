// Package pack provides sound pack management for ccbell.
// Sound packs bundle sounds for all notification events and are distributed via GitHub releases.
package pack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	// PackIndexURL is the URL to fetch the pack index from.
	PackIndexURL = "https://api.github.com/repos/mpolatcan/ccbell-sound-packs/releases"
	// PackOwner is the GitHub owner for sound pack releases.
	PackOwner = "mpolatcan"
	// PackRepo is the repository name for sound packs.
	PackRepo = "ccbell-sound-packs"
	// PacksDir is the directory name for installed packs.
	PacksDir = "packs"
	// FileMode is the permission mode for pack files.
	FileMode = 0600
)

// Pack represents a sound pack.
type Pack struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Version     string            `json:"version"`
	Events      map[string]string `json:"events"` // event_type -> sound_filename
	PreviewURL  string            `json:"previewUrl,omitempty"`
	DownloadURL string            `json:"downloadUrl"`
	PublishedAt string            `json:"publishedAt"`
	TagName     string            `json:"tagName,omitempty"`
}

// PackIndex represents the index of available sound packs.
type PackIndex struct {
	Packs     []Pack `json:"packs"`
	UpdatedAt string `json:"updatedAt"`
}

// PackManifest represents the manifest inside a pack archive.
type PackManifest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Version     string            `json:"version"`
	Events      map[string]string `json:"events"` // event_type -> sound_filename
}

// InstalledPack represents an installed pack in the local filesystem.
type InstalledPack struct {
	Manifest   PackManifest
	InstallDir string
}

// Manager handles sound pack operations.
type Manager struct {
	homeDir    string
	packsDir   string
	configPath string
	httpClient *http.Client
}

// NewManager creates a new pack manager.
func NewManager(homeDir string) *Manager {
	packsDir := ""
	configPath := ""
	if homeDir != "" {
		packsDir = filepath.Join(homeDir, ".claude", "ccbell", PacksDir)
		configPath = filepath.Join(homeDir, ".claude", "ccbell.config.json")
	}

	return &Manager{
		homeDir:    homeDir,
		packsDir:   packsDir,
		configPath: configPath,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListAvailable fetches and returns available packs from GitHub releases.
func (m *Manager) ListAvailable() ([]Pack, error) {
	if m.httpClient == nil {
		m.httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	req, err := http.NewRequest("GET", PackIndexURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ccbell")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pack index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch pack index: %s", string(body))
	}

	var releases []struct {
		TagName string `json:"tag_name"`
		Name    string `json:"name"`
		Body    string `json:"body"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
		PublishedAt string `json:"published_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode pack index: %w", err)
	}

	var packs []Pack
	for _, release := range releases {
		pack := Pack{
			ID:          release.TagName,
			Name:        release.Name,
			Description: release.Body,
			Version:     strings.TrimPrefix(release.TagName, "v"),
			PublishedAt: release.PublishedAt,
			Events:      make(map[string]string),
			TagName:     release.TagName,
		}

		// Find the pack.json asset
		for _, asset := range release.Assets {
			if asset.Name == "pack.json" {
				pack.DownloadURL = asset.BrowserDownloadURL
				break
			}
		}

		// Find preview audio
		for _, asset := range release.Assets {
			if strings.HasPrefix(asset.Name, "preview.") {
				pack.PreviewURL = asset.BrowserDownloadURL
				break
			}
		}

		if pack.DownloadURL != "" {
			packs = append(packs, pack)
		}
	}

	return packs, nil
}

// Install downloads and installs a pack.
func (m *Manager) Install(packID string) error {
	if m.packsDir == "" {
		return fmt.Errorf("home directory not set")
	}

	// Fetch pack info (includes asset URLs from GitHub releases)
	packs, err := m.ListAvailable()
	if err != nil {
		return err
	}

	var targetPack Pack
	for _, p := range packs {
		if p.ID == packID || p.ID == "v"+packID {
			targetPack = p
			break
		}
	}

	if targetPack.DownloadURL == "" {
		return fmt.Errorf("pack not found: %s", packID)
	}

	// Create pack directory
	packDir := filepath.Join(m.packsDir, packID)
	if err := os.MkdirAll(packDir, 0755); err != nil {
		return fmt.Errorf("failed to create pack directory: %w", err)
	}

	// Download pack.json
	packData, err := m.downloadFile(targetPack.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download pack.json: %w", err)
	}

	// Save pack.json
	manifestPath := filepath.Join(packDir, "pack.json")
	if err := os.WriteFile(manifestPath, packData, FileMode); err != nil {
		return fmt.Errorf("failed to save pack manifest: %w", err)
	}

	// Parse pack.json to get the events map (sound filenames)
	var manifest PackManifest
	if err := json.Unmarshal(packData, &manifest); err != nil {
		return fmt.Errorf("failed to parse pack.json: %w", err)
	}

	// Fetch the release assets by tag to get download URLs for sound files
	releaseAssets, err := m.fetchReleaseAssets(targetPack.TagName)
	if err != nil {
		return fmt.Errorf("failed to fetch release assets: %w", err)
	}

	// Build a map of asset name -> download URL
	assetURLs := make(map[string]string)
	for _, asset := range releaseAssets {
		assetURLs[asset.Name] = asset.DownloadURL
	}

	// Download each sound file referenced in the manifest
	for _, soundFile := range manifest.Events {
		assetURL, ok := assetURLs[soundFile]
		if !ok {
			fmt.Printf("Warning: sound file '%s' not found in release assets, skipping\n", soundFile)
			continue
		}

		soundData, err := m.downloadFile(assetURL)
		if err != nil {
			return fmt.Errorf("failed to download sound file %s: %w", soundFile, err)
		}

		soundPath := filepath.Join(packDir, soundFile)
		if err := os.WriteFile(soundPath, soundData, FileMode); err != nil {
			return fmt.Errorf("failed to save sound file %s: %w", soundFile, err)
		}
	}

	packDirAbs, _ := filepath.Abs(packDir)
	fmt.Printf("Pack '%s' installed to %s\n", targetPack.Name, packDirAbs)

	return nil
}

// releaseAsset represents a single asset in a GitHub release.
type releaseAsset struct {
	Name        string
	DownloadURL string
}

// fetchReleaseAssets fetches the asset list for a specific release by tag.
func (m *Manager) fetchReleaseAssets(tagName string) ([]releaseAsset, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", PackOwner, PackRepo, tagName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ccbell")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch release %s: %s", tagName, string(body))
	}

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	assets := make([]releaseAsset, len(release.Assets))
	for i, a := range release.Assets {
		assets[i] = releaseAsset{Name: a.Name, DownloadURL: a.BrowserDownloadURL}
	}
	return assets, nil
}

// downloadFile downloads a file from a URL and returns its contents.
func (m *Manager) downloadFile(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Uninstall removes an installed pack.
func (m *Manager) Uninstall(packID string) error {
	if m.packsDir == "" {
		return fmt.Errorf("home directory not set")
	}

	packDir := filepath.Join(m.packsDir, packID)
	if _, err := os.Stat(packDir); os.IsNotExist(err) {
		return fmt.Errorf("pack not installed: %s", packID)
	}

	if err := os.RemoveAll(packDir); err != nil {
		return fmt.Errorf("failed to remove pack: %w", err)
	}

	return nil
}

// ListInstalled returns a list of installed packs.
func (m *Manager) ListInstalled() ([]InstalledPack, error) {
	if m.packsDir == "" {
		return nil, fmt.Errorf("home directory not set")
	}

	entries, err := os.ReadDir(m.packsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []InstalledPack{}, nil
		}
		return nil, fmt.Errorf("failed to read packs directory: %w", err)
	}

	var installed []InstalledPack
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(m.packsDir, entry.Name(), "pack.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue // Skip invalid packs
		}

		var manifest PackManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}

		installed = append(installed, InstalledPack{
			Manifest:   manifest,
			InstallDir: filepath.Join(m.packsDir, entry.Name()),
		})
	}

	return installed, nil
}

// GetPackPath returns the path to an installed pack's sound file.
func (m *Manager) GetPackPath(packID, soundFile string) (string, error) {
	if m.packsDir == "" {
		return "", fmt.Errorf("home directory not set")
	}

	packDir := filepath.Join(m.packsDir, packID)
	path := filepath.Join(packDir, soundFile)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("sound file not found in pack: %s/%s", packID, soundFile)
	}

	return path, nil
}

// UsePack applies a pack's sounds to the configuration and updates config file.
func (m *Manager) UsePack(packID string) error {
	installed, err := m.ListInstalled()
	if err != nil {
		return err
	}

	var target InstalledPack
	for _, p := range installed {
		if p.Manifest.ID == packID || p.Manifest.ID == "v"+packID {
			target = p
			break
		}
	}

	if target.Manifest.ID == "" {
		return fmt.Errorf("pack not installed: %s (use /ccbell:packs install %s first)", packID, packID)
	}

	// Update config file with pack sounds
	if m.configPath != "" {
		if err := m.updateConfigWithPack(target); err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}
	}

	return nil
}

// updateConfigWithPack updates the config file to use sounds from the pack.
func (m *Manager) updateConfigWithPack(pack InstalledPack) error {
	// Read existing config
	var config map[string]interface{}
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		// Create new config if doesn't exist
		config = make(map[string]interface{})
	} else {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// Set active pack
	config["activePack"] = pack.Manifest.ID

	// Ensure events section exists
	if _, ok := config["events"]; !ok {
		config["events"] = make(map[string]interface{})
	}

	events := config["events"].(map[string]interface{})

	// Update each event with pack sound
	for eventType, soundFile := range pack.Manifest.Events {
		events[eventType] = map[string]interface{}{
			"sound": fmt.Sprintf("pack:%s:%s", pack.Manifest.ID, soundFile),
		}
	}

	// Write updated config
	output, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetPackSound returns the sound file path for a specific event in a pack.
func (m *Manager) GetPackSound(packID, eventType string) (string, error) {
	installed, err := m.ListInstalled()
	if err != nil {
		return "", err
	}

	for _, p := range installed {
		if p.Manifest.ID == packID || p.Manifest.ID == "v"+packID {
			soundFile, ok := p.Manifest.Events[eventType]
			if !ok {
				return "", fmt.Errorf("event %s not found in pack %s", eventType, packID)
			}
			return filepath.Join(p.InstallDir, soundFile), nil
		}
	}

	return "", fmt.Errorf("pack not installed: %s", packID)
}

// PacksDir returns the packs installation directory.
func (m *Manager) PacksDir() string {
	return m.packsDir
}

// Preview plays a preview sound from an available pack.
func (m *Manager) Preview(packID string) error {
	packs, err := m.ListAvailable()
	if err != nil {
		return err
	}

	var target Pack
	for _, p := range packs {
		if p.ID == packID || p.ID == "v"+packID {
			target = p
			break
		}
	}

	if target.PreviewURL == "" {
		return fmt.Errorf("pack %s has no preview sound", packID)
	}

	// Download preview to temp file
	resp, err := m.httpClient.Get(target.PreviewURL)
	if err != nil {
		return fmt.Errorf("failed to download preview: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download preview: HTTP %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "ccbell-preview-*."+getAudioExtension(target.PreviewURL))
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to save preview: %w", err)
	}
	tmpFile.Close()

	// Play the preview
	return playAudio(tmpFile.Name())
}

// playAudio plays an audio file using the appropriate player for the platform.
func playAudio(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("afplay", path).Start()
	case "linux":
		// Try different players
		players := []string{"mpv", "ffplay", "paplay", "aplay"}
		for _, player := range players {
			if _, err := exec.LookPath(player); err == nil {
				var cmd *exec.Cmd
				switch player {
				case "mpv":
					cmd = exec.Command(player, "--really-quiet", path)
				case "ffplay":
					cmd = exec.Command(player, "-nodisp", "-autoexit", path)
				default:
					cmd = exec.Command(player, path)
				}
				return cmd.Start()
			}
		}
		return fmt.Errorf("no audio player found (install mpv or ffmpeg)")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// getAudioExtension returns the file extension from a URL.
func getAudioExtension(url string) string {
	ext := filepath.Ext(url)
	if ext != "" {
		return strings.TrimPrefix(ext, ".")
	}
	return "aiff"
}

// packNameRegex validates pack names.
var packNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidatePackID validates a pack identifier.
func ValidatePackID(packID string) error {
	packID = strings.TrimPrefix(packID, "v")
	if !packNameRegex.MatchString(packID) {
		return fmt.Errorf("invalid pack ID: %s (must be alphanumeric with hyphens/underscores)", packID)
	}
	return nil
}
