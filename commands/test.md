---
description: Test ccbell sound notifications
argument-hint: "[stop|permission_prompt|idle_prompt|all]"
allowed-tools: ["Bash", "Read"]
---

# Test ccbell Sounds

Test sound notifications for Claude Code events.

## Arguments

$ARGUMENTS

- `stop` - Test the stop event sound
- `permission_prompt` - Test the permission prompt sound
- `idle_prompt` - Test the idle prompt sound
- `all` or no argument - Test all enabled sounds

## Instructions

### 1. Determine Plugin Location

The plugin is located at `$CLAUDE_LOCAL_PLUGINS_PATH/ccbell` or `~/.claude/plugins/local/ccbell`.

### 2. Test Sounds

Based on the argument provided, run the appropriate test:

**For specific event:**
```bash
"$HOME/.claude/plugins/local/ccbell/hooks/play-sound.sh" <event_name>
```

**For all events:**
```bash
for event in stop permission_prompt idle_prompt; do
  "$HOME/.claude/plugins/local/ccbell/hooks/play-sound.sh" "$event"
  sleep 1
done
```

### 3. Report Results

After testing, report which sounds played:

```
## ccbell Sound Test Results

**Stop Event:** Played Glass.aiff (system)
**Permission Prompt:** Played Ping.aiff (system)
**Idle Prompt:** Disabled - skipped

All enabled sounds working correctly!

To change sounds, run /ccbell:configure
```

### 4. Troubleshooting

If sounds don't play:
1. Check if the plugin is enabled: `cat ~/.claude/ccbell.config.json`
2. Verify sound files exist
3. On macOS, ensure audio output is working: `afplay /System/Library/Sounds/Glass.aiff`
