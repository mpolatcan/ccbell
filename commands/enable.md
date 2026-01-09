---
description: Enable ccbell sound notifications
allowed-tools: ["Read", "Write", "Bash"]
---

# Enable ccbell

Enable ccbell sound notifications globally.

## Instructions

### 1. Find Configuration File

Check for config at:
- Project: `.claude/ccbell.config.json`
- Global: `~/.claude/ccbell.config.json`

### 2. Update or Create Configuration

If config exists, read it and set `enabled: true`.

If no config exists, create a default one:

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
      "enabled": true,
      "sound": "system:Submarine",
      "volume": 0.5
    }
  }
}
```

### 3. Write Configuration

Write the updated config to `~/.claude/ccbell.config.json`.

### 4. Confirm

Tell the user:
```
ccbell sound notifications enabled!

Sounds will play for:
- Stop (Claude finishes responding)
- Permission Prompt (needs approval)
- Idle Prompt (waiting for input)

Run /ccbell:test to verify sounds work.
Run /ccbell:configure to customize sounds.
```
