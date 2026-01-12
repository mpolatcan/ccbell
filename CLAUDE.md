# ccbell

Go-based CLI binary that provides audio notifications for Claude Code events.

**GitHub:** mpolatcan/ccbell

## Purpose

Plays sound notifications when Claude finishes responding, needs permission, or when a subagent completes.

## Build Commands

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

Event types: `stop`, `permission_prompt`, `subagent`

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
