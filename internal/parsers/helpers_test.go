package parsers

import (
	"net/http"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func float64Ptr(v float64) *float64 { return &v }

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input string
		want  *float64
	}{
		{"100", float64Ptr(100)},
		{"3.14", float64Ptr(3.14)},
		{"", nil},
		{"abc", nil},
		{" 42 ", float64Ptr(42)},
	}

	for _, tt := range tests {
		got := ParseFloat(tt.input)
		if tt.want == nil {
			if got != nil {
				t.Errorf("ParseFloat(%q) = %v, want nil", tt.input, *got)
			}
		} else {
			if got == nil {
				t.Errorf("ParseFloat(%q) = nil, want %v", tt.input, *tt.want)
			} else if *got != *tt.want {
				t.Errorf("ParseFloat(%q) = %v, want %v", tt.input, *got, *tt.want)
			}
		}
	}
}

func TestParseResetTime(t *testing.T) {
	ts := ParseResetTime("1700000000")
	if ts == nil {
		t.Fatal("expected non-nil for unix timestamp")
	}
	expected := time.Unix(1700000000, 0)
	if !ts.Equal(expected) {
		t.Errorf("got %v, want %v", ts, expected)
	}

	ts = ParseResetTime("2025-01-01T00:00:00Z")
	if ts == nil {
		t.Fatal("expected non-nil for RFC3339")
	}

	before := time.Now()
	ts = ParseResetTime("30s")
	if ts == nil {
		t.Fatal("expected non-nil for duration")
	}
	if ts.Before(before.Add(29 * time.Second)) {
		t.Error("duration parse too far in past")
	}

	ts = ParseResetTime("")
	if ts != nil {
		t.Error("expected nil for empty")
	}
}

func TestRedactHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("Authorization", "Bearer sk-1234567890abcdef")
	h.Set("Content-Type", "application/json")
	h.Set("X-RateLimit-Remaining", "42")

	redacted := RedactHeaders(h)

	if redacted["Authorization"] == "Bearer sk-1234567890abcdef" {
		t.Error("Authorization should be redacted")
	}
	if redacted["Content-Type"] != "application/json" {
		t.Error("Content-Type should not be redacted")
	}
	if redacted["X-Ratelimit-Remaining"] != "42" {
		t.Errorf("X-RateLimit-Remaining = %q, want '42'", redacted["X-Ratelimit-Remaining"])
	}
}

func TestApplyRateLimitGroup(t *testing.T) {
	t.Run("uninitialized snapshot with limit, remaining, and reset headers", func(t *testing.T) {
		snap := &core.UsageSnapshot{}
		h := http.Header{}
		h.Set("X-RateLimit-Limit", "1000")
		h.Set("X-RateLimit-Remaining", "500")
		h.Set("X-RateLimit-Reset", "1700000000")

		ApplyRateLimitGroup(h, snap, "requests", "req", "1m", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset")

		if snap.Metrics == nil {
			t.Fatal("expected snap.Metrics to be initialized")
		}
		if snap.Resets == nil {
			t.Fatal("expected snap.Resets to be initialized")
		}
		metric, ok := snap.Metrics["requests"]
		if !ok {
			t.Fatal("expected snap.Metrics['requests'] to exist")
		}
		if metric.Limit == nil || *metric.Limit != 1000 {
			t.Errorf("got limit %v, want 1000", metric.Limit)
		}
		if metric.Remaining == nil || *metric.Remaining != 500 {
			t.Errorf("got remaining %v, want 500", metric.Remaining)
		}
		if metric.Unit != "req" || metric.Window != "1m" {
			t.Errorf("got unit %q window %q, want 'req' and '1m'", metric.Unit, metric.Window)
		}
		reset, ok := snap.Resets["requests_reset"]
		if !ok {
			t.Fatal("expected snap.Resets['requests_reset'] to exist")
		}
		expectedReset := time.Unix(1700000000, 0)
		if !reset.Equal(expectedReset) {
			t.Errorf("got reset %v, want %v", reset, expectedReset)
		}
	})

	t.Run("nil snapshot does not panic", func(t *testing.T) {
		h := http.Header{}
		h.Set("X-RateLimit-Limit", "1000")
		h.Set("X-RateLimit-Remaining", "500")
		h.Set("X-RateLimit-Reset", "1700000000")

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("unexpected panic on nil snapshot: %v", r)
			}
		}()

		ApplyRateLimitGroup(h, nil, "requests", "req", "1m", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset")
	})

	t.Run("header without rate limits returns gracefully without modifying empty snapshot", func(t *testing.T) {
		snap := &core.UsageSnapshot{}
		h := http.Header{}

		ApplyRateLimitGroup(h, snap, "requests", "req", "1m", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset")

		if snap.Metrics != nil {
			t.Errorf("expected snap.Metrics to remain nil, got %v", snap.Metrics)
		}
		if snap.Resets != nil {
			t.Errorf("expected snap.Resets to remain nil, got %v", snap.Resets)
		}
	})
}

