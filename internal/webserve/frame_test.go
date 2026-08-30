package webserve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/tui"
)

func TestRenderTUIFrame_MatchesBrandAndAccount(t *testing.T) {
	ensureTrueColor()
	_ = tui.ThemeTokensForName("Deep Space")

	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	snaps := []core.UsageSnapshot{demoClaude(now), demoCursor(now)}
	cfg := config.DefaultConfig()
	cfg.Theme = "Deep Space"
	cfg.Dashboard.UsageMode = config.UsageModeRemaining

	frame := renderTUIFrame(cfg, snaps, 0, 120, 36)
	plain := tui.StripANSI(frame)
	if !strings.Contains(plain, "agentUsage") {
		t.Fatalf("missing brand:\n%s", plain)
	}
	if !strings.Contains(plain, "claude-code") {
		t.Fatalf("missing selected account:\n%s", plain)
	}
	if !strings.Contains(plain, "Usage") {
		t.Fatalf("missing detail usage section:\n%s", plain)
	}

	frame1 := renderTUIFrame(cfg, snaps, 1, 120, 36)
	plain1 := tui.StripANSI(frame1)
	if !strings.Contains(plain1, "cursor-ide") {
		t.Fatalf("cursor selection missing cursor-ide:\n%s", plain1)
	}
}

func TestSnapshotsIncludeDetailCards(t *testing.T) {
	srv := testServer(t, Options{Demo: true, Theme: "Deep Space"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var env Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if len(env.Views) == 0 {
		t.Fatal("expected views")
	}
	if len(env.Views[0].DetailCards) == 0 {
		t.Fatal("expected detail_cards")
	}
	if env.ProviderCount == 0 {
		t.Fatal("expected provider_count")
	}
	if env.TimeWindowLabel == "" {
		t.Fatal("expected time_window_label")
	}
}

func TestSnapshotsIncludeUnmappedChrome(t *testing.T) {
	srv := testServer(t, Options{
		Collect: func() (Envelope, error) {
			snap := core.NewUsageSnapshot("claude_code", "cc")
			snap.Status = core.StatusOK
			snap.Diagnostics = map[string]string{
				"telemetry_unmapped_providers": "anthropic",
				"telemetry_unmapped_meta":      "anthropic=unconfigured:anthropic",
			}
			return Envelope{Source: "direct", Snapshots: []core.UsageSnapshot{snap}}, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var env Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.UnmappedCount != 1 {
		t.Fatalf("unmapped_count = %d, want 1", env.UnmappedCount)
	}
	if env.UnmappedPhrase == "" {
		t.Fatal("expected unmapped_phrase")
	}
}

func TestResetHint_OpenCodeMonthlyLimitUsesMonthlyReset(t *testing.T) {
	now := time.Date(2026, 8, 30, 18, 15, 0, 0, time.UTC)
	zero, thirtyTwo, hundred := 0.0, 32.0, 100.0
	snap := core.UsageSnapshot{
		AccountID:  "opencode-mohammed",
		ProviderID: "opencode",
		Status:     core.StatusLimited,
		Timestamp:  now,
		Metrics: map[string]core.Metric{
			"rolling_usage":     {Remaining: &hundred},
			"weekly_usage":      {Remaining: &thirtyTwo},
			"monthly_usage_pct": {Remaining: &zero},
		},
		Resets: map[string]time.Time{
			"rolling_usage":     now.Add(4*time.Hour + 59*time.Minute),
			"weekly_usage":      now.Add(5*time.Hour + 45*time.Minute),
			"monthly_usage_pct": now.Add(9*24*time.Hour + 3*time.Hour),
		},
	}
	srv := testServer(t, Options{
		Now: func() time.Time { return now },
		Collect: func() (Envelope, error) {
			return Envelope{Source: "direct", Snapshots: []core.UsageSnapshot{snap}}, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var env Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if len(env.Views) != 1 {
		t.Fatalf("views = %d", len(env.Views))
	}
	hint := env.Views[0].ResetHint
	if !strings.Contains(hint, "9 days") {
		t.Fatalf("reset_hint = %q, want monthly Resets in 9 days", hint)
	}
	if strings.Contains(strings.ToLower(hint), "week") {
		t.Fatalf("reset_hint should not use weekly limit, got %q", hint)
	}
}
