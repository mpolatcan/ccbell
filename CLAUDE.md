# ccbell

Go-based CLI binary that provides audio notifications for Claude Code events.

**GitHub:** mpolatcan/ccbell

## Purpose

Plays sound notifications when Claude finishes responding, needs permission, is waiting for input, or when a subagent completes.

## Architecture

### Plugin Distribution (cc-plugins repo)

The `cc-plugins` repository distributes the plugin with:
- `plugins/ccbell/.claude-plugin/plugin.json` - Plugin manifest (no scripts/postinstall support in Claude Code)
- `plugins/ccbell/hooks/hooks.json` - Hook definitions for Claude Code events
- `plugins/ccbell/sounds/*.aiff` - Audio files for each event type
- `plugins/ccbell/commands/*.md` - Slash command documentation

### Binary Distribution (ccbell repo)

The `ccbell` repository builds and releases Go binaries via GitHub Actions.

**Note:** Claude Code plugins don't support `scripts/postinstall` in the manifest. Binary installation must be done manually or via a different mechanism.

## Commands

```bash
make build           # Build for current platform
make test            # Run tests with race detection
make coverage        # Generate coverage report
make lint            # Run linter (golangci-lint or go vet)
make dist            # Cross-compile for all platforms
make release         # Build, checksum, and create archives
make check           # Run fmt, lint, test, and build
make install         # Install to ~/.claude/plugins/local/ccbell/bin/
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
3. GitHub Actions builds binaries and creates releases automatically at https://github.com/mpolatcan/ccbell/releases
