package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestRenderMatrixView_Basic(t *testing.T) {
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
	snap2 := core.UsageSnapshot{
		ProviderID: "openai",
		AccountID:  "openai-main",
		Status:     core.StatusOK,
		Timestamp:  now,
		Metrics: map[string]core.Metric{
			"credits": {Remaining: core.Float64Ptr(85.0)},
		},
	}

	m := Model{
		dashboardView: dashboardViewMatrix,
		sortedIDs:     []string{"claude-work", "openai-main"},
		snapshots: map[string]core.UsageSnapshot{
			"claude-work": snap1,
			"openai-main": snap2,
		},
		cursor:        0,
		referenceTime: now,
	}

	out := m.renderMatrixView(100, 24)

	if !strings.Contains(out, "claude-work") {
		t.Fatalf("expected claude-work in matrix view, got:\n%s", out)
	}
	if !strings.Contains(out, "openai-main") {
		t.Fatalf("expected openai-main in matrix view, got:\n%s", out)
	}
	if !strings.Contains(out, "ACCOUNT") {
		t.Fatalf("expected table header in matrix view, got:\n%s", out)
	}
	if !strings.Contains(out, "❯") {
		t.Fatalf("expected cursor indicator on selected row, got:\n%s", out)
	}
}

func TestRenderMatrixView_Empty(t *testing.T) {
	m := Model{
		dashboardView: dashboardViewMatrix,
		sortedIDs:     nil,
		snapshots:     nil,
	}

	out := m.renderMatrixView(80, 10)
	if !strings.Contains(out, "No provider accounts found") {
		t.Fatalf("expected empty state message, got:\n%s", out)
	}
}
