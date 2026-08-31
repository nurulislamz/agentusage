package tui

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{
			name: "negative duration treated as 0s",
			d:    -10 * time.Second,
			want: "0s",
		},
		{
			name: "zero duration",
			d:    0,
			want: "0s",
		},
		{
			name: "single second",
			d:    1 * time.Second,
			want: "1s",
		},
		{
			name: "4 seconds",
			d:    4 * time.Second,
			want: "4s",
		},
		{
			name: "5 seconds",
			d:    5 * time.Second,
			want: "5s",
		},
		{
			name: "45 seconds",
			d:    45 * time.Second,
			want: "45s",
		},
		{
			name: "59 seconds",
			d:    59 * time.Second,
			want: "59s",
		},
		{
			name: "exact 1 minute",
			d:    1 * time.Minute,
			want: "1m0s",
		},
		{
			name: "1 minute 30 seconds",
			d:    1*time.Minute + 30*time.Second,
			want: "1m30s",
		},
		{
			name: "5 minutes 5 seconds",
			d:    5*time.Minute + 5*time.Second,
			want: "5m5s",
		},
		{
			name: "59 minutes 59 seconds",
			d:    59*time.Minute + 59*time.Second,
			want: "59m59s",
		},
		{
			name: "exact 1 hour",
			d:    1 * time.Hour,
			want: "1h0m",
		},
		{
			name: "1 hour 15 minutes",
			d:    1*time.Hour + 15*time.Minute,
			want: "1h15m",
		},
		{
			name: "23 hours 59 minutes",
			d:    23*time.Hour + 59*time.Minute,
			want: "23h59m",
		},
		{
			name: "exact 1 day",
			d:    24 * time.Hour,
			want: "1d0h",
		},
		{
			name: "1 day 5 hours",
			d:    29 * time.Hour,
			want: "1d5h",
		},
		{
			name: "2 days exact",
			d:    48 * time.Hour,
			want: "2d0h",
		},
		{
			name: "3 days 18 hours",
			d:    90 * time.Hour,
			want: "3d18h",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatDuration(tc.d)
			if got != tc.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tc.d, got, tc.want)
			}
		})
	}
}

func TestFormatLastRefreshed(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		timestamp time.Time
		want      string
	}{
		{
			name:      "zero timestamp returns empty string",
			timestamp: time.Time{},
			want:      "",
		},
		{
			name:      "future timestamp treated as just now",
			timestamp: now.Add(10 * time.Second),
			want:      "Last refreshed just now",
		},
		{
			name:      "exact now",
			timestamp: now,
			want:      "Last refreshed just now",
		},
		{
			name:      "2 seconds ago (< 5s)",
			timestamp: now.Add(-2 * time.Second),
			want:      "Last refreshed just now",
		},
		{
			name:      "4 seconds ago (< 5s)",
			timestamp: now.Add(-4 * time.Second),
			want:      "Last refreshed just now",
		},
		{
			name:      "5 seconds ago (threshold)",
			timestamp: now.Add(-5 * time.Second),
			want:      "Last refreshed 5s ago",
		},
		{
			name:      "30 seconds ago",
			timestamp: now.Add(-30 * time.Second),
			want:      "Last refreshed 30s ago",
		},
		{
			name:      "59 seconds ago",
			timestamp: now.Add(-59 * time.Second),
			want:      "Last refreshed 59s ago",
		},
		{
			name:      "1 minute ago",
			timestamp: now.Add(-1 * time.Minute),
			want:      "Last refreshed 1m0s ago",
		},
		{
			name:      "2 minutes 15 seconds ago",
			timestamp: now.Add(-2*time.Minute - 15*time.Second),
			want:      "Last refreshed 2m15s ago",
		},
		{
			name:      "10 minutes ago",
			timestamp: now.Add(-10 * time.Minute),
			want:      "Last refreshed 10m0s ago",
		},
		{
			name:      "1 hour ago",
			timestamp: now.Add(-1 * time.Hour),
			want:      "Last refreshed 1h0m ago",
		},
		{
			name:      "3 hours 20 minutes ago",
			timestamp: now.Add(-3*time.Hour - 20*time.Minute),
			want:      "Last refreshed 3h20m ago",
		},
		{
			name:      "1 day ago",
			timestamp: now.Add(-24 * time.Hour),
			want:      "Last refreshed 1d0h ago",
		},
		{
			name:      "2 days 2 hours ago",
			timestamp: now.Add(-50 * time.Hour),
			want:      "Last refreshed 2d2h ago",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatLastRefreshed(tc.timestamp, now)
			if got != tc.want {
				t.Errorf("formatLastRefreshed(%v, %v) = %q, want %q", tc.timestamp, now, got, tc.want)
			}
		})
	}
}

func TestTitleCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "whitespace only",
			input: "   ",
			want:  "",
		},
		{
			name:  "single ASCII lowercase",
			input: "a",
			want:  "A",
		},
		{
			name:  "single ASCII uppercase",
			input: "A",
			want:  "A",
		},
		{
			name:  "single multibyte rune lowercase",
			input: "ü",
			want:  "Ü",
		},
		{
			name:  "single multibyte rune uppercase",
			input: "Ü",
			want:  "Ü",
		},
		{
			name:  "normal word",
			input: "hello",
			want:  "Hello",
		},
		{
			name:  "multibyte word ecran",
			input: "écran",
			want:  "Écran",
		},
		{
			name:  "multibyte word uber",
			input: "über",
			want:  "Über",
		},
		{
			name:  "mixed casing",
			input: "hELLO",
			want:  "Hello",
		},
		{
			name:  "mixed casing with multibyte",
			input: "éCRAN",
			want:  "Écran",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := titleCase(tc.input)
			if got != tc.want {
				t.Errorf("titleCase(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

