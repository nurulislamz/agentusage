package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

func TestRenderCockpit_HeroAndSubhero(t *testing.T) {
	now := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	snap := core.UsageSnapshot{
		ProviderID: "openrouter",
		AccountID:  "openrouter-prod",
		Status:     core.StatusOK,
		Timestamp:  now.Add(-2 * time.Minute),
		Metrics: map[string]core.Metric{
			"credits_remaining": {
				Remaining: core.Float64Ptr(42.50),
			},
		},
	}

	out := RenderCockpit(snap, now, 80, 0.20, 0.05, core.TimeWindow30d, false, config.UsageModeRemaining)

	if !strings.Contains(out, "openrouter-prod") {
		t.Fatalf("expected account ID in hero, got:\n%s", out)
	}
	if !strings.Contains(out, "⚡ USAGE & QUOTAS") {
		t.Fatalf("expected USAGE & QUOTAS card header, got:\n%s", out)
	}
	if !strings.Contains(out, "⏱ TIMERS & SCHEDULE") {
		t.Fatalf("expected TIMERS & SCHEDULE card header, got:\n%s", out)
	}
}

func TestRenderCockpit_QuotasAndTimers(t *testing.T) {
	now := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	snap := cursorPlanSnap(now)

	out := RenderCockpit(snap, now, 80, 0.20, 0.05, core.TimeWindow30d, false, config.UsageModeRemaining)

	if !strings.Contains(out, "93.00%") {
		t.Fatalf("expected 93.00%% in usage & quotas, got:\n%s", out)
	}
	if !strings.Contains(out, "⏱ TIMERS & SCHEDULE") {
		t.Fatalf("expected timers section, got:\n%s", out)
	}
}

func TestRenderCockpit_ActivityAndTrend(t *testing.T) {
	now := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	snap := core.UsageSnapshot{
		ProviderID: "openai",
		AccountID:  "openai-main",
		Status:     core.StatusOK,
		Timestamp:  now,
		DailySeries: map[string][]core.TimePoint{
			"cost": {
				{Date: "2026-09-02", Value: 1.20},
				{Date: "2026-09-03", Value: 2.50},
				{Date: "2026-09-04", Value: 3.10},
			},
		},
	}

	out := RenderCockpit(snap, now, 80, 0.20, 0.05, core.TimeWindow30d, false, config.UsageModeRemaining)

	if !strings.Contains(out, "📈 ACTIVITY & TREND") {
		t.Fatalf("expected ACTIVITY & TREND card header, got:\n%s", out)
	}
	if !strings.Contains(out, "Today:") {
		t.Fatalf("expected today stat in activity card, got:\n%s", out)
	}
}

func TestRenderCockpit_NarrowTerminal(t *testing.T) {
	now := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	snap := core.UsageSnapshot{
		ProviderID: "cursor",
		AccountID:  "cursor-personal",
		Status:     core.StatusOK,
		Timestamp:  now,
	}

	out := RenderCockpit(snap, now, 25, 0.20, 0.05, core.TimeWindow30d, false, config.UsageModeRemaining)
	if out == "" {
		t.Fatal("expected non-empty output on narrow terminal")
	}
}
