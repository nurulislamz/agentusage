package main

import (
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestFindAccount(t *testing.T) {
	accounts := []core.AccountConfig{
		{ID: "antigravity", Provider: "antigravity"},
		{ID: "antigravity-nurulz", Provider: "antigravity"},
		{ID: "cursor-physics", Provider: "cursor"},
		{ID: "opencode-mohammed", Provider: "opencode"},
	}
	accounts[1].SetHint("box_name", "nurulz")
	accounts[2].SetHint("box_name", "physics")

	// 1. Exact ID
	acct, ok := findAccount(accounts, "antigravity-nurulz")
	if !ok || acct.ID != "antigravity-nurulz" {
		t.Errorf("expected to find antigravity-nurulz, got %+v, ok=%v", acct, ok)
	}

	// 2. Case-insensitive exact ID
	acct, ok = findAccount(accounts, "ANTIGRAVITY-NURULZ")
	if !ok || acct.ID != "antigravity-nurulz" {
		t.Errorf("expected case-insensitive match, got %+v, ok=%v", acct, ok)
	}

	// 3. Antigravity box name match (nurulz -> antigravity-nurulz)
	acct, ok = findAccount(accounts, "nurulz")
	if !ok || acct.ID != "antigravity-nurulz" {
		t.Errorf("expected box name nurulz to resolve to antigravity-nurulz, got %+v, ok=%v", acct, ok)
	}

	// 4. Cursor box name match (physics -> cursor-physics)
	acct, ok = findAccount(accounts, "physics")
	if !ok || acct.ID != "cursor-physics" {
		t.Errorf("expected physics to resolve to cursor-physics, got %+v, ok=%v", acct, ok)
	}

	// 5. Unknown account
	_, ok = findAccount(accounts, "nonexistent-box")
	if ok {
		t.Error("expected nonexistent-box to return false")
	}
}

func TestBuildGetResponse_FiveHourDefault(t *testing.T) {
	acct := core.AccountConfig{
		ID:       "antigravity-nurulz",
		Provider: "antigravity",
	}

	limit := 100.0
	geminiRemaining := 85.0
	geminiUsed := 15.0
	claudeRemaining := 95.0
	claudeUsed := 5.0

	resetTime := time.Now().UTC().Add(2 * time.Hour)

	snap := core.UsageSnapshot{
		ProviderID: "antigravity",
		AccountID:  "antigravity-nurulz",
		Status:     core.StatusOK,
		Metrics: map[string]core.Metric{
			"quota_gemini_5h": {
				Limit:     &limit,
				Used:      &geminiUsed,
				Remaining: &geminiRemaining,
				Unit:      "%",
				Window:    "5h",
			},
			"quota_claude_5h": {
				Limit:     &limit,
				Used:      &claudeUsed,
				Remaining: &claudeRemaining,
				Unit:      "%",
				Window:    "5h",
			},
		},
		Resets: map[string]time.Time{
			"quota_gemini_5h": resetTime,
		},
	}

	resp := buildGetResponse(acct, snap, "5h")

	if resp.ID != "antigravity-nurulz" {
		t.Errorf("expected ID antigravity-nurulz, got %s", resp.ID)
	}
	if resp.Window != "5h" {
		t.Errorf("expected Window 5h, got %s", resp.Window)
	}
	if resp.Remaining == nil || *resp.Remaining != 85.0 {
		t.Errorf("expected bottleneck remaining 85.0, got %v", resp.Remaining)
	}
	if len(resp.Pools) != 2 {
		t.Fatalf("expected 2 pools, got %d", len(resp.Pools))
	}
	if p, ok := resp.Pools["gemini_5h"]; !ok || *p.Remaining != 85.0 {
		t.Errorf("expected gemini_5h pool with 85.0 remaining, got %+v", p)
	}
	if p, ok := resp.Pools["claude_5h"]; !ok || *p.Remaining != 95.0 {
		t.Errorf("expected claude_5h pool with 95.0 remaining, got %+v", p)
	}
	if resp.ResetsIn == "" || !strings.Contains(resp.ResetsIn, "h") {
		t.Errorf("expected formatted ResetsIn countdown, got %q", resp.ResetsIn)
	}
}

func TestBuildGetResponse_WeeklyWindow(t *testing.T) {
	acct := core.AccountConfig{
		ID:       "antigravity-nurulz",
		Provider: "antigravity",
	}

	limit := 100.0
	geminiWeeklyRem := 60.0
	geminiWeeklyUsed := 40.0

	snap := core.UsageSnapshot{
		ProviderID: "antigravity",
		AccountID:  "antigravity-nurulz",
		Status:     core.StatusOK,
		Metrics: map[string]core.Metric{
			"quota_gemini_5h": {
				Limit:  &limit,
				Window: "5h",
			},
			"quota_gemini_weekly": {
				Limit:     &limit,
				Used:      &geminiWeeklyUsed,
				Remaining: &geminiWeeklyRem,
				Unit:      "%",
				Window:    "7d",
			},
		},
	}

	resp := buildGetResponse(acct, snap, "weekly")

	if resp.Window != "weekly" {
		t.Errorf("expected Window weekly, got %s", resp.Window)
	}
	if resp.Remaining == nil || *resp.Remaining != 60.0 {
		t.Errorf("expected remaining 60.0, got %v", resp.Remaining)
	}
	if _, ok := resp.Pools["gemini_weekly"]; !ok {
		t.Errorf("expected gemini_weekly pool in response: %+v", resp.Pools)
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{0, "now"},
		{-5 * time.Minute, "now"},
		{30 * time.Second, "30s"},
		{2*time.Minute + 15*time.Second, "2m 15s"},
		{3*time.Hour + 20*time.Minute, "3h 20m"},
		{26 * time.Hour, "1d 2h"},
	}

	for _, c := range cases {
		got := formatDuration(c.d)
		if got != c.want {
			t.Errorf("formatDuration(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}
