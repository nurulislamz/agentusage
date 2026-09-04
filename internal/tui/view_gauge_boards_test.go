package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func testGaugeBoardModel() Model {
	now := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	snap1 := core.UsageSnapshot{
		ProviderID: "anthropic",
		AccountID:  "claude-work",
		Status:     core.StatusOK,
		Timestamp:  now,
		Metrics: map[string]core.Metric{
			"session": {
				Remaining: core.Float64Ptr(58.0),
				Limit:     core.Float64Ptr(100.0),
				Unit:      "%",
			},
		},
	}
	snap2 := core.UsageSnapshot{
		ProviderID: "openai",
		AccountID:  "openai-main",
		Status:     core.StatusOK,
		Timestamp:  now,
		Metrics: map[string]core.Metric{
			"credits": {
				Remaining: core.Float64Ptr(85.0),
				Limit:     core.Float64Ptr(100.0),
				Unit:      "%",
			},
		},
	}

	return Model{
		sortedIDs: []string{"claude-work", "openai-main"},
		snapshots: map[string]core.UsageSnapshot{
			"claude-work": snap1,
			"openai-main": snap2,
		},
		cursor:        0,
		referenceTime: now,
	}
}

func TestRenderBarsView_Basic(t *testing.T) {
	m := testGaugeBoardModel()
	m.dashboardView = dashboardViewBars

	out := m.renderBarsView(90, 24)

	if !strings.Contains(out, "claude-work") {
		t.Fatalf("expected claude-work in bars view, got:\n%s", out)
	}
	if !strings.Contains(out, "ANTHROPIC") {
		t.Fatalf("expected ANTHROPIC in bars view, got:\n%s", out)
	}
}

func TestRenderDialsView_Basic(t *testing.T) {
	m := testGaugeBoardModel()
	m.dashboardView = dashboardViewDials

	out := m.renderDialsView(90, 24)

	if !strings.Contains(out, "claude-work") {
		t.Fatalf("expected claude-work in dials view, got:\n%s", out)
	}
	if !strings.Contains(out, "╭─") {
		t.Fatalf("expected radial dial arc in dials view, got:\n%s", out)
	}
}

func TestRenderStripsView_Basic(t *testing.T) {
	m := testGaugeBoardModel()
	m.dashboardView = dashboardViewStrips

	out := m.renderStripsView(90, 24)

	if !strings.Contains(out, "claude-work") {
		t.Fatalf("expected claude-work in strips view, got:\n%s", out)
	}
	if !strings.Contains(out, "█") {
		t.Fatalf("expected bar glyph in strips view, got:\n%s", out)
	}
}

func TestRenderGaugeBoards_Empty(t *testing.T) {
	m := Model{
		sortedIDs: nil,
		snapshots: nil,
	}

	if out := m.renderBarsView(80, 10); !strings.Contains(out, "No provider accounts found") {
		t.Fatalf("expected empty state in bars, got:\n%s", out)
	}
	if out := m.renderDialsView(80, 10); !strings.Contains(out, "No provider accounts found") {
		t.Fatalf("expected empty state in dials, got:\n%s", out)
	}
	if out := m.renderStripsView(80, 10); !strings.Contains(out, "No provider accounts found") {
		t.Fatalf("expected empty state in strips, got:\n%s", out)
	}
}
