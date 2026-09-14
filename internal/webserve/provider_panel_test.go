package webserve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
