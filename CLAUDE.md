# ccbell

Go-based CLI binary that provides audio notifications for Claude Code events.

**GitHub:** mpolatcan/ccbell

## Purpose

Plays sound notifications when Claude finishes responding, needs permission, is waiting for input, or when a subagent completes.

## Architecture

The `ccbell` repository builds and releases Go binaries via GitHub Actions.

The `cc-plugins` repository distributes the plugin with:
- `.claude-plugin/plugin.json` - Plugin manifest with commands and hooks
- `sounds/*.aiff` - Audio files
- `commands/*.md` - Slash command documentation
- `scripts/ccbell.sh` - Downloads binary from GitHub releases on first use

## Official Documentation (ALWAYS refer to these)

Plugin schemas and hook events may change. Always consult the official documentation.

| Documentation | URL |
|--------------|-----|
| **Plugins Reference** | https://code.claude.com/docs/en/plugins-reference |
| **Hooks Reference** | https://code.claude.com/docs/en/hooks |

### Key Sections

- **Plugin Manifest** - https://code.claude.com/docs/en/plugins-reference#manifest
- **Commands** - https://code.claude.com/docs/en/plugins-reference#commands
- **Hooks** - https://code.claude.com/docs/en/plugins-reference#hooks
- **Hook Events** - https://code.claude.com/docs/en/hooks#events
- **Hook Types** - https://code.claude.com/docs/en/hooks#hook-types

## Commands

```bash
make build           # Build for current platform
make test            # Run tests with race detection
make coverage        # Generate coverage report
make lint            # Run linter (golangci-lint or go vet)
make dist            # Cross-compile for all platforms
make release         # Build, checksum, and create archives
make check           # Run fmt, lint, test, and build
```

## Usage

```bash
ccbell <event_type>
```

Event types: `stop`, `permission_prompt`, `idle_prompt`, `subagent`

## Audio Backends

| Platform | Backend |
|----------|---------|
| macOS | `afplay` |
| Linux | `paplay`, `aplay`, `mpv`, or `ffplay` |
| Windows | PowerShell `Media.SoundPlayer` |

## Release Process

1. Create git tag: `git tag vX.Y.Z`
2. Push tag: `git push origin vX.Y.Z`
3. GitHub Actions builds and creates releases at https://github.com/mpolatcan/ccbell/releases
