---
description: Configure ccbell sound notifications for different events
allowed-tools: ["Read", "Write", "AskUserQuestion", "Bash"]
---

# Configure ccbell Sound Notifications

Help the user configure sound notifications for Claude Code events.

## Available macOS System Sounds

First, let the user know the available system sounds on macOS:
- Basso, Blow, Bottle, Frog, Funk, Glass, Hero, Morse, Ping, Pop, Purr, Sosumi, Submarine, Tink

## Configuration Steps

### 1. Ask User for Event Selection

Use AskUserQuestion to ask which events should trigger sounds:

```json
{
  "questions": [
    {
      "question": "Which events should trigger sound notifications?",
      "header": "Events",
      "multiSelect": true,
      "options": [
        {"label": "Stop (Claude finishes)", "description": "Play sound when Claude completes responding"},
        {"label": "Permission Prompt", "description": "Play sound when Claude needs permission"},
        {"label": "Idle Prompt", "description": "Play sound when Claude is waiting for input"}
      ]
    }
  ]
}
```

### 2. For Each Selected Event, Ask Sound Choice

For each enabled event, ask the user to choose a sound:

```json
{
  "questions": [
    {
      "question": "Choose sound for the Stop event:",
      "header": "Stop sound",
      "options": [
        {"label": "Glass (Recommended)", "description": "Soft glass chime - system:Glass"},
        {"label": "Ping", "description": "Quick ping sound - system:Ping"},
        {"label": "Hero", "description": "Achievement sound - system:Hero"},
        {"label": "Submarine", "description": "Sonar ping - system:Submarine"}
      ]
    }
  ]
}
```

### 3. Ask for Volume

```json
{
  "questions": [
    {
      "question": "What volume level? (0.0 to 1.0)",
      "header": "Volume",
      "options": [
        {"label": "Low (0.3)", "description": "Quiet notifications"},
        {"label": "Medium (0.5)", "description": "Balanced volume"},
        {"label": "High (0.7)", "description": "Louder notifications"},
        {"label": "Full (1.0)", "description": "Maximum volume"}
      ]
    }
  ]
}
```

### 4. Write Configuration

Create the configuration file at `.claude/ccbell.config.json` (project) or `~/.claude/ccbell.config.json` (global):

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

### 5. Confirm Configuration

After writing the config, confirm to the user and offer to test the sounds with `/ccbell:test`.

## Sound Specification Formats

- `system:Glass` - macOS system sound
- `custom:/path/to/sound.mp3` - Custom audio file
- `bundled:stop` - Bundled default sound
