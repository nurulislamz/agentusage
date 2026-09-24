package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestBuildAntigravityDetailUsageSection_EmptyUnknownSnapshot_ReturnsNil(t *testing.T) {
	snap := core.UsageSnapshot{
		ProviderID: "antigravity",
		Status:     core.StatusUnknown,
		Metrics:    map[string]core.Metric{},
	}
	lines := buildAntigravityDetailUsageSection(snap, 80, 0.25, 0.1, time.Now(), false)
	if len(lines) != 0 {
		t.Fatalf("expected nil or empty lines for empty unknown snapshot, got %v", lines)
	}
}

func TestBuildAntigravityDetailUsageSection_ExpiredReset_DecaysTo100(t *testing.T) {
	staleRem := 88.0
	now := time.Now()
	snap := core.UsageSnapshot{
		ProviderID: "antigravity",
		Status:     core.StatusOK,
		Metrics: map[string]core.Metric{
			"quota_gemini_5h": {Remaining: &staleRem, Limit: core.Float64Ptr(100), Unit: "%", Window: "5h"},
		},
		Resets: map[string]time.Time{
			"quota_gemini_5h": now.Add(-2 * time.Hour), // expired 2 hours ago
		},
	}
	lines := buildAntigravityDetailUsageSection(snap, 80, 0.25, 0.1, now, false)
	content := strings.Join(lines, "\n")
	plain := StripANSI(content)
	if strings.Contains(plain, "88") {
		t.Fatalf("expected 88%% to decay to 100%% remaining, but found '88' in:\n%s", plain)
	}
	if !strings.Contains(plain, "100.00%") {
		t.Fatalf("expected '100.00%%' in output, got:\n%s", plain)
	}
}
