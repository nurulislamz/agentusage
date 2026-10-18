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

func TestProvidersPanel_DragReorderAndMainBoardCleanliness(t *testing.T) {
	srv := testServer(t, Options{Demo: true})

	// 1. Main board views (matrix, bento, bars) must NOT have provider-drag-handle
	for _, layout := range []string{"matrix", "bento", "bars", "dials", "strips"} {
		w := getHTML(t, srv, "/partial/app?layout="+layout)
		if w.Code != http.StatusOK {
			t.Fatalf("GET /partial/app?layout=%s status = %d", layout, w.Code)
		}
		body := w.Body.String()
		if strings.Contains(body, "provider-drag-handle") {
			t.Errorf("layout %s should not contain provider-drag-handle on main board", layout)
		}
		if strings.Contains(body, `class="provider-group-header" draggable="true"`) {
			t.Errorf("layout %s should not have draggable=true on provider-group-header", layout)
		}
	}

	// 2. Providers submenu (/partial/providers) MUST have provider-drag-handle and data-provider
	wProviders := getHTML(t, srv, "/partial/providers")
	if wProviders.Code != http.StatusOK {
		t.Fatalf("GET /partial/providers status = %d", wProviders.Code)
	}
	provBody := wProviders.Body.String()
	if !strings.Contains(provBody, "provider-drag-handle") {
		t.Error("providers panel should contain provider-drag-handle")
	}
	if !strings.Contains(provBody, `data-provider="`) {
		t.Error("providers panel should contain data-provider attribute on groups")
	}
	if !strings.Contains(provBody, `class="providers-group-title" draggable="true"`) {
		t.Error("providers panel group titles should be draggable")
	}
	if !strings.Contains(provBody, "provider-move-btn") {
		t.Error("providers panel should contain provider-move-btn switch buttons")
	}

	// 3. Save a specific provider order and check that /partial/providers renders in that order
	postResp := postForm(t, srv, "/actions/provider-order", "order=codex|cursor|antigravity&silent=1")
	if postResp.Code != http.StatusOK {
		t.Fatalf("provider-order status = %d", postResp.Code)
	}

	// Request /partial/providers with the cookie from the response
	req := httptest.NewRequest(http.MethodGet, "/partial/providers", nil)
	for _, c := range postResp.Result().Cookies() {
		if c.Name == cookieProviderOrder {
			req.AddCookie(c)
		}
	}
	wOrdered := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wOrdered, req)
	if wOrdered.Code != http.StatusOK {
		t.Fatalf("GET /partial/providers status = %d", wOrdered.Code)
	}
	orderedBody := wOrdered.Body.String()

	idxCodex := strings.Index(orderedBody, `data-provider="codex"`)
	idxCursor := strings.Index(orderedBody, `data-provider="cursor"`)
	if idxCodex == -1 || idxCursor == -1 {
		t.Fatalf("expected codex and cursor in provider panel, got body: %s", orderedBody)
	}
	if idxCodex > idxCursor {
		t.Errorf("expected codex (%d) to appear before cursor (%d) after ordering", idxCodex, idxCursor)
	}
}

