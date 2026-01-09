# ccbell

Audio notifications for Claude Code events. Get notified when Claude finishes responding, needs permission, or is waiting for your input.

## Features

- Play sounds on multiple Claude Code events
- Support for system sounds, custom audio files, and bundled defaults
- Cross-platform support (macOS, Linux, Windows)
- Easy configuration via slash commands
- Per-event volume control

## Installation

### Option 1: Via Marketplace (Recommended)

The easiest way to install ccbell is through the Claude Code plugin marketplace:

**Step 1: Add the marketplace**
```
/plugin marketplace add mpolatcan/claude-plugins
```

**Step 2: Install the plugin**
```
/plugin install ccbell
```

That's it! The plugin is now installed and ready to use.

### Option 2: Manual Installation

For local development or if you prefer manual installation:

**Step 1: Clone to Local Plugins**

```bash
git clone https://github.com/mpolatcan/ccbell.git ~/.claude/plugins/local/ccbell
```

**Step 2: Make Scripts Executable**

```bash
chmod +x ~/.claude/plugins/local/ccbell/hooks/play-sound.sh
```

**Step 3: Enable the Plugin**

Add to `~/.claude/settings.json`:

```json
{
  "enabledPlugins": {
    "ccbell@local": true
  }
}
```

**Step 4: Restart Claude Code**

Restart Claude Code to load the plugin.

## Quick Start

1. Enable the plugin:
   ```
   /ccbell:enable
   ```

2. Test sounds:
   ```
   /ccbell:test
   ```

3. Customize (optional):
   ```
   /ccbell:configure
   ```

## Commands

| Command | Description |
|---------|-------------|
| `/ccbell:configure` | Interactive setup for sounds and events |
| `/ccbell:test [event]` | Test sounds (all or specific event) |
| `/ccbell:enable` | Enable all notifications |
| `/ccbell:disable` | Disable all notifications |
| `/ccbell:status` | Show current configuration |
| `/ccbell:help` | Show help documentation |

## Supported Events

| Event | When it triggers |
|-------|-----------------|
| `stop` | Claude finishes responding |
| `permission_prompt` | Claude needs your permission to proceed |
| `idle_prompt` | Claude is waiting for your input |

## Sound Options

### System Sounds (macOS)

Available sounds: Basso, Blow, Bottle, Frog, Funk, Glass, Hero, Morse, Ping, Pop, Purr, Sosumi, Submarine, Tink

Format: `system:Glass`

### Custom Sounds

Use your own audio files (MP3, WAV, AIFF, M4A).

Format: `custom:/path/to/your/sound.mp3`

### Bundled Defaults

Pre-packaged sounds included with the plugin.

Format: `bundled:stop`, `bundled:permission_prompt`, `bundled:idle_prompt`

## Configuration

Configuration is stored in JSON format at:
- **Project-level:** `.claude/ccbell.config.json`
- **Global:** `~/.claude/ccbell.config.json`

Project config takes precedence over global config.

### Example Configuration

```json
{
  "enabled": true,
  "events": {
    "stop": {
      "enabled": true,
      "sound": "system:Glass",
      "volume": 0.5
    },
    "permission_prompt": {
      "enabled": true,
      "sound": "system:Ping",
      "volume": 0.7
    },
    "idle_prompt": {
      "enabled": false,
      "sound": "system:Submarine",
      "volume": 0.5
    }
  }
}
```

## Cross-Platform Support

| Platform | Audio Backend | Status |
|----------|--------------|--------|
| macOS | `afplay` | Full support |
| Linux | `paplay`, `aplay`, `mpv`, `ffplay` | Requires one installed |
| Windows | PowerShell `Media.SoundPlayer` | Basic support |

## Troubleshooting

### Commands not found?

**If installed via marketplace:**
1. Verify the marketplace is added: `/plugin marketplace list`
2. Check if the plugin is installed: `/plugins`
3. Try reinstalling: `/plugin install ccbell`

**If installed manually:**
1. Verify the plugin is installed at `~/.claude/plugins/local/ccbell`
2. Check that the plugin is enabled in your settings:
   ```bash
   cat ~/.claude/settings.json | grep ccbell
   ```
3. Ensure the settings contain `"ccbell@local": true`
4. Restart Claude Code after enabling

### Sounds not playing?

1. Check your system volume isn't muted
2. Verify the plugin is enabled: `/ccbell:status`
3. Test audio manually:
   ```bash
   afplay /System/Library/Sounds/Glass.aiff
   ```

### Wrong sounds playing?

Run `/ccbell:configure` to update your sound preferences.

### Permission errors?

Make sure the script is executable:
```bash
chmod +x ~/.claude/plugins/local/ccbell/hooks/play-sound.sh
```

## Repository Structure

```
ccbell/
├── .claude-plugin/
│   └── plugin.json          # Plugin manifest
├── commands/
│   ├── configure.md         # /ccbell:configure
│   ├── disable.md           # /ccbell:disable
│   ├── enable.md            # /ccbell:enable
│   ├── help.md              # /ccbell:help
│   ├── status.md            # /ccbell:status
│   └── test.md              # /ccbell:test
├── hooks/
│   ├── hooks.json           # Hook event configuration
│   └── play-sound.sh        # Cross-platform sound player
├── sounds/
│   ├── idle_prompt.aiff     # Default idle sound
│   ├── permission_prompt.aiff # Default permission sound
│   └── stop.aiff            # Default stop sound
└── README.md
```

## Uninstallation

### If installed via marketplace:

```
/plugin uninstall ccbell
```

Optionally remove the marketplace:
```
/plugin marketplace remove mpolatcan/claude-plugins
```

### If installed manually:

**Step 1: Disable the Plugin**

Remove `"ccbell@local": true` from your settings file:
- Global: `~/.claude/settings.json`
- Or per-project: `.claude/settings.local.json`

**Step 2: Remove Plugin Files**

```bash
rm -rf ~/.claude/plugins/local/ccbell
```

### Remove Configuration (Optional)

```bash
rm ~/.claude/ccbell.config.json
rm .claude/ccbell.config.json  # If project-level config exists
```

## License

MIT
