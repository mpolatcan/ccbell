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
│   ├── audio/     # Playback   │   │   ├── plugin.json
│   ├── config/    # Config     │   │   └── marketplace.json
│   ├── logger/    # Logging    │   ├── hooks/hooks.json
│   └── state/     # Cooldown   │   ├── sounds/*.aiff
├── go.mod                      │   ├── commands/*.md
└── Makefile                    └── scripts/install.sh
```

**Flow:**
1. User installs plugin from cc-plugins marketplace
2. `postinstall` script downloads binary from this repo's GitHub Releases
3. Claude Code hooks call the binary for sound notifications

## Development

### Prerequisites

- Go 1.25+
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
- `ccbell-windows-amd64.exe`

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
3. Creates archives (tar.gz/zip)
4. Generates checksums
5. Creates GitHub Release with all artifacts

## Configuration

The binary reads configuration from:
- **Project:** `.claude/ccbell.config.json`
- **Global:** `~/.claude/ccbell.config.json`

See the [plugin README](https://github.com/mpolatcan/cc-plugins/tree/main/plugins/ccbell) for configuration options.

## Platform Support

| Platform | Audio Backend |
|----------|--------------|
| macOS | `afplay` (built-in) |
| Linux | `paplay`, `aplay`, `mpv`, or `ffplay` |
| Windows | PowerShell `Media.SoundPlayer` |

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
