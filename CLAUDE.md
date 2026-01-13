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

**CRITICAL: Follow these steps in order. DO NOT SKIP any step.**

### Step 1: Run Local Checks
```bash
make check
```
This runs `fmt`, `lint`, `test`, and `build`. Wait for all checks to pass.

### Step 2: Verify CI Pipeline Passed
Go to https://github.com/mpolatcan/ccbell/actions/workflows/ci.yml

- Find the latest workflow run
- Confirm ALL checks passed (green checkmarks)
- **DO NOT PROCEED if CI failed** - fix issues first

### Step 3: Create Git Tag
```bash
git tag vX.Y.Z
```
Replace `X.Y.Z` with the new version number.

### Step 4: Push Tag
```bash
git push origin vX.Y.Z
```

### Step 5: Wait for GitHub Release
- GitHub Actions will automatically build and create the release
- Monitor: https://github.com/mpolatcan/ccbell/releases
- Wait for the release asset to be uploaded

### Step 6: Sync Version to cc-plugins
```bash
cd ../ccbell
make sync-version VERSION=vX.Y.Z
```

### Step 7: Commit and Push cc-plugins
```bash
cd ../cc-plugins
git add plugins/ccbell/.claude-plugin/plugin.json plugins/ccbell/scripts/ccbell.sh
git commit -m "chore(ccbell): sync version to vX.Y.Z"
git push
```
