package config

import (
	"testing"
	"time"
)

func TestIsInQuietHours(t *testing.T) {
	// Get current time for testing
	now := time.Now()
	currentHour := now.Hour()

	tests := []struct {
		name       string
		quietHours *QuietHours
		want       bool
	}{
		{
			name:       "no quiet hours configured",
			quietHours: nil,
			want:       false,
		},
		{
			name:       "empty start",
			quietHours: &QuietHours{Start: "", End: "07:00"},
			want:       false,
		},
		{
			name:       "empty end",
			quietHours: &QuietHours{Start: "22:00", End: ""},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{QuietHours: tt.quietHours}
			if got := cfg.IsInQuietHours(); got != tt.want {
				t.Errorf("IsInQuietHours() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test time-dependent cases
	t.Run("currently in quiet hours (same day)", func(t *testing.T) {
		// Set quiet hours to include current time
		startHour := (currentHour - 1 + 24) % 24
		endHour := (currentHour + 1) % 24

		cfg := &Config{
			QuietHours: &QuietHours{
				Start: formatTime(startHour, 0),
				End:   formatTime(endHour, 0),
			},
		}

		// Only valid if start < end (same day)
		if startHour < endHour {
			if !cfg.IsInQuietHours() {
				t.Error("expected to be in quiet hours")
			}
		}
	})

	t.Run("currently outside quiet hours", func(t *testing.T) {
		// Set quiet hours to definitely not include current time
		startHour := (currentHour + 2) % 24
		endHour := (currentHour + 4) % 24

		cfg := &Config{
			QuietHours: &QuietHours{
				Start: formatTime(startHour, 0),
				End:   formatTime(endHour, 0),
			},
		}

		// Only valid if this doesn't wrap around to include current time
		if startHour < endHour && currentHour < startHour {
			if cfg.IsInQuietHours() {
				t.Error("expected to not be in quiet hours")
			}
		}
	})

	t.Run("overnight quiet hours - before midnight", func(t *testing.T) {
		cfg := &Config{
			QuietHours: &QuietHours{
				Start: "22:00",
				End:   "07:00",
			},
		}

		// Test at 23:00 - should be in quiet hours
		if currentHour == 23 {
			if !cfg.IsInQuietHours() {
				t.Error("23:00 should be in quiet hours (22:00-07:00)")
			}
		}

		// Test at 06:00 - should be in quiet hours
		if currentHour == 6 {
			if !cfg.IsInQuietHours() {
				t.Error("06:00 should be in quiet hours (22:00-07:00)")
			}
		}

		// Test at 12:00 - should NOT be in quiet hours
		if currentHour == 12 {
			if cfg.IsInQuietHours() {
				t.Error("12:00 should NOT be in quiet hours (22:00-07:00)")
			}
		}
	})
}

func TestParseTimeToMinutes(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"00:00", 0, false},
		{"01:30", 90, false},
		{"12:00", 720, false},
		{"23:59", 1439, false},
		{"22:00", 1320, false},
		{"07:00", 420, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseTimeToMinutes(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimeToMinutes(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseTimeToMinutes(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func formatTime(hour, _ int) string {
	return padZero(hour) + ":00"
}

func padZero(n int) string {
	if n < 0 || n > 99 {
		return "00" // Safe fallback for invalid values
	}
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
