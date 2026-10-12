package webserve

import (
	"encoding/json"
	"testing"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/tui"
)

func TestSnapshotsForCacheKeepsDisabledAccounts(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Accounts = []core.AccountConfig{
		{ID: "hide-me", Provider: "openai"},
		{ID: "keep-me", Provider: "openai"},
	}
	cfg.Dashboard.Providers = []config.DashboardProviderConfig{
		{AccountID: "hide-me", Enabled: false},
		{AccountID: "keep-me", Enabled: true},
	}
	snaps := map[string]core.UsageSnapshot{
		"hide-me": core.NewUsageSnapshot("openai", "hide-me"),
		"keep-me": core.NewUsageSnapshot("openai", "keep-me"),
	}

	filtered := tui.NewWebProjectorFromConfig(cfg).OrderSnapshots(snaps)
	if len(filtered) != 1 || filtered[0].AccountID != "keep-me" {
		t.Fatalf("dashboard projector should omit hidden accounts, got %+v", filtered)
	}

	cached := snapshotsForCache(cfg, snaps)
	ids := make(map[string]bool, len(cached))
	for _, snap := range cached {
		ids[snap.AccountID] = true
	}
	if !ids["hide-me"] || !ids["keep-me"] {
		t.Fatalf("cache snapshots must keep hidden accounts, got %+v", cached)
	}
	if len(cfg.Dashboard.Providers) != 2 {
		t.Fatalf("snapshotsForCache must not mutate caller's providers, got %d", len(cfg.Dashboard.Providers))
	}
}

func TestSetAccountVisibility_ShowAfterCachedHide(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Accounts = []core.AccountConfig{
		{ID: "hide-me", Provider: "openai"},
		{ID: "keep-me", Provider: "openai"},
	}
	cfg.Dashboard.Providers = []config.DashboardProviderConfig{
		{AccountID: "hide-me", Enabled: true},
		{AccountID: "keep-me", Enabled: true},
	}
	all := []core.UsageSnapshot{
		core.NewUsageSnapshot("openai", "hide-me"),
		core.NewUsageSnapshot("openai", "keep-me"),
	}
	for i := range all {
		all[i].Status = core.StatusOK
	}

	srv := testServer(t, Options{
		Demo:   true,
		Config: &cfg,
		Collect: func() (Envelope, error) {
			return Envelope{Source: "test", Snapshots: all}, nil
		},
	})

	if visible := panelViewIDs(t, srv); !visible["hide-me"] {
		t.Fatal("hide-me should start visible")
	}

	hideResp := postForm(t, srv, "/actions/providers", "op=hide&account_id=hide-me")
	if hideResp.Code != 200 {
		t.Fatalf("hide status = %d", hideResp.Code)
	}
	if visible := panelViewIDs(t, srv); visible["hide-me"] {
		t.Fatal("hidden account must leave dashboard views")
	}

	reqSnaps := getHTML(t, srv, "/api/v1/snapshots")
	var env Envelope
	if err := json.NewDecoder(reqSnaps.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, snap := range env.Snapshots {
		if snap.AccountID == "hide-me" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("hidden account must remain in cached snapshots so Show can restore it")
	}

	showResp := postForm(t, srv, "/actions/providers", "op=show&account_id=hide-me")
	if showResp.Code != 200 {
		t.Fatalf("show status = %d", showResp.Code)
	}
	if visible := panelViewIDs(t, srv); !visible["hide-me"] {
		t.Fatal("shown account must return to dashboard views from cached snapshots")
	}
}
