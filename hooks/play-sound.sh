#!/bin/bash

# ccbell - Sound notification hook for Claude Code
# Usage: play-sound.sh <event_type>
# Event types: stop, permission_prompt, idle_prompt

set -euo pipefail

EVENT_TYPE="${1:-stop}"
PLUGIN_ROOT="${CLAUDE_PLUGIN_ROOT:-$HOME/.claude/plugins/local/ccbell}"
CONFIG_FILE="${CLAUDE_PROJECT_DIR:-.}/.claude/ccbell.config.json"
GLOBAL_CONFIG="$HOME/.claude/ccbell.config.json"

# Read stdin (hook input) - required for hooks but don't block
# Use timeout to prevent hanging when stdin is open but empty
if [[ ! -t 0 ]]; then
    timeout 0.1 cat > /dev/null 2>&1 || true
fi

# Function to read config value using jq (if available) or defaults
get_config() {
    local key="$1"
    local default="$2"
    local config_path=""

    # Check project config first, then global
    if [[ -f "$CONFIG_FILE" ]]; then
        config_path="$CONFIG_FILE"
    elif [[ -f "$GLOBAL_CONFIG" ]]; then
        config_path="$GLOBAL_CONFIG"
    else
        echo "$default"
        return
    fi

    # Try jq if available
    if command -v jq &>/dev/null; then
        local value
        value=$(jq -r "$key // empty" "$config_path" 2>/dev/null || echo "")
        echo "${value:-$default}"
    else
        # Fallback: just use default
        echo "$default"
    fi
}

# Check if plugin is enabled globally
ENABLED=$(get_config ".enabled" "true")
if [[ "$ENABLED" != "true" ]]; then
    exit 0
fi

# Check if this event is enabled
EVENT_ENABLED=$(get_config ".events.${EVENT_TYPE}.enabled" "true")
if [[ "$EVENT_ENABLED" != "true" ]]; then
    exit 0
fi

# Get sound configuration for this event
SOUND=$(get_config ".events.${EVENT_TYPE}.sound" "bundled:${EVENT_TYPE}")
VOLUME=$(get_config ".events.${EVENT_TYPE}.volume" "0.5")

# Resolve sound path
resolve_sound_path() {
    local sound_spec="$1"

    case "$sound_spec" in
        system:*)
            # macOS system sound
            local sound_name="${sound_spec#system:}"
            echo "/System/Library/Sounds/${sound_name}.aiff"
            ;;
        custom:*)
            # Custom file path
            echo "${sound_spec#custom:}"
            ;;
        bundled:*)
            # Bundled default sound
            local bundled_name="${sound_spec#bundled:}"
            echo "${PLUGIN_ROOT}/sounds/${bundled_name}.aiff"
            ;;
        *)
            # Assume it's a direct path
            echo "$sound_spec"
            ;;
    esac
}

SOUND_PATH=$(resolve_sound_path "$SOUND")

# Check if sound file exists
if [[ ! -f "$SOUND_PATH" ]]; then
    # Fallback to default bundled sound
    SOUND_PATH="${PLUGIN_ROOT}/sounds/${EVENT_TYPE}.aiff"
    if [[ ! -f "$SOUND_PATH" ]]; then
        # Final fallback: system default
        SOUND_PATH="/System/Library/Sounds/Glass.aiff"
    fi
fi

# Play sound based on platform
play_sound() {
    local file="$1"
    local vol="$2"

    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS - use afplay
        afplay -v "$vol" "$file" &>/dev/null &
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux - try multiple players
        if command -v paplay &>/dev/null; then
            paplay "$file" &>/dev/null &
        elif command -v aplay &>/dev/null; then
            aplay -q "$file" &>/dev/null &
        elif command -v mpv &>/dev/null; then
            mpv --really-quiet --volume="$(echo "$vol * 100" | bc)" "$file" &>/dev/null &
        elif command -v ffplay &>/dev/null; then
            ffplay -nodisp -autoexit -volume "$(echo "$vol * 100" | bc)" "$file" &>/dev/null &
        fi
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "win32" ]]; then
        # Windows
        powershell -c "(New-Object Media.SoundPlayer '$file').Play()" &>/dev/null &
    fi
}

# Play the sound
play_sound "$SOUND_PATH" "$VOLUME"

# Exit successfully (don't block Claude)
exit 0
