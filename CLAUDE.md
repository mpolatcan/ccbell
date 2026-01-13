# ccbell

Go-based CLI binary that provides audio notifications for Claude Code events.

**GitHub:** mpolatcan/ccbell

## Purpose

Plays sound notifications when Claude finishes responding, needs permission, is waiting for input, or when a subagent completes.

## Build Commands

```bash
make build           # Build for current platform
make test            # Run tests with race detection
make coverage        # Generate coverage report
make lint            # Run linter (golangci-lint or go vet)
make fmt             # Format code
make clean           # Remove build artifacts
make install         # Install ccbell binary to plugin directory
make uninstall       # Remove from plugin directory
make dist            # Cross-compile for all platforms (cleans build dir first)
make release         # Build, checksum, and create archives
make check           # Run fmt, lint, test, and build
make dev             # Quick development build
make run             # Build and run with arguments
make sync-version    # Sync version to cc-plugins marketplace
make version         # Show version information
```

## Usage

```bash
ccbell <event_type>
```

Event types: `stop`, `permission_prompt`, `idle_prompt`, `subagent`

## Audio Backends

| Platform | Backend |
|----------|---------|
| macOS | `afplay` (built-in) |
| Linux | `paplay`, `aplay`, `mpv`, or `ffplay` |

## Release Process

1. Create git tag: `git tag vX.Y.Z`
2. Push tag: `git push origin vX.Y.Z`
3. GitHub Actions builds and creates releases at https://github.com/mpolatcan/ccbell/releases
