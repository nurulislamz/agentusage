package shared

import (
	"strings"
	"testing"
)

func TestNormalizeLooseModelName(t *testing.T) {
	if got := NormalizeLooseModelName("  claude-sonnet-4  "); got != "claude-sonnet-4" {
		t.Fatalf("NormalizeLooseModelName(trimmed) = %q", got)
	}
	if got := NormalizeLooseModelName("   "); got != "unknown" {
		t.Fatalf("NormalizeLooseModelName(empty) = %q", got)
	}
}

func TestNormalizeLooseClientName(t *testing.T) {
	if got := NormalizeLooseClientName("  CLI  "); got != "CLI" {
		t.Fatalf("NormalizeLooseClientName(trimmed) = %q", got)
	}
	if got := NormalizeLooseClientName(""); got != "Other" {
		t.Fatalf("NormalizeLooseClientName(empty) = %q", got)
	}
}

func TestSanitizeMetricName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{" GPT-4.1 / Mini ", "gpt_4_1_mini"},
		{"", "unknown"},
		{"   ", "unknown"},
		{"---###@@@", "unknown"},
		{"___model___", "model"},
		{"simple123", "simple123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := SanitizeMetricName(tt.input); got != tt.want {
				t.Errorf("SanitizeMetricName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSummarizeShareUsage(t *testing.T) {
	t.Run("basic sorting and normalization", func(t *testing.T) {
		got := SummarizeShareUsage(map[string]float64{
			"beta":  25,
			"alpha": 75,
			"zero":  0,
		}, 2, func(name string) string { return strings.ToUpper(name) })
		want := "ALPHA: 75%, BETA: 25%"
		if got != want {
			t.Fatalf("SummarizeShareUsage() = %q, want %q", got, want)
		}
	})

	t.Run("nil normalizeLabel defaults to strings.TrimSpace", func(t *testing.T) {
		got := SummarizeShareUsage(map[string]float64{
			"alpha": 100,
		}, 0, nil)
		want := "alpha: 100%"
		if got != want {
			t.Fatalf("SummarizeShareUsage() = %q, want %q", got, want)
		}
	})

	t.Run("empty or zero total returns empty string", func(t *testing.T) {
		if got := SummarizeShareUsage(nil, 5, nil); got != "" {
			t.Errorf("expected empty string for nil map, got %q", got)
		}
		if got := SummarizeShareUsage(map[string]float64{"a": 0, "b": -10}, 5, nil); got != "" {
			t.Errorf("expected empty string for zero/negative values, got %q", got)
		}
	})

	t.Run("alphabetical tie breaker on equal values", func(t *testing.T) {
		got := SummarizeShareUsage(map[string]float64{
			"zebra": 50,
			"apple": 50,
		}, 2, nil)
		want := "apple: 50%, zebra: 50%"
		if got != want {
			t.Fatalf("SummarizeShareUsage() = %q, want %q", got, want)
		}
	})
}

func TestSummarizeCountUsage(t *testing.T) {
	t.Run("basic sorting and normalization", func(t *testing.T) {
		got := SummarizeCountUsage(map[string]float64{
			"beta":  2,
			"alpha": 3,
		}, "req", 2, func(name string) string { return strings.ToUpper(name) })
		want := "ALPHA: 3 req, BETA: 2 req"
		if got != want {
			t.Fatalf("SummarizeCountUsage() = %q, want %q", got, want)
		}
	})

	t.Run("nil normalizeLabel defaults to strings.TrimSpace", func(t *testing.T) {
		got := SummarizeCountUsage(map[string]float64{
			"alpha": 5,
		}, "tokens", 0, nil)
		want := "alpha: 5 tokens"
		if got != want {
			t.Fatalf("SummarizeCountUsage() = %q, want %q", got, want)
		}
	})

	t.Run("empty or all zero returns empty string", func(t *testing.T) {
		if got := SummarizeCountUsage(nil, "req", 5, nil); got != "" {
			t.Errorf("expected empty string for nil map, got %q", got)
		}
		if got := SummarizeCountUsage(map[string]float64{"a": 0, "b": -5}, "req", 5, nil); got != "" {
			t.Errorf("expected empty string for zero/negative values, got %q", got)
		}
	})

	t.Run("maxItems truncates list and tie-breaks alphabetically", func(t *testing.T) {
		got := SummarizeCountUsage(map[string]float64{
			"c": 10,
			"b": 10,
			"a": 10,
			"d": 5,
		}, "units", 2, nil)
		want := "a: 10 units, b: 10 units"
		if got != want {
			t.Fatalf("SummarizeCountUsage() = %q, want %q", got, want)
		}
	})
}
