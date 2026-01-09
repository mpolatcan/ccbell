---
description: Show current ccbell configuration status
allowed-tools: ["Read", "Bash"]
---

# ccbell Status

Show current configuration and status.

## Instructions

### 1. Read Configuration

Check for config at:
- Project: `.claude/ccbell.config.json`
- Global: `~/.claude/ccbell.config.json`

### 2. Display Status

If config exists, parse and display:

```
## ccbell Status

**Global Status:** Enabled/Disabled
**Config Location:** ~/.claude/ccbell.config.json

### Event Configuration

| Event | Enabled | Sound | Volume |
|-------|---------|-------|--------|
| Stop | Yes | Glass (system) | 50% |
| Permission Prompt | Yes | Ping (system) | 70% |
| Idle Prompt | No | - | - |

### Quick Commands

- `/ccbell:enable` - Enable all notifications
- `/ccbell:disable` - Disable all notifications
- `/ccbell:configure` - Change sound settings
- `/ccbell:test` - Test sounds
```

### 3. If No Config

If no config file exists:

```
## ccbell Status

**Status:** Not configured

No configuration file found. ccbell will use default settings:
- All events enabled
- System sounds (Glass, Ping, Submarine)
- 50% volume

Run /ccbell:configure to set up your preferences.
Run /ccbell:enable to create a default config.
```
