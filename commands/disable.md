---
description: Disable ccbell sound notifications
allowed-tools: ["Read", "Write"]
---

# Disable ccbell

Disable ccbell sound notifications.

## Instructions

### 1. Find Configuration File

Check for config at:
- Project: `.claude/ccbell.config.json`
- Global: `~/.claude/ccbell.config.json`

### 2. Update Configuration

Read the existing config and set `enabled: false`.

If no config exists, create one with `enabled: false`:

```json
{
  "enabled": false,
  "events": {
    "stop": { "enabled": true, "sound": "system:Glass", "volume": 0.5 },
    "permission_prompt": { "enabled": true, "sound": "system:Ping", "volume": 0.7 },
    "idle_prompt": { "enabled": true, "sound": "system:Submarine", "volume": 0.5 }
  }
}
```

### 3. Write Configuration

Write the updated config.

### 4. Confirm

Tell the user:
```
ccbell sound notifications disabled.

No sounds will play until you run /ccbell:enable.
Your sound preferences have been preserved.
```
