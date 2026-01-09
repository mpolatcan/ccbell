---
description: Show ccbell help and documentation
---

# ccbell Help

Show help and documentation for the ccbell plugin.

## Response

Provide the following information:

```
# ccbell - Sound Notifications for Claude Code

Play audio notifications when Claude Code events occur, so you can work in
other windows without constantly checking if Claude is done.

## Supported Events

- **Stop** - Plays when Claude finishes responding
- **Permission Prompt** - Plays when Claude needs your permission
- **Idle Prompt** - Plays when Claude is waiting for your input

## Commands

| Command | Description |
|---------|-------------|
| /ccbell:configure | Interactive setup for sounds and events |
| /ccbell:test [event] | Test sounds (all or specific event) |
| /ccbell:enable | Enable all notifications |
| /ccbell:disable | Disable all notifications |
| /ccbell:status | Show current configuration |
| /ccbell:help | Show this help |

## Sound Options

### System Sounds (macOS)
Basso, Blow, Bottle, Frog, Funk, Glass, Hero, Morse, Ping, Pop, Purr,
Sosumi, Submarine, Tink

Format: `system:Glass`

### Custom Sounds
Use your own MP3, WAV, AIFF, or M4A files.

Format: `custom:/path/to/your/sound.mp3`

### Bundled Defaults
Pre-packaged sounds included with the plugin.

Format: `bundled:stop`, `bundled:permission_prompt`, `bundled:idle_prompt`

## Configuration

Config is stored at:
- Project: `.claude/ccbell.config.json`
- Global: `~/.claude/ccbell.config.json`

Project config takes precedence over global config.

## Cross-Platform Support

- **macOS:** Full support (uses afplay)
- **Linux:** Requires paplay, aplay, mpv, or ffplay
- **Windows:** Uses PowerShell Media.SoundPlayer

## Quick Start

1. Run `/ccbell:enable` to enable with defaults
2. Run `/ccbell:test` to verify sounds work
3. Run `/ccbell:configure` to customize

## Troubleshooting

**Commands not found?**
If via marketplace:
1. Check marketplace: `/plugin marketplace list`
2. Reinstall: `/plugin install ccbell`

If manual install:
1. Verify plugin at `~/.claude/plugins/local/ccbell`
2. Check settings contain: `"ccbell@local": true`
3. Restart Claude Code

**Sounds not playing?**
1. Check volume isn't muted
2. Run `/ccbell:status` to verify config
3. Test manually: `afplay /System/Library/Sounds/Glass.aiff`

**Wrong sounds playing?**
Run `/ccbell:configure` to update your preferences.

## Installation

**Via Marketplace (recommended):**
/plugin marketplace add mpolatcan/ccbell
/plugin install ccbell

**Manual installation:**
git clone https://github.com/mpolatcan/ccbell.git ~/.claude/plugins/local/ccbell
Then add "ccbell@local": true to ~/.claude/settings.json
```
