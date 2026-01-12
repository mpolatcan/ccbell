# ccbell

Go-based CLI binary that provides audio notifications for Claude Code events.

**GitHub:** mpolatcan/ccbell

## Purpose

Plays sound notifications when Claude finishes responding, needs permission, is waiting for input, or when a subagent completes.

## Architecture

### Plugin Distribution (cc-plugins repo)

The `cc-plugins` repository distributes the plugin with:
- `plugins/ccbell/.claude-plugin/plugin.json` - Plugin manifest with commands and hooks
- `plugins/ccbell/sounds/*.aiff` - Audio files for each event type
- `plugins/ccbell/commands/*.md` - Slash command documentation
- `plugins/ccbell/scripts/ccbell.sh` - Downloader script for the Go binary

### Binary Distribution (ccbell repo)

The `ccbell` repository builds and releases Go binaries via GitHub Actions.

**Note:** Binary installation is automatically handled by the `ccbell.sh` script in the cc-plugins repository, which downloads the correct Go binary for your platform from GitHub releases.

## Official Documentation (ALWAYS refer to these)

**IMPORTANT:** Plugin schemas and hook events may change. Always consult the official documentation for the latest specifications.

| Documentation | URL |
|--------------|-----|
| Plugins Overview | https://code.claude.com/docs/en/plugins |
| Discover Plugins | https://code.claude.com/docs/en/discover-plugins |
| **Plugins Reference** | https://code.claude.com/docs/en/plugins-reference |
| **Hooks Reference** | https://code.claude.com/docs/en/hooks |
| Plugin Marketplaces | https://code.claude.com/docs/en/plugin-marketplaces |

### Key Documentation Sections

- **Plugin Manifest** - See [Plugins Reference: Manifest](https://code.claude.com/docs/en/plugins-reference#manifest)
- **Commands** - See [Plugins Reference: Commands](https://code.claude.com/docs/en/plugins-reference#commands)
- **Hooks** - See [Hooks Reference](https://code.claude.com/docs/en/hooks) and [Plugins Reference: Hooks](https://code.claude.com/docs/en/plugins-reference#hooks)

### Hooks Documentation (ALWAYS refer to these)

**Main Hooks Documentation:**
- **Hooks Reference** - https://code.claude.com/docs/en/hooks
  - **Events** - https://code.claude.com/docs/en/hooks#events (Stop, PermissionPrompt, Notification, UserPromptSubmit, SubagentStop, etc.)
  - **Hook Types** - https://code.claude.com/docs/en/hooks#hook-types (command, agent, skill)
  - **Matcher** - https://code.claude.com/docs/en/hooks#matcher
  - **Timeout** - https://code.claude.com/docs/en/hooks#timeout

**Related in Plugins Reference:**
- **Hooks in Plugins** - https://code.claude.com/docs/en/plugins-reference#hooks

**ALWAYS consult these pages for:**
- Complete list of available hook events
- Hook type specifications (command, agent, skill)
- Matcher patterns and syntax
- Timeout configuration
- Input/output handling

### Hook Events Used

| Event | Description |
|-------|-------------|
| `Stop` | Claude finishes responding |
| `Notification` (permission_prompt) | Claude needs permission |
| `Notification` (idle_prompt) | User waiting for input |
| `SubagentStop` | Subagent task completes |

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
