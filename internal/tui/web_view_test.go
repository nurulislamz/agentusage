package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

func TestWebProjectorDemoSnapshot(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	snap := core.NewUsageSnapshot("claude_code", "claude-code")
	snap.Timestamp = now
	snap.Status = core.StatusOK
	snap.Message = "~$42.18 today"
	used := 38.0
	limit := 100.0
	rem := 62.0
	snap.Metrics = map[string]core.Metric{
		"today_api_cost":  {Used: &used, Unit: "USD", Window: "today"},
		"usage_five_hour": {Used: &used, Limit: &limit, Remaining: &rem, Unit: "%", Window: "rolling-5h"},
	}

	p := WebProjector{
		TimeWindow:    core.TimeWindow3d,
		WarnThreshold: 0.25,
		CritThreshold: 0.1,
		UsageMode:     config.UsageModeRemaining,
		TileWidth:     72,
		DetailWidth:   80,
		Now:           now,
	}
	view := p.ProjectSnapshot(snap, "Claude Code")

	if view.Key != "claude_code:claude-code" {
		t.Fatalf("key = %q", view.Key)
	}
	if view.Status != string(core.StatusOK) {
		t.Fatalf("status = %q", view.Status)
	}
	if view.Summary == "" {
		t.Fatal("expected summary")
	}
	if len(view.TileLines) < 3 {
		t.Fatalf("expected tile lines, got %d", len(view.TileLines))
	}
	if view.GaugePercent < 0 {
		t.Fatalf("expected gauge percent, got %v", view.GaugePercent)
	}
	body := strings.Join(view.TileLines, "\n")
	if !strings.Contains(body, "claude-code") {
		t.Fatalf("tile should include account id:\n%s", body)
	}
}

func TestStripANSI(t *testing.T) {
	in := "\x1b[31mhello\x1b[0m world"
	if got := StripANSI(in); got != "hello world" {
		t.Fatalf("StripANSI = %q", got)
	}
}

func TestOrderSnapshotsRespectsDashboardOrder(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	snaps := map[string]core.UsageSnapshot{
		"cursor-ide": {
			ProviderID: "cursor",
			AccountID:  "cursor-ide",
			Timestamp:  now,
			Status:     core.StatusOK,
		},
		"claude-code": {
			ProviderID: "claude_code",
			AccountID:  "claude-code",
			Timestamp:  now,
			Status:     core.StatusOK,
		},
	}
	cfg := config.DefaultConfig()
	cfg.Dashboard.Providers = []config.DashboardProviderConfig{
		{AccountID: "claude-code", Enabled: true},
		{AccountID: "cursor-ide", Enabled: true},
	}
	cfg.Accounts = []core.AccountConfig{
		{ID: "claude-code", Provider: "claude_code"},
		{ID: "cursor-ide", Provider: "cursor"},
	}
	p := NewWebProjectorFromConfig(cfg)
	ordered := p.OrderSnapshots(snaps)
	if len(ordered) != 2 {
		t.Fatalf("ordered len = %d", len(ordered))
	}
	if ordered[0].AccountID != "claude-code" {
		t.Fatalf("first = %q, want claude-code", ordered[0].AccountID)
	}

	cfg.Dashboard.Providers[1].Enabled = false
	p = NewWebProjectorFromConfig(cfg)
	ordered = p.OrderSnapshots(snaps)
	if len(ordered) != 1 || ordered[0].AccountID != "claude-code" {
		t.Fatalf("disabled account should be excluded, got %+v", ordered)
	}
}

func TestWebProjectorOpenCodeDetailCards(t *testing.T) {
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

	p := WebProjector{
		TimeWindow:    core.TimeWindow3d,
		WarnThreshold: 0.25,
		CritThreshold: 0.1,
		UsageMode:     config.UsageModeRemaining,
		TileWidth:     72,
		DetailWidth:   80,
		Now:           now,
	}
	view := p.ProjectSnapshot(snap, "OpenCode")
	if !strings.Contains(view.StatusBadge, "MONTHLY") {
		t.Fatalf("status badge = %q, want MONTHLY LIMIT", view.StatusBadge)
	}
	if view.CycleSchedule == "" {
		t.Fatal("expected cycle schedule on compact header")
	}
	if view.LastRefreshed != "Last refreshed just now" {
		t.Fatalf("last_refreshed = %q, want Last refreshed just now", view.LastRefreshed)
	}

	stale := snap
	stale.Timestamp = now.Add(-10 * time.Minute)
	staleView := p.ProjectSnapshot(stale, "OpenCode")
	if staleView.LastRefreshed != "Last refreshed 10m0s ago" {
		t.Fatalf("stale last_refreshed = %q", staleView.LastRefreshed)
	}
	if !view.HasGauge {
		t.Fatal("expected has_gauge for remaining metrics including 0%")
	}

	usage := cardByTitle(view.DetailCards, "Usage")
	if usage.Title == "" {
		t.Fatalf("missing Usage card, got %#v", view.DetailCards)
	}
	gauges := rowsOfKind(usage, "gauge")
	if len(gauges) < 3 {
		t.Fatalf("want >=3 usage gauges, got %d in %#v", len(gauges), usage.Rows)
	}
	want := map[string]float64{
		"Five Hour Limit Remaining": 100,
		"Weekly Limit Remaining":    32,
		"Monthly Limit Remaining":   0,
	}
	for _, row := range gauges {
		pct, ok := want[row.Label]
		if !ok {
			continue
		}
		if row.Percent == nil || *row.Percent != pct {
			t.Errorf("gauge %q percent = %v, want %v", row.Label, row.Percent, pct)
		}
		delete(want, row.Label)
	}
	if len(want) > 0 {
		t.Fatalf("missing gauges: %v\nrows=%#v", want, gauges)
	}

	timers := cardByTitle(view.DetailCards, "Timers")
	if timers.Title == "" {
		t.Fatal("missing Timers card")
	}
	if len(rowsOfKind(timers, "timer")) < 3 {
		t.Fatalf("want timer rows, got %#v", timers.Rows)
	}
}

func TestWebProjectorGroupsOpenCodeAccounts(t *testing.T) {
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	snaps := []core.UsageSnapshot{
		{ProviderID: "opencode", AccountID: "opencode-mohammed", Timestamp: now, Status: core.StatusLimited},
		{ProviderID: "opencode", AccountID: "opencode-nurulz", Timestamp: now, Status: core.StatusOK},
		{ProviderID: "cursor", AccountID: "cursor-nurulz", Timestamp: now, Status: core.StatusOK},
	}
	p := WebProjector{Now: now, UsageMode: config.UsageModeRemaining}
	views := p.ProjectSnapshots(snaps, map[string]string{"opencode": "OpenCode", "cursor": "Cursor"})
	if len(views) != 3 {
		t.Fatalf("views = %d", len(views))
	}
	if views[0].ProviderID != "opencode" || views[1].ProviderID != "opencode" {
		t.Fatalf("expected OpenCode pair first, got %q %q", views[0].ProviderID, views[1].ProviderID)
	}
}

func TestWebProjectorAuthHasNoGauge(t *testing.T) {
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	snap := core.UsageSnapshot{
		ProviderID: "cursor",
		AccountID:  "cursor-auth",
		Status:     core.StatusAuth,
		Timestamp:  now,
		Message:    "Authentication required",
	}
	view := WebProjector{Now: now, UsageMode: config.UsageModeRemaining}.ProjectSnapshot(snap, "Cursor")
	if view.HasGauge {
		t.Fatal("AUTH row must not project a fake gauge")
	}
	if !strings.Contains(strings.ToLower(view.Summary), "auth") {
		t.Fatalf("summary = %q, want authentication copy", view.Summary)
	}
}

func TestWebUnmappedSummary(t *testing.T) {
	snaps := []core.UsageSnapshot{{
		ProviderID: "claude_code",
		AccountID:  "cc",
		Diagnostics: map[string]string{
			"telemetry_unmapped_providers": "anthropic,openai",
			"telemetry_unmapped_meta":      "anthropic=unconfigured:anthropic,openai=unconfigured",
		},
	}}
	count, phrase := WebUnmappedSummary(snaps)
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
	if !strings.Contains(phrase, "telemetry sources") {
		t.Fatalf("phrase = %q", phrase)
	}
}

func cardByTitle(cards []WebDetailCard, title string) WebDetailCard {
	for _, c := range cards {
		if c.Title == title {
			return c
		}
	}
	return WebDetailCard{}
}

func rowsOfKind(card WebDetailCard, kind string) []WebDetailRow {
	out := make([]WebDetailRow, 0)
	for _, r := range card.Rows {
		if r.Kind == kind {
			out = append(out, r)
		}
	}
	return out
}

func TestRowsFromSectionLinesSkipsChartArt(t *testing.T) {
	rows := rowsFromSectionLines([]string{
		"Daily Usage",
		"Cost ⠠⠔⠢⠊⠑ $5.23",
		"$7.4│ ⣀⡀",
		"│ ⣀⡠⠤⠒⠊⠉ ⠈⠉⠒⠒⠤⢄⣀",
		"$4.8└────────────────────────────────",
		"● Cost",
	}, false)
	got := make([]string, 0, len(rows))
	for _, r := range rows {
		got = append(got, r.Value)
	}
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "Daily Usage") {
		t.Fatalf("kept heading missing: %q", joined)
	}
	if strings.Contains(joined, "⣀") || strings.Contains(joined, "└") {
		t.Fatalf("chart art should be skipped: %q", joined)
	}
}

func TestThemeTokensForName(t *testing.T) {
	tokens := ThemeTokensForName("Tokyo Night")
	if tokens.Base == "" || tokens.Text == "" {
		t.Fatalf("incomplete theme tokens: %+v", tokens)
	}
	if tokens.Name == "" {
		t.Fatal("expected theme name")
	}
}
