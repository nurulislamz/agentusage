package webserve

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func getHTML(t *testing.T, srv *Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	return w
}

func postForm(t *testing.T, srv *Server, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	return w
}

func envelopeTheme(t *testing.T, srv *Server) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/meta", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("meta status = %d", w.Code)
	}
	var body struct {
		Theme string `json:"theme"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body.Theme
}

func envelopeUsageMode(t *testing.T, srv *Server) string {
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
	return env.UsageMode
}

func TestShellServesServerRenderedChrome(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	w := getHTML(t, srv, "/")
	if w.Code != http.StatusOK {
		t.Fatalf("GET / status = %d", w.Code)
	}
	html := w.Body.String()
	for _, want := range []string{
		"agentUsage",
		`id="empty-error"`,
		`id="filter-input"`,
		`id="token-modal"`,
		`id="app" class="shell layout-split" data-layout="split" hidden hx-get="partial/app" hx-trigger="load`,
		`src="htmx.min.js"`,
		`src="app.js"`,
		`href="app.css"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("shell missing %q", want)
		}
	}
	if strings.Contains(html, "tui-frame") {
		t.Error("shell should not paint a TUI frame")
	}
	if strings.Contains(html, "agentusage serve --demo") {
		t.Error("shell should not advertise --demo in the empty state")
	}
	if strings.Contains(html, `src="/app.js"`) || strings.Contains(html, `href="/app.css"`) {
		t.Error("shell must use relative asset URLs for base-path support")
	}
}

func TestAppFragmentRendersDemoViews(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	w := getHTML(t, srv, "/partial/app")
	if w.Code != http.StatusOK {
		t.Fatalf("GET /partial/app status = %d", w.Code)
	}
	html := w.Body.String()
	snapReq := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	snapW := httptest.NewRecorder()
	srv.Handler().ServeHTTP(snapW, snapReq)
	var env Envelope
	if err := json.NewDecoder(snapW.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	wantCount := fmt.Sprintf("%d agents · Split", len(env.Views))

	for _, want := range []string{
		`id="app" class="shell layout-split`,
		`id="nav" class="nav"`,
		`id="panel"`,
		`class="logo"`,
		wantCount,
		`class="item nav-item`,
		`class="hero"`,
		`id="footer-btn-mode"`,
		`id="footer-btn-layout"`,
		`id="footer-theme-select"`,
		`id="key-next"`,
		`hx-post="actions/usage-mode"`,
		"codex-cli",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("app fragment missing %q", want)
		}
	}

	cockpit := getHTML(t, srv, "/partial/app?account=opencode-pro").Body.String()
	for _, want := range []string{"Usage &amp; quotas", `class="lin-track"`, "background:var(--accent)"} {
		if !strings.Contains(cockpit, want) {
			t.Errorf("opencode cockpit missing %q", want)
		}
	}
	if strings.Contains(cockpit, "ZgotmplZ") {
		t.Error("inline CSS values must not be sanitized into ZgotmplZ")
	}
}

func TestAppFragmentLayouts(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	cases := []struct {
		layout  string
		markers []string
	}{
		{"split", []string{`class="hero"`, `class="item nav-item`}},
		{"matrix", []string{`class="matrix-table"`, `class="matrix-row`, `class="provider-group-box"`}},
		{"bento", []string{`class="bento-tile`, `class="bento-quota-bar"`}},
		{"bars", []string{`class="board-grid board-bars"`, `class="lin-track"`}},
		{"dials", []string{`class="board-grid board-dials"`, `class="dial-svg"`}},
		{"strips", []string{`class="board-grid board-strips"`, `class="strip-track"`}},
	}
	for _, tc := range cases {
		t.Run(tc.layout, func(t *testing.T) {
			w := getHTML(t, srv, "/partial/app?layout="+tc.layout)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d", w.Code)
			}
			html := w.Body.String()
			for _, marker := range tc.markers {
				if !strings.Contains(html, marker) {
					t.Errorf("layout %s missing %q", tc.layout, marker)
				}
			}
			if !strings.Contains(html, `data-layout="`+tc.layout+`"`) {
				t.Errorf("layout %s missing data-layout marker", tc.layout)
			}
			if tc.layout != "split" {
				if !strings.Contains(html, `aria-label="Account roster" hx-indicator="#fetching-detail" hidden`) {
					t.Errorf("layout %s should hide the account roster", tc.layout)
				}
			}
			found := false
			for _, c := range w.Result().Cookies() {
				if c.Name == cookieLayout && c.Value == tc.layout {
					found = true
				}
			}
			if !found {
				t.Errorf("layout %s should persist in the %s cookie", tc.layout, cookieLayout)
			}
		})
	}
}

func TestAppFragmentFilterAndSelection(t *testing.T) {
	srv := testServer(t, Options{Demo: true})

	filtered := getHTML(t, srv, "/partial/app?q=cursor")
	if filtered.Code != http.StatusOK {
		t.Fatalf("filtered status = %d", filtered.Code)
	}
	body := filtered.Body.String()
	if !strings.Contains(body, "cursor-ide") {
		t.Error("filtered fragment should keep cursor-ide")
	}
	if strings.Contains(body, "codex-cli") {
		t.Error("filtered fragment should drop codex-cli")
	}
	if !strings.Contains(body, "1 agents (filtered)") {
		t.Error("filtered header should report the filtered count")
	}

	empty := getHTML(t, srv, "/partial/app?q=zzz-nothing").Body.String()
	if !strings.Contains(empty, "No matches.") {
		t.Error("empty filter result should render the no-matches state")
	}

	selected := getHTML(t, srv, "/partial/app?account=opencode-pro").Body.String()
	if !strings.Contains(selected, "nav-item selected") || !strings.Contains(selected, "opencode-pro") {
		t.Error("explicit account should be selected in the roster")
	}
	if !strings.Contains(selected, "mobile-view-detail") {
		t.Error("selecting an account should switch the mobile split state to detail")
	}
}

func TestAppFragmentRequiresAuthButShellDoesNot(t *testing.T) {
	srv := testServer(t, Options{Demo: true, AuthToken: "s3cret"})

	if w := getHTML(t, srv, "/"); w.Code != http.StatusOK {
		t.Fatalf("shell status = %d, want 200", w.Code)
	}
	req := httptest.NewRequest(http.MethodGet, "/partial/app", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated fragment status = %d, want 401", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/partial/app", nil)
	req.Header.Set("Authorization", "Bearer s3cret")
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("authenticated fragment status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "codex-cli") {
		t.Error("authenticated fragment should contain demo accounts")
	}
}

func TestThemeActionCyclesThemeInFragment(t *testing.T) {
	srv := testServer(t, Options{Demo: true, Theme: "Gruvbox"})
	before := envelopeTheme(t, srv)

	w := postForm(t, srv, "/actions/theme", "")
	if w.Code != http.StatusOK {
		t.Fatalf("theme action status = %d body %s", w.Code, w.Body.String())
	}
	after := envelopeTheme(t, srv)
	if after == before {
		t.Fatalf("theme did not change (still %q)", after)
	}
	body := w.Body.String()
	if !strings.Contains(body, `id="theme-vars"`) {
		t.Error("theme action should return the fragment with live theme vars")
	}
	if !strings.Contains(body, `value="`+after+`" selected`) {
		t.Errorf("theme selector should select %q", after)
	}

	back := postForm(t, srv, "/actions/theme", "direction=backward")
	if back.Code != http.StatusOK {
		t.Fatalf("backward theme status = %d", back.Code)
	}
	if got := envelopeTheme(t, srv); got != before {
		t.Errorf("backward cycle = %q, want %q", got, before)
	}
}

func TestUsageModeActionTogglesFragment(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	if mode := envelopeUsageMode(t, srv); mode != "remaining" {
		t.Fatalf("initial usage mode = %q", mode)
	}
	w := postForm(t, srv, "/actions/usage-mode", "")
	if w.Code != http.StatusOK {
		t.Fatalf("usage mode action status = %d", w.Code)
	}
	if mode := envelopeUsageMode(t, srv); mode != "used" {
		t.Fatalf("usage mode after toggle = %q, want used", mode)
	}
	if !strings.Contains(w.Body.String(), ">Used<") {
		t.Error("footer should show the used-mode label")
	}
}

func TestLayoutActionStoresLayoutAndRendersIt(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	w := postForm(t, srv, "/actions/layout", "layout=matrix")
	if w.Code != http.StatusOK {
		t.Fatalf("layout action status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `data-layout="matrix"`) || !strings.Contains(body, `class="matrix-table"`) {
		t.Error("layout action should render the requested layout")
	}
	if !strings.Contains(body, "Layout: Matrix") {
		t.Error("layout action should toast the new layout")
	}
	var cookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == cookieLayout {
			cookie = c
		}
	}
	if cookie == nil || cookie.Value != "matrix" {
		t.Fatalf("layout cookie = %+v, want matrix", cookie)
	}
}

func TestAppFragmentKeepsLoadingIndicators(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	html := getHTML(t, srv, "/partial/app").Body.String()

	headerMain := strings.Index(html, `class="header-main"`)
	fetchingHeader := strings.Index(html, `id="fetching-header"`)
	if headerMain == -1 || fetchingHeader == -1 || fetchingHeader < headerMain {
		t.Errorf("fetching-header must live inside header-main (header=%d fetching=%d)", headerMain, fetchingHeader)
	}
	panel := strings.Index(html, `id="panel"`)
	fetchingDetail := strings.Index(html, `id="fetching-detail"`)
	if panel == -1 || fetchingDetail == -1 || fetchingDetail < panel {
		t.Errorf("fetching-detail must live inside the panel (panel=%d fetching=%d)", panel, fetchingDetail)
	}
	footerMain := strings.Index(html, `class="footer-main"`)
	fetchingFooter := strings.Index(html, `id="fetching-footer"`)
	if footerMain == -1 || fetchingFooter == -1 || fetchingFooter < footerMain {
		t.Errorf("fetching-footer must live inside footer-main (footer=%d fetching=%d)", footerMain, fetchingFooter)
	}
	if !strings.Contains(html, `hx-indicator="#fetching-detail"`) {
		t.Error("the app fragment should wire htmx indicators")
	}
}

func TestInspectFragmentRendersCockpit(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	w := getHTML(t, srv, "/inspect?account=opencode-pro")
	if w.Code != http.StatusOK {
		t.Fatalf("inspect status = %d", w.Code)
	}
	body := w.Body.String()
	for _, want := range []string{`id="inspect-title"`, "opencode-pro", `class="hero"`, "Usage &amp; quotas"} {
		if !strings.Contains(body, want) {
			t.Errorf("inspect fragment missing %q", want)
		}
	}
}

func TestAppFragmentHonorsStoredCookies(t *testing.T) {
	srv := testServer(t, Options{Demo: true})

	req := httptest.NewRequest(http.MethodGet, "/partial/app", nil)
	req.Header.Set("Cookie", cookieLayout+"=matrix")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `data-layout="matrix"`) || !strings.Contains(body, `class="matrix-table"`) {
		t.Error("stored layout cookie should drive the rendered layout")
	}
	if !strings.Contains(body, `class="layout-btn active" data-layout="matrix"`) {
		t.Error("stored layout should be marked active in the header")
	}

	req = httptest.NewRequest(http.MethodGet, "/partial/app", nil)
	req.Header.Set("Cookie", cookieFilter+"=opencode")
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	body = w.Body.String()
	if !strings.Contains(body, "opencode-pro") || strings.Contains(body, "codex-cli") {
		t.Error("stored filter cookie should restrict the roster")
	}
	if !strings.Contains(body, "1 agents (filtered)") {
		t.Error("stored filter should be reflected in the header count")
	}
}

func TestViewActionSwitchesMobilePane(t *testing.T) {
	srv := testServer(t, Options{Demo: true})

	split := getHTML(t, srv, "/partial/app").Body.String()
	for _, want := range []string{"mobile-back-bar", "mobile-back-btn", "mobile-pager-count", "mobile-pager-btn", `hx-post="actions/view"`} {
		if !strings.Contains(split, want) {
			t.Errorf("split fragment missing mobile roster control %q", want)
		}
	}

	w := postForm(t, srv, "/actions/view", "view=roster")
	if w.Code != http.StatusOK {
		t.Fatalf("view action status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "mobile-view-roster") {
		t.Error("view action should render the roster pane immediately, not after the next request")
	}
	if strings.Contains(body, "mobile-view-detail") {
		t.Error("view action should drop the detail pane state")
	}

	back := postForm(t, srv, "/actions/view", "view=detail")
	if back.Code != http.StatusOK || !strings.Contains(back.Body.String(), "mobile-view-detail") {
		t.Error("view action should be able to return to the detail pane")
	}
}

func TestStylesheetKeepsLayoutHooks(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	w := getHTML(t, srv, "/app.css")
	if w.Code != http.StatusOK {
		t.Fatalf("app.css status = %d", w.Code)
	}
	css := w.Body.String()
	for _, want := range []string{
		".board-bars", ".board-dials", ".board-strips", ".lin-track", ".dial-svg",
		".strip-track", ".metric-table", ".gauge-group", ".bento-tile", ".matrix-table",
		".item.refreshing", ".fetching.htmx-request", ".sr-only",
		"prefers-reduced-motion",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("app.css missing %q", want)
		}
	}
	if strings.Contains(css, ".shell.refreshing .footer-main { display: none; }") {
		t.Error("app.css should not hide the footer while refreshing")
	}
}
