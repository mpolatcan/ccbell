package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// IsInQuietHours checks if the current time is within quiet hours.
func (c *Config) IsInQuietHours() bool {
	if c.QuietHours == nil || c.QuietHours.Start == "" || c.QuietHours.End == "" {
		return false
	}

	startMins, err1 := parseTimeToMinutes(c.QuietHours.Start)
	endMins, err2 := parseTimeToMinutes(c.QuietHours.End)
	if err1 != nil || err2 != nil {
		return false // Invalid format, don't block
	}

	now := time.Now()
	currentMins := now.Hour()*60 + now.Minute()

	// Handle start == end (24-hour quiet period, meaning quiet hours disabled)
	if startMins == endMins {
		return false
	}

	// Handle overnight periods (e.g., 22:00 - 07:00)
	if startMins > endMins {
		// Quiet hours span midnight
		return currentMins >= startMins || currentMins < endMins
	}

	// Normal period (e.g., 09:00 - 17:00)
	return currentMins >= startMins && currentMins < endMins
}

// parseTimeToMinutes converts "HH:MM" to minutes since midnight.
func parseTimeToMinutes(timeStr string) (int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time format: %q (expected HH:MM)", timeStr)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hours in time %q: %w", timeStr, err)
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes in time %q: %w", timeStr, err)
	}

	if hours < 0 || hours > 23 || minutes < 0 || minutes > 59 {
		return 0, fmt.Errorf("time out of range: %q", timeStr)
	}

	return hours*60 + minutes, nil
}

// QuietHoursStatus returns human-readable quiet hours status.
func (c *Config) QuietHoursStatus() string {
	if c.QuietHours == nil || c.QuietHours.Start == "" || c.QuietHours.End == "" {
		return "not configured"
	}

	if c.IsInQuietHours() {
		return "active (currently in quiet period)"
	}

	return "configured but not active"
}
