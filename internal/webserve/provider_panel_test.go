package webserve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

func panelViewIDs(t *testing.T, srv *Server) map[string]bool {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("snapshots status = %d", w.Code)
	}
	var env Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	ids := make(map[string]bool, len(env.Views))
	for _, v := range env.Views {
		ids[v.AccountID] = true
	}
	return ids
}

func TestProvidersPanel_HideShowDelete(t *testing.T) {
	srv := testServer(t, Options{Demo: true})

	// 1. The dropdown lists every account with a Hide and a Delete action.
	w := getHTML(t, srv, "/partial/providers")
	if w.Code != http.StatusOK {
		t.Fatalf("partial/providers status = %d", w.Code)
	}
	panel := w.Body.String()
	for _, want := range []string{"cursor-ide", "codex-cli", `hx-post="actions/providers"`, `"op":"hide"`, `"op":"delete"`, "providers-group-title"} {
		if !strings.Contains(panel, want) {
			t.Errorf("providers panel missing %q", want)
		}
	}

	visible := panelViewIDs(t, srv)
	if !visible["cursor-ide"] {
		t.Fatal("demo cursor-ide should start visible")
	}

	// 2. Hiding an account drops it from the dashboard and flips the row to Show.
	hideResp := postForm(t, srv, "/actions/providers", "op=hide&account_id=cursor-ide")
	if hideResp.Code != http.StatusOK {
		t.Fatalf("hide status = %d", hideResp.Code)
	}
	body := hideResp.Body.String()
	if !strings.Contains(body, `id="providers-content" hx-swap-oob="innerHTML"`) {
		t.Error("hide response must refresh the dropdown out-of-band")
	}
	if !strings.Contains(body, `"op":"show"`) {
		t.Error("hidden account should offer a Show action")
	}
	if visible := panelViewIDs(t, srv); visible["cursor-ide"] {
		t.Error("hidden account must not render on the dashboard")
	}

	// 3. Showing it again restores the dashboard row.
	showResp := postForm(t, srv, "/actions/providers", "op=show&account_id=cursor-ide")
	if showResp.Code != http.StatusOK {
		t.Fatalf("show status = %d", showResp.Code)
	}
	if !strings.Contains(showResp.Body.String(), `"op":"hide"`) {
		t.Error("restored account should offer a Hide action")
	}
	if visible := panelViewIDs(t, srv); !visible["cursor-ide"] {
		t.Error("shown account must render on the dashboard again")
	}

	// 4. Deleting removes the account from config and keeps it hidden so
	// re-detection cannot resurface the box.
	delResp := postForm(t, srv, "/actions/providers", "op=delete&account_id=codex-cli")
	if delResp.Code != http.StatusOK {
		t.Fatalf("delete status = %d", delResp.Code)
	}
	cfg := *srv.collector.opts.Config
	for _, acct := range cfg.Accounts {
		if acct.ID == "codex-cli" {
			t.Error("deleted account still present in config accounts")
		}
	}
	entries := 0
	for _, p := range cfg.Dashboard.Providers {
		if p.AccountID == "codex-cli" {
			entries++
			if p.Enabled {
				t.Error("deleted account must stay hidden on the dashboard")
			}
		}
	}
	if entries != 1 {
		t.Errorf("dashboard provider entries for codex-cli = %d, want 1", entries)
	}
	if visible := panelViewIDs(t, srv); visible["codex-cli"] {
		t.Error("deleted account must not render on the dashboard")
	}

	// 5. Unknown ops are rejected.
	if bad := postForm(t, srv, "/actions/providers", "op=nuke&account_id=cursor-ide"); bad.Code != http.StatusBadRequest {
		t.Errorf("unknown op status = %d, want 400", bad.Code)
	}
	if missing := postForm(t, srv, "/actions/providers", "op=hide"); missing.Code != http.StatusBadRequest {
		t.Errorf("missing account_id status = %d, want 400", missing.Code)
	}
}

// Hide/show must RMW a single Enabled bit. Persisting the serve process's
// stale in-memory dashboard.providers list would wipe concurrent TUI edits.
func TestSetAccountVisibility_PreservesConcurrentProviderEdits(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("config dir is resolved from APPDATA on windows")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".config", "agentusage", "settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	hide := true
	onDisk := config.DefaultConfig()
	onDisk.Theme = "Nord"
	onDisk.Accounts = []core.AccountConfig{
		{ID: "cursor-ide", Provider: "cursor"},
		{ID: "work-openai", Provider: "openai"},
	}
	onDisk.Dashboard.Providers = []config.DashboardProviderConfig{
		{AccountID: "cursor-ide", Enabled: true},
		{AccountID: "work-openai", Enabled: false, HideCosts: &hide},
	}
	if err := config.SaveTo(path, onDisk); err != nil {
		t.Fatal(err)
	}

	// Serve still holds the pre-TUI snapshot: work-openai visible, no hide_costs.
	stale := config.DefaultConfig()
	stale.Accounts = append([]core.AccountConfig(nil), onDisk.Accounts...)
	stale.Dashboard.Providers = []config.DashboardProviderConfig{
		{AccountID: "cursor-ide", Enabled: true},
		{AccountID: "work-openai", Enabled: true},
	}
	srv := testServer(t, Options{
		Demo: false,
		Config: &stale,
		Collect: func() (Envelope, error) {
			return Envelope{Source: "test", Snapshots: []core.UsageSnapshot{
				{ProviderID: "cursor", AccountID: "cursor-ide", Status: core.StatusOK},
				{ProviderID: "openai", AccountID: "work-openai", Status: core.StatusOK},
			}}, nil
		},
	})

	hideResp := postForm(t, srv, "/actions/providers", "op=hide&account_id=cursor-ide")
	if hideResp.Code != http.StatusOK {
		t.Fatalf("hide status = %d body=%s", hideResp.Code, hideResp.Body.String())
	}

	loaded, err := config.LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Theme != "Nord" {
		t.Errorf("theme = %q, want Nord", loaded.Theme)
	}
	byID := map[string]config.DashboardProviderConfig{}
	for _, p := range loaded.Dashboard.Providers {
		byID[p.AccountID] = p
	}
	if byID["cursor-ide"].Enabled {
		t.Error("cursor-ide should be hidden on disk")
	}
	work := byID["work-openai"]
	if work.Enabled {
		t.Error("concurrent work-openai hide was wiped")
	}
	if work.HideCosts == nil || !*work.HideCosts {
		t.Error("concurrent work-openai hide_costs was wiped")
	}
	mem := srv.collector.opts.Config
	if mem == nil {
		t.Fatal("expected in-memory config after hide")
	}
	for _, p := range mem.Dashboard.Providers {
		if p.AccountID == "work-openai" {
			if p.Enabled {
				t.Error("in-memory config should refresh work-openai as hidden")
			}
			if p.HideCosts == nil || !*p.HideCosts {
				t.Error("in-memory config should refresh work-openai hide_costs")
			}
		}
		if p.AccountID == "cursor-ide" && p.Enabled {
			t.Error("in-memory cursor-ide should be hidden")
		}
	}
}
