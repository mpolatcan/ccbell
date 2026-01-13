# ccbell :bell:

Go-based CLI binary that provides audio notifications for Claude Code events.

**:octocat: GitHub:** mpolatcan/ccbell

## Purpose :rocket:

Plays sound notifications when Claude finishes responding, needs permission, is waiting for input, or when a subagent completes.

## Build Commands :hammer:

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

## Usage :speaker:

```bash
ccbell <event_type>
```

Event types: `stop`, `permission_prompt`, `idle_prompt`, `subagent`

## Audio Backends :computer:

| Platform | Backend |
|----------|---------|
| macOS | `afplay` (built-in) |
| Linux | `mpv`, `paplay`, `aplay`, or `ffplay` |

## External Release Check Rule :warning:

**CRITICAL:** Before tagging a new version, ensure the release is complete:

1. Wait for GitHub Actions to finish building and uploading release assets
2. Verify the release appears at https://github.com/mpolatcan/ccbell/releases
3. Only after release is confirmed, run `make sync-version VERSION=vX.Y.Z` (from the ccbell directory where the Makefile is)

**NEVER push version tags unless the release is complete.**

## Release Process :rotating_light:

:warning: **CRITICAL: Follow these steps in order. DO NOT SKIP any step.** :warning:

### Step 1: Run Local Checks :white_check_mark:
```bash
make check
```
This runs `fmt`, `lint`, `test`, and `build`. Wait for all checks to pass.

### Step 2: Verify CI Pipeline Passed :mag:

Use `gh` CLI to check the latest workflow run:

```bash
# Check latest run status
gh run list --workflow=ci.yml --limit 1

# View the run details
gh run view $(gh run list --workflow=ci.yml --limit 1 --json databaseId --jq '.databaseId')
```

- Confirm `status: "COMPLETED"` and `conclusion: "SUCCESS"`
- :no_entry: **DO NOT PROCEED if CI failed** - fix issues first :no_entry:

Or view in browser:
https://github.com/mpolatcan/ccbell/actions/workflows/ci.yml

### Step 3: Create Git Tag :label:
```bash
git tag vX.Y.Z
```
Replace `X.Y.Z` with the new version number.

### Step 4: Push Tag :arrow_up:
```bash
git push origin vX.Y.Z
```

### Step 5: Wait for GitHub Release :hourglass_flowing_sand:
- GitHub Actions will automatically build and create the release
- Monitor: https://github.com/mpolatcan/ccbell/releases
- Wait for the release asset to be uploaded

### Step 6: Sync Version to cc-plugins :arrows_counterclockwise:
```bash
make sync-version VERSION=vX.Y.Z
```

### Step 7: Commit and Push cc-plugins :package:
```bash
cd ../cc-plugins
git add plugins/ccbell/.claude-plugin/plugin.json plugins/ccbell/scripts/ccbell.sh
git commit -m "chore(ccbell): sync version to vX.Y.Z"
git push
```
