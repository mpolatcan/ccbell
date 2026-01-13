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

**:warning: CRITICAL: Version must be synced between both repositories.**

After tagging and pushing in the ccbell repository, you MUST also update the version in cc-plugins to the SAME version:

**Use the make target (recommended):**
```bash
make sync-version VERSION=v1.0.0
```

**Or manually from cc-plugins:**
```bash
cd ../cc-plugins/plugins/ccbell

# Get version from ccbell (assuming ccbell is at ../../../ccbell)
NEW_VERSION=$(cd ../../../ccbell && git describe --tags --always --dirty | sed 's/^v//')

# Update both files
sed -i '' "s/\"version\": \"[0-9.]*\"/\"version\": \"${NEW_VERSION}\"/" .claude-plugin/plugin.json
sed -i '' "s/PLUGIN_VERSION=\"[0-9.]*\"/PLUGIN_VERSION=\"${NEW_VERSION}\"/" scripts/ccbell.sh

# Commit and push
git add .claude-plugin/plugin.json scripts/ccbell.sh
git commit -m "chore(ccbell): sync version to v${NEW_VERSION}"
git push
```

**:no_entry_sign: NEVER skip version sync - both repositories must use the same version number.**

**Release Checklist:**
1. [ ] Make changes in ccbell
2. [ ] Tag and push in ccbell: `git tag v<version> && git push origin v<version>`
3. [ ] Wait for GitHub Release to be created
4. [ ] Sync version to cc-plugins: `make sync-version VERSION=v<version>`
5. [ ] Commit and push version sync in cc-plugins

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
