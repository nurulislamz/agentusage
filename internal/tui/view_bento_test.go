package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestRenderBentoView_Basic(t *testing.T) {
	now := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	snap1 := core.UsageSnapshot{
		ProviderID: "anthropic",
		AccountID:  "claude-work",
		Status:     core.StatusOK,
		Timestamp:  now,
		Metrics: map[string]core.Metric{
			"session": {Remaining: core.Float64Ptr(58.0)},
		},
	}

	m := Model{
		dashboardView: dashboardViewBento,
		sortedIDs:     []string{"claude-work"},
		snapshots: map[string]core.UsageSnapshot{
			"claude-work": snap1,
		},
		cursor:        0,
		referenceTime: now,
	}

	out := m.renderBentoView(100, 24)

	if !strings.Contains(out, "claude-work") {
		t.Fatalf("expected claude-work in bento view, got:\n%s", out)
	}
	if !strings.Contains(out, "ANTHROPIC") {
		t.Fatalf("expected ANTHROPIC group header in bento view, got:\n%s", out)
	}
}

func TestRenderBentoView_Empty(t *testing.T) {
	m := Model{
		dashboardView: dashboardViewBento,
		sortedIDs:     nil,
		snapshots:     nil,
	}

	out := m.renderBentoView(80, 10)
	if !strings.Contains(out, "No provider accounts found") {
		t.Fatalf("expected empty state message, got:\n%s", out)
	}
}
