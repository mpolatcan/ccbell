# ccbell

Go source code for the ccbell Claude Code plugin - audio notifications for Claude Code events.

> **Looking to install the plugin?** See [cc-plugins/plugins/ccbell](https://github.com/mpolatcan/cc-plugins)

## Overview

This repository contains the Go source code for the `ccbell` binary. The distributable plugin (with sounds, hooks, and commands) is maintained in the [cc-plugins](https://github.com/mpolatcan/cc-plugins) marketplace repository.

## Architecture

```
ccbell (this repo)              cc-plugins (marketplace)
├── cmd/ccbell/main.go          ├── plugins/ccbell/
├── internal/                   │   ├── .claude-plugin/
│   ├── audio/     # Playback   │   │   ├── plugin.json (metadata only)
│   ├── config/    # Config     │   ├── sounds/*.aiff
│   ├── logger/    # Logging    │   ├── commands/*.md
│   └── state/     # Cooldown   └── scripts/ccbell.sh
├── go.mod
└── Makefile
```

**Flow:**
1. User installs plugin from cc-plugins marketplace
2. `ccbell.sh` script downloads binary on first use from this repo's GitHub Releases
3. Claude Code hooks call `ccbell.sh` for sound notifications

## Development

### Prerequisites

- Go 1.25.5+
- Make (optional)

### Build

```bash
# Build for current platform
make build

# Or without make
go build -o bin/ccbell ./cmd/ccbell
```

### Test

```bash
make test

# With coverage
make coverage
```

### Lint

```bash
make lint
```

### Run Locally

```bash
# Test the binary
./bin/ccbell --version
./bin/ccbell --help
./bin/ccbell stop  # Play stop sound
```

## Project Structure

```
ccbell/
├── cmd/
│   └── ccbell/
│       └── main.go          # Entry point
├── internal/
│   ├── audio/
│   │   ├── player.go        # Cross-platform audio playback
│   │   └── player_test.go
│   ├── config/
│   │   ├── config.go        # Configuration loading
│   │   ├── config_test.go
│   │   ├── quiethours.go    # Quiet hours logic
│   │   └── quiethours_test.go
│   ├── logger/
│   │   └── logger.go        # Debug logging
│   └── state/
│       └── state.go         # Cooldown state management
├── .github/
│   └── workflows/
│       ├── ci.yml           # Test, lint, build
│       └── release.yml      # Create releases on tag
├── Makefile
├── go.mod
└── README.md
```

## Cross-Compilation

Build for all supported platforms:

```bash
make dist
```

Creates binaries in `dist/`:
- `ccbell-darwin-amd64` (macOS Intel)
- `ccbell-darwin-arm64` (macOS Apple Silicon)
- `ccbell-linux-amd64`
- `ccbell-linux-arm64`

## Creating a Release

Releases are automated via GitHub Actions when you push a tag:

```bash
# Tag a new version
git tag v1.0.0
git push origin v1.0.0
```

This triggers `.github/workflows/release.yml` which:
1. Runs tests
2. Builds binaries for all platforms
3. Creates archives (tar.gz)
4. Generates checksums
5. Creates GitHub Release with all artifacts

### Sync Version to cc-plugins Marketplace

After tagging and pushing in the ccbell repository, you MUST also update the version in cc-plugins:

```bash
cd ../cc-plugins/plugins/ccbell

# Step 1: Get current version from plugin.json
CURRENT_VERSION=$(grep '"version"' .claude-plugin/plugin.json | sed 's/.*: *"\([^"]*\)".*/\1/')
echo "Current version: $CURRENT_VERSION"

# Step 2: Calculate new version (adjust based on your release type)
# For patch: NEW_VERSION=$(echo $CURRENT_VERSION | awk -F. '{print $1"."$2"."($3+1)}')
# For minor: NEW_VERSION=$(echo $CURRENT_VERSION | awk -F. '{print $1"."($2+1)".0"}')
# For major: NEW_VERSION=$(echo $CURRENT_VERSION | awk -F. '{print ($1+1)".0.0"}')

# Example: Set new version to v0.3.0 (adjust as needed)
NEW_VERSION="0.3.0"

# Step 3: Update version files
sed -i '' "s/\"version\": *\"${CURRENT_VERSION}\"/\"version\": \"${NEW_VERSION}\"/" .claude-plugin/plugin.json
sed -i '' "s/PLUGIN_VERSION=\"${CURRENT_VERSION}\"/PLUGIN_VERSION=\"${NEW_VERSION}\"/" scripts/ccbell.sh

# Step 4: Commit and push
git add .claude-plugin/plugin.json scripts/ccbell.sh
git commit -m "chore(ccbell): sync version to v${NEW_VERSION}"
git push
```

**Release Checklist:**
1. [ ] Tag and push in ccbell: `git tag v<version> && git push origin v<version>`
2. [ ] Wait for GitHub Release to be created
3. [ ] Get current version from cc-plugins plugins/ccbell/.claude-plugin/plugin.json
4. [ ] Calculate/determine new version
5. [ ] Update `plugin.json` and `ccbell.sh` with new version
6. [ ] Commit and push version sync in cc-plugins

## Configuration

The binary reads configuration from:
- **Global:** `~/.claude/ccbell.config.json`

See the [plugin README](https://github.com/mpolatcan/cc-plugins/tree/main/plugins/ccbell) for configuration options.

## Platform Support

| Platform | Audio Backend |
|----------|--------------|
| macOS | `afplay` (built-in) |
| Linux | `paplay`, `aplay`, `mpv`, or `ffplay` |

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes
4. Run `make check` (fmt, lint, test, build)
5. Submit a pull request

## Related Repositories

- [cc-plugins](https://github.com/mpolatcan/cc-plugins) - Plugin marketplace containing the distributable ccbell plugin

## License

MIT
