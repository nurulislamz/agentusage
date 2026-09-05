package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestResolveRecentActivity(t *testing.T) {
	refTime := time.Date(2026, 9, 5, 14, 0, 0, 0, time.UTC)

	t.Run("active 12 minutes ago with percent attribute", func(t *testing.T) {
		snap := core.NewUsageSnapshot("claude_code", "claude-code")
		snap.Attributes["last_active_at"] = refTime.Add(-12 * time.Minute).Format(time.RFC3339)
		snap.Attributes["recent_activity_pct"] = "2.5"

		act := ResolveRecentActivity(snap, refTime)
		if !act.ActiveRecently {
			t.Errorf("expected ActiveRecently to be true for 12m ago")
		}
		if act.TimeAgo != "12m ago" {
			t.Errorf("expected TimeAgo '12m ago', got %q", act.TimeAgo)
		}
		if !act.HasPercent || act.Percent != 2.5 {
			t.Errorf("expected Percent 2.5 and HasPercent true, got %v (%f)", act.HasPercent, act.Percent)
		}
	})

	t.Run("active 75 minutes ago is not recently active", func(t *testing.T) {
		snap := core.NewUsageSnapshot("codex", "codex-main")
		snap.Attributes["last_active_at"] = refTime.Add(-75 * time.Minute).Format(time.RFC3339)
		snap.Attributes["recent_activity_pct"] = "10.0"

		act := ResolveRecentActivity(snap, refTime)
		if act.ActiveRecently {
			t.Errorf("expected ActiveRecently to be false for 75m ago (> 60m)")
		}
		if act.TimeAgo != "1h ago" {
			t.Errorf("expected TimeAgo '1h ago', got %q", act.TimeAgo)
		}
		if !act.HasPercent || act.Percent != 10.0 {
			t.Errorf("expected Percent 10.0, got %f", act.Percent)
		}
	})

	t.Run("active 30 seconds ago is just now", func(t *testing.T) {
		snap := core.NewUsageSnapshot("cursor", "cursor-pro")
		snap.Attributes["telemetry_last_event_at"] = refTime.Add(-30 * time.Second).Format(time.RFC3339)

		act := ResolveRecentActivity(snap, refTime)
		if !act.ActiveRecently {
			t.Errorf("expected ActiveRecently true for 30s ago")
		}
		if act.TimeAgo != "just now" {
			t.Errorf("expected TimeAgo 'just now', got %q", act.TimeAgo)
		}
		if act.HasPercent {
			t.Errorf("expected HasPercent false when no pct is provided")
		}
	})

	t.Run("percent from metrics when not in attributes", func(t *testing.T) {
		snap := core.NewUsageSnapshot("opencode", "opencode-agent")
		snap.Attributes["recent_activity_at"] = refTime.Add(-45 * time.Minute).Format(time.RFC3339)
		usedVal := 7.8
		snap.Metrics["recent_activity_pct"] = core.Metric{Used: &usedVal}

		act := ResolveRecentActivity(snap, refTime)
		if !act.ActiveRecently {
			t.Errorf("expected ActiveRecently true")
		}
		if !act.HasPercent || act.Percent != 7.8 {
			t.Errorf("expected Percent 7.8, got %v (%f)", act.HasPercent, act.Percent)
		}
	})

	t.Run("empty snapshot has no recent activity", func(t *testing.T) {
		snap := core.NewUsageSnapshot("ollama", "local-llama")
		act := ResolveRecentActivity(snap, refTime)
		if act.ActiveRecently {
			t.Errorf("expected ActiveRecently false for empty snapshot")
		}
		if act.HasPercent {
			t.Errorf("expected HasPercent false")
		}
	})
}

func TestRenderRecentActivityVisuals(t *testing.T) {
	dot := RenderActivityDot()
	if !strings.Contains(dot, "●") {
		t.Errorf("expected dot to contain '●', got %q", dot)
	}

	bar := RenderRecentActivityMicroBar(25.0, 8)
	if !strings.Contains(bar, "[") || !strings.Contains(bar, "]") {
		t.Errorf("expected micro-bar to contain brackets, got %q", bar)
	}
	if !strings.Contains(bar, "█") {
		t.Errorf("expected micro-bar to contain fill blocks, got %q", bar)
	}

	act := RecentActivity{
		ActiveRecently: true,
		TimeAgo:        "12m ago",
		Percent:        2.5,
		HasPercent:     true,
	}
	line := RenderRecentActivityLine(act, 8)
	if !strings.Contains(line, "Active 12m ago") {
		t.Errorf("expected line to contain 'Active 12m ago', got %q", line)
	}
	if !strings.Contains(line, "2.5%") {
		t.Errorf("expected line to contain '2.5%%', got %q", line)
	}
}

func TestCockpitRecentActivityCardAndStarts(t *testing.T) {
	now := time.Date(2026, 9, 5, 14, 0, 0, 0, time.UTC)
	snap := core.NewUsageSnapshot("claude_code", "claude-code")
	snap.Timestamp = now
	snap.Status = core.StatusOK
	snap.Attributes["last_active_at"] = now.Add(-10 * time.Minute).Format(time.RFC3339)
	snap.Attributes["recent_activity_pct"] = "4.5"

	out := RenderCockpit(snap, now, 80, 0.20, 0.05, core.TimeWindow30d, false, "remaining")
	if !strings.Contains(out, "● RECENT ACTIVITY") {
		t.Fatalf("expected Cockpit to contain '● RECENT ACTIVITY', got:\n%s", out)
	}
	if !strings.Contains(out, "Active 10m ago") {
		t.Fatalf("expected Cockpit to contain 'Active 10m ago', got:\n%s", out)
	}

	starts := CockpitSectionStarts(snap, now, 80, 0.20, 0.05, core.TimeWindow30d, false, "remaining")
	if len(starts) < 2 {
		t.Fatalf("expected at least 2 section starts, got %v", starts)
	}
	// First section is Recent Activity at line 4
	if starts[0] != 4 {
		t.Errorf("expected starts[0] to be 4, got %d", starts[0])
	}
	// Second section is Usage & Quotas at line 4 + 3 + 2 = 9
	if starts[1] != 9 {
		t.Errorf("expected starts[1] to be 9, got %d", starts[1])
	}
}
