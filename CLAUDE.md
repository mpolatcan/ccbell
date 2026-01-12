# ccbell

Go-based CLI binary that provides audio notifications for Claude Code events.

**GitHub:** mpolatcan/ccbell

## Purpose

Plays sound notifications when Claude finishes responding, needs permission, is waiting for input, or when a subagent completes.

## Architecture

### Plugin Distribution (cc-plugins repo)

The `cc-plugins` repository distributes the plugin with:
- `plugins/ccbell/plugin.json` - Plugin manifest
- `plugins/ccbell/hooks/hooks.json` - Hook definitions for Claude Code events
- `plugins/ccbell/sounds/*.aiff` - Audio files for each event type
- `plugins/ccbell/commands/*.md` - Slash command documentation

**Important:** Always refer to the official Claude Code documentation for the latest plugin and hooks specifications:
- [Plugins Reference](https://code.claude.com/docs/en/plugins-reference)
- [Hooks Reference](https://code.claude.com/docs/en/hooks)

Plugin schemas and hook events may change. The documentation above reflects the current state at the time of writing.

**Hook Events Used:**
| Event | Description |
|-------|-------------|
| `Stop` | Claude finishes responding |
| `PermissionPrompt` | Claude needs permission |
| `UserPromptSubmit` | User waiting for input |
| `SubagentStop` | Subagent task completes |

**Hook Structure (hooks.json):**
```json
{
  "Stop": [
    {
      "matcher": "*",
      "hooks": [
        {
          "type": "command",
          "command": "ccbell stop"
        }
      ]
    }
  ]
}
```

### Binary Distribution (ccbell repo)

The `ccbell` repository builds and releases Go binaries via GitHub Actions.

**Note:** Binary installation is automatically handled by the `ccbell.sh` script in the cc-plugins repository, which downloads the correct Go binary for your platform from GitHub releases.

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

Event types: `stop`, `permission_prompt`, `user_prompt_submit`, `subagent_stop`

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
