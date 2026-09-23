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

func TestBoards_ShowLimitReachedAndRealKPIBanner(t *testing.T) {
	env := Envelope{
		UsageMode: "remaining",
		Views: []AccountView{
			{
				Key: "cursor-exhausted", ProviderID: "cursor", ProviderName: "Cursor",
				AccountID: "cursor-exhausted", Status: "LIMITED", StatusBadge: "MONTHLY LIMIT",
				Summary: "0.0% remaining", HasGauge: true, GaugePercent: 0,
				UsageLines: []UsageLine{{Label: "Included", Short: "Included", Percent: f64(0), Tone: "crit", ResetIn: "13d12h"}},
			},
			{
				Key: "codex-cli", ProviderID: "codex", ProviderName: "Codex",
				AccountID: "codex-cli", Status: "OK", StatusBadge: "OK",
				Summary: "58.00% remaining", HasGauge: true, GaugePercent: 58,
				UsageLines: []UsageLine{{Label: "Five Hour Limit", Short: "5h", Percent: f64(58), Tone: "ok", ResetIn: "2h39m"}},
			},
		},
	}

	// Every board keeps the exhausted window unmistakable.
	for layout, wants := range map[string][]string{
		"split":  {"Limit reached", "tone-crit", "58.00% remaining"},
		"matrix": {"Limit reached", "58% left", "tone-crit"},
		"bento":  {"Limit reached", "58% left", "tone-crit"},
		"bars":   {"Limit reached", "58% left", "tone-crit"},
		"dials":  {`class="limit-flag">Limit reached<`, "tone-crit"},
		"strips": {`strip-metric tone-crit depleted`, "<em>Limit</em>"},
	} {
		html := renderFragment(t, env, renderInput{Layout: layout})
		for _, want := range wants {
			if !strings.Contains(html, want) {
				t.Errorf("%s layout missing %q", layout, want)
			}
		}
	}

	// KPI banner reports real fleet numbers, not placeholders.
	banner := renderFragment(t, env, renderInput{Layout: "bento"})
	for _, want := range []string{
		"Accounts At Limit", "Tightest Window", "Average Headroom",
		"0%", "29%", "across 2 quota windows", "cursor-exhausted",
	} {
		if !strings.Contains(banner, want) {
			t.Errorf("KPI banner missing %q", want)
		}
	}
	for _, banned := range []string{"3.48M", "$242.30", "88.4%", "Token Velocity", "Cache Hit"} {
		if strings.Contains(banner, banned) {
			t.Errorf("KPI banner still renders fabricated stat %q", banned)
		}
	}
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
	if strings.Contains(html, `data-theme="deep-space"`) {
		t.Error("shell should not hardcode deep-space theme")
	}
	if strings.Contains(html, `content="#0c0e16"`) {
		t.Error("shell should not hardcode #0c0e16 dark theme color")
	}
	if strings.Contains(html, `>aU</text>`) {
		t.Error("shell favicon should not use plain aU text mark")
	}
	if !strings.Contains(html, `M 9.5 21.5`) {
		t.Error("shell favicon should use vector squircle SVG emblem")
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
	wantCount := fmt.Sprintf("%d agents", len(env.Views))

	for _, want := range []string{
		`id="app" class="shell layout-split`,
		`id="nav" class="nav"`,
		`id="panel"`,
		`class="logo"`,
		wantCount,
		`class="item nav-item`,
		`class="hero"`,
		`class="header-count"`,
		`class="fleet-note"`,
		`id="footer-btn-mode"`,
		`id="footer-btn-refresh"`,
		`id="footer-btn-refresh-all"`,
		`id="footer-btn-providers"`,
		`id="footer-btn-keys"`,
		`id="footer-theme-select"`,
		`class="btn-label">refresh all</span>`,
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
		{"bento", []string{`class="bento-tiles-grid"`, `class="bento-tile`, `class="bento-quota-bar"`}},
		{"bars", []string{`class="board-grid board-bars"`, `class="lin-track"`}},
		{"dials", []string{`class="board-grid board-dials"`, `class="dial-svg"`, `class="card-footer"`, `class="card-inspect"`}},
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
			if tc.layout == "bento" || tc.layout == "bars" || tc.layout == "dials" || tc.layout == "strips" || tc.layout == "matrix" {
				if !strings.Contains(html, "provider-group-box") {
					t.Errorf("layout %s should use provider-group-box grouping", tc.layout)
				}
			}
			if tc.layout == "bento" {
				if !strings.Contains(html, "bento-wide") && !strings.Contains(html, "bento-compact") {
					t.Errorf("layout bento missing dynamic sizing classes (bento-wide/bento-compact)")
				}
			}
			if tc.layout == "matrix" {
				theadCount := strings.Count(html, "<thead>")
				if theadCount != 1 {
					t.Errorf("matrix layout should have exactly 1 <thead>, got %d", theadCount)
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
	if !strings.Contains(body, "1 agents") {
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

func TestProviderOrderAction_PersistsAndSupportsSilent(t *testing.T) {
	srv := testServer(t, Options{Demo: true})

	// 1. Silent save from smooth drag-and-drop
	wSilent := postForm(t, srv, "/actions/provider-order", "order=cursor|antigravity|codex&silent=1")
	if wSilent.Code != http.StatusOK {
		t.Fatalf("silent provider-order status = %d", wSilent.Code)
	}
	var cookie *http.Cookie
	for _, c := range wSilent.Result().Cookies() {
		if c.Name == cookieProviderOrder {
			cookie = c
		}
	}
	if cookie == nil || cookie.Value != "cursor%7Cantigravity%7Ccodex" {
		t.Fatalf("provider order cookie = %+v", cookie)
	}
	if cookie.MaxAge <= 0 {
		t.Errorf("cookie MaxAge = %d, want > 0 for persistence", cookie.MaxAge)
	}
	if !strings.Contains(wSilent.Body.String(), `"status":"ok"`) {
		t.Errorf("silent response body = %q, want json status ok", wSilent.Body.String())
	}

	// 2. Standard save with rendered HTML
	wStandard := postForm(t, srv, "/actions/provider-order", "order=antigravity|cursor")
	if wStandard.Code != http.StatusOK {
		t.Fatalf("standard provider-order status = %d", wStandard.Code)
	}
	if !strings.Contains(wStandard.Body.String(), "Provider order saved") {
		t.Errorf("standard response should contain toast, got: %s", wStandard.Body.String())
	}

	// 3. Fallback to in-memory/config order when cookie is absent
	req := httptest.NewRequest(http.MethodGet, "/partial/app", nil)
	gotOrder := srv.resolveProviderOrder(req)
	if gotOrder != "antigravity|cursor" {
		t.Errorf("resolveProviderOrder without cookie = %q, want antigravity|cursor", gotOrder)
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
	if !strings.Contains(body, `id="footer-btn-layout"`) || !strings.Contains(body, "Matrix") {
		t.Error("stored layout should be reflected by the footer views button label")
	}

	req = httptest.NewRequest(http.MethodGet, "/partial/app", nil)
	req.Header.Set("Cookie", cookieFilter+"=opencode")
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	body = w.Body.String()
	if !strings.Contains(body, "opencode-pro") || strings.Contains(body, "codex-cli") {
		t.Error("stored filter cookie should restrict the roster")
	}
	if !strings.Contains(body, "1 agents") {
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
		".strip-track", ".metric-table", ".gauge-group", ".bento-tile", ".bento-tiles-grid", ".matrix-table",
		".item.refreshing", ".fetching.htmx-request", ".sr-only",
		"prefers-reduced-motion",
		"padding: 14px 20px 56px;",
		"repeat(auto-fill, minmax(",
		"repeat(3, 1fr)",
		".bento-quota-row { display: grid; grid-template-columns: minmax(84px, auto)",
		".matrix-container .provider-group-box:not(:first-child) thead",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("app.css missing %q", want)
		}
	}
	if strings.Contains(css, "grid-template-columns: 64px 1fr auto;") {
		t.Error("app.css should not use rigid 64px label column in bento-quota-row")
	}
	if strings.Contains(css, ".shell.refreshing .footer-main { display: none; }") {
		t.Error("app.css should not hide the footer while refreshing")
	}
}

func renderFragment(t *testing.T, env Envelope, in renderInput) string {
	t.Helper()
	var buf strings.Builder
	model := buildRenderModel(env, in)
	if err := executeTemplate(&buf, "app", model); err != nil {
		t.Fatalf("renderFragment: %v", err)
	}
	return buf.String()
}

func renderInspect(t *testing.T, env Envelope, in renderInput) string {
	t.Helper()
	var buf strings.Builder
	model := buildRenderModel(env, in)
	if model.Selected == nil {
		t.Fatalf("renderInspect: model.Selected is nil")
	}
	if err := executeTemplate(&buf, "inspect", cockpitCtx{M: model, V: *model.Selected}); err != nil {
		t.Fatalf("renderInspect: %v", err)
	}
	return buf.String()
}

func TestErrorSanitization_SuppressesCannotRenderError(t *testing.T) {
	env := Envelope{
		Views: []AccountView{
			{
				Key:          "err-acct",
				ProviderID:   "codex",
				ProviderName: "OpenAI Codex CLI",
				AccountID:    "codex-cli",
				Status:       "OK",
				StatusBadge:  "ALL OK",
				Summary:      "cannot render as bar or graph",
				UsageLines: []UsageLine{
					{Label: "Rate", Value: "cannot render as bar or graph"},
					{Label: "cannot render as bar or graph", Value: "100"},
					{Label: "Tokens", Value: "45% used", Percent: f64(45)},
				},
				DetailCards: []DetailCard{
					{
						ID:    "only-errors",
						Title: "Diagnostic Errors",
						Rows: []DetailRow{
							{Kind: "kv", Label: "Err 1", Value: "cannot render as bar or graph"},
							{Kind: "kv", Label: "cannot render", Value: "failed"},
						},
					},
					{
						ID:    "mixed-card",
						Title: "Quota Details",
						Rows: []DetailRow{
							{Kind: "kv", Label: "Err Row", Value: "cannot render as bar or graph"},
							{Kind: "kv", Label: "Valid Row", Value: "99 reqs"},
						},
					},
				},
			},
		},
	}

	layouts := []string{"split", "matrix", "bento", "bars", "dials", "strips"}
	for _, layout := range layouts {
		t.Run(layout, func(t *testing.T) {
			html := renderFragment(t, env, renderInput{Layout: layout})
			if strings.Contains(strings.ToLower(html), "cannot render as bar or graph") {
				t.Errorf("layout %s leaked raw error string 'cannot render as bar or graph'", layout)
			}
			if strings.Contains(strings.ToLower(html), "cannot render") {
				t.Errorf("layout %s leaked 'cannot render'", layout)
			}
		})
	}

	t.Run("inspect", func(t *testing.T) {
		html := renderInspect(t, env, renderInput{Account: "codex-cli"})
		if strings.Contains(strings.ToLower(html), "cannot render as bar or graph") {
			t.Errorf("inspect leaked raw error string 'cannot render as bar or graph'")
		}
		if strings.Contains(strings.ToLower(html), "cannot render") {
			t.Errorf("inspect leaked 'cannot render'")
		}
		if strings.Contains(html, "only-errors") || strings.Contains(html, "Diagnostic Errors") {
			t.Errorf("card with all error rows should be dropped")
		}
		if !strings.Contains(html, "Valid Row") || !strings.Contains(html, "99 reqs") {
			t.Errorf("mixed card should retain valid rows")
		}
	})
}

func TestNonPercentageQuotaCalculations(t *testing.T) {
	t.Run("UsageLine Ratio Calculation", func(t *testing.T) {
		v := AccountView{
			ProviderID: "gemini_cli",
			AccountID:  "gemini-cli",
			UsageLines: []UsageLine{
				{Label: "Monthly Spend", Value: "$8.50 / $20.00"},
			},
		}
		lines := buildUsageLines(v, false)
		if len(lines) == 0 {
			t.Fatalf("expected usage lines, got none")
		}
		if lines[0].Pct == nil {
			t.Fatalf("expected computed Pct, got nil")
		}
		pct := *lines[0].Pct
		if pct == 0 {
			t.Errorf("expected non-zero percentage, got 0")
		}
		if pct != 57.5 {
			t.Errorf("expected 57.5%% remaining, got %f", pct)
		}

		linesUsed := buildUsageLines(v, true)
		if len(linesUsed) == 0 || linesUsed[0].Pct == nil {
			t.Fatalf("expected computed Pct in used mode")
		}
		if *linesUsed[0].Pct != 42.5 {
			t.Errorf("expected 42.5%% used, got %f", *linesUsed[0].Pct)
		}
	})

	t.Run("Summary Ratio Calculation When No UsageLines", func(t *testing.T) {
		v := AccountView{
			ProviderID: "gemini_cli",
			AccountID:  "gemini-cli",
			Summary:    "$8.50 / $20.00 · in 08d",
		}
		lines := buildUsageLines(v, false)
		if len(lines) == 0 {
			t.Fatalf("expected usage lines, got none")
		}
		if lines[0].Pct == nil {
			t.Fatalf("expected computed Pct, got nil")
		}
		pct := *lines[0].Pct
		if pct == 0 {
			t.Errorf("expected non-zero percentage, got 0")
		}
		if pct != 57.5 {
			t.Errorf("expected 57.5%% remaining, got %f", pct)
		}
	})

	t.Run("Rendered Layout Gauges Not Zero", func(t *testing.T) {
		env := Envelope{
			Views: []AccountView{
				{
					Key:          "gemini-acct",
					ProviderID:   "gemini_cli",
					ProviderName: "Gemini CLI",
					AccountID:    "gemini-cli",
					Status:       "OK",
					StatusBadge:  "OK",
					Summary:      "$8.50 / $20.00",
				},
			},
		}

		barsHTML := renderFragment(t, env, renderInput{Layout: "bars"})
		if !strings.Contains(barsHTML, `class="lin-track"`) {
			t.Errorf("bars layout missing lin-track gauge")
		}
		if strings.Contains(barsHTML, `width:0%`) || strings.Contains(barsHTML, `width:0.0%`) {
			t.Errorf("bars layout should not render 0%% width progress bar")
		}
		if !strings.Contains(barsHTML, `width:57.5%`) && !strings.Contains(barsHTML, `width:58%`) {
			t.Errorf("bars layout missing expected 57.5%% or 58%% progress bar")
		}

		dialsHTML := renderFragment(t, env, renderInput{Layout: "dials"})
		if !strings.Contains(dialsHTML, `class="dial-fill"`) {
			t.Errorf("dials layout missing dial-fill gauge")
		}
		if strings.Contains(dialsHTML, `stroke-dasharray="0 100"`) {
			t.Errorf("dials layout should not render 0 dasharray")
		}
		if !strings.Contains(dialsHTML, `stroke-dasharray="58 100"`) {
			t.Errorf("dials layout missing expected stroke-dasharray='58 100'")
		}

		stripsHTML := renderFragment(t, env, renderInput{Layout: "strips"})
		if !strings.Contains(stripsHTML, `class="strip-track"`) {
			t.Errorf("strips layout missing strip-track")
		}
		if !strings.Contains(stripsHTML, `width:58%`) {
			t.Errorf("strips layout missing expected 58%% gauge width")
		}

		splitHTML := renderFragment(t, env, renderInput{Layout: "split"})
		if !strings.Contains(splitHTML, `class="mini"`) || !strings.Contains(splitHTML, `width:58%`) {
			t.Errorf("split nav missing mini gauge capsule with 58%% width")
		}
		if !strings.Contains(splitHTML, `58%`) {
			t.Errorf("split nav missing 58%% percentage stat")
		}

		envWithReset := Envelope{
			Views: []AccountView{
				{
					Key:          "gemini-acct-reset",
					ProviderID:   "gemini_cli",
					ProviderName: "Gemini CLI",
					AccountID:    "gemini-cli",
					Status:       "OK",
					StatusBadge:  "OK",
					Summary:      "$8.50 / $20.00 · in 08d",
				},
			},
		}
		stripsWithResetHTML := renderFragment(t, envWithReset, renderInput{Layout: "strips"})
		if strings.Contains(stripsWithResetHTML, `<span class="strip-lab">In</span>`) {
			t.Errorf("strips layout should not render orphan 'In' label")
		}
		if strings.Contains(stripsWithResetHTML, `<th>In</th>`) {
			t.Errorf("table should not render orphan 'In' header")
		}
	})
}

func TestFilterNoMatch_PreservesAppShell(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	w := getHTML(t, srv, "/partial/app?q=zzz-nothing")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()

	// 1. #app does NOT have hidden attribute
	appIdx := strings.Index(body, `id="app"`)
	if appIdx == -1 {
		t.Fatal("missing #app element")
	}
	appClose := strings.Index(body[appIdx:], ">")
	if appClose == -1 {
		t.Fatal("malformed #app element tag")
	}
	appTag := body[appIdx : appIdx+appClose]
	if strings.Contains(appTag, "hidden") {
		t.Errorf("#app should not have hidden on filter mismatch: %s", appTag)
	}

	// 2. #empty-state DOES have hidden attribute
	emptyIdx := strings.Index(body, `id="empty-state"`)
	if emptyIdx == -1 {
		t.Fatal("missing #empty-state element")
	}
	emptyClose := strings.Index(body[emptyIdx:], ">")
	if emptyClose == -1 {
		t.Fatal("malformed #empty-state element tag")
	}
	emptyTag := body[emptyIdx : emptyIdx+emptyClose]
	if !strings.Contains(emptyTag, "hidden") {
		t.Errorf("#empty-state must have hidden attribute on filter mismatch: %s", emptyTag)
	}

	// 3. Rendered HTML contains .empty-filter-state, No providers matching "zzz-nothing", and the clear button
	for _, want := range []string{
		`class="empty-filter-state"`,
		`class="empty-filter-icon"`,
		`class="empty-filter-msg"`,
		`No providers matching "zzz-nothing"`,
		`class="btn-clear-filter"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing expected empty filter state marker %q", want)
		}
	}

	// 4. Header search box reflects query and contains .btn-search-clear
	for _, want := range []string{
		`class="search-box has-query"`,
		`class="btn-search-clear"`,
		`zzz-nothing`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing expected search box marker %q", want)
		}
	}
}

func TestThemeSynchronization_BrandStudioPillAndAppleLogo(t *testing.T) {
	srv := testServer(t, Options{Demo: true, Theme: "Ceramic Studio"})
	w := getHTML(t, srv, "/partial/app")
	if w.Code != http.StatusOK {
		t.Fatalf("GET /partial/app status = %d", w.Code)
	}
	html := w.Body.String()

	// 1. Apple Logo SVG squircle emblem
	for _, want := range []string{
		`<span class="logo" aria-hidden="true"><svg class="au-apple-logo"`,
		`id="auSquircleGrad"`,
		`id="auGaugeGrad"`,
		`id="auBorderGrad"`,
		`<path d="M 8.5 16.5 L 12 16.5 L 14 12.5 L 17 20.5 L 19 16.5 L 23.5 16.5"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("app fragment missing apple-logo marker %q", want)
		}
	}

	// 2. Dynamic brand-studio pill renders active theme
	activeTheme := envelopeTheme(t, srv)
	wantInitial := fmt.Sprintf(`<span class="brand-studio">%s</span>`, activeTheme)
	if !strings.Contains(html, wantInitial) {
		t.Errorf("app fragment missing initial brand-studio theme name %q, got %s", wantInitial, html)
	}

	// 3. Changing theme updates brand-studio pill dynamically
	themeResp := postForm(t, srv, "/actions/theme", "theme=Nord")
	if themeResp.Code != http.StatusOK {
		t.Fatalf("actions/theme status = %d", themeResp.Code)
	}
	themeHTML := themeResp.Body.String()
	if !strings.Contains(themeHTML, `<span class="brand-studio">Nord</span>`) {
		t.Errorf("expected dynamic brand-studio pill <span class=\"brand-studio\">Nord</span> after theme switch, got html excerpt: %s", themeHTML[:500])
	}

	// 4. Fallback when ThemeTokens.Name is empty
	fallbackHTML := renderFragment(t, Envelope{}, renderInput{})
	if !strings.Contains(fallbackHTML, `<span class="brand-studio">Ceramic Studio</span>`) {
		t.Errorf("expected fallback to Ceramic Studio when ThemeTokens.Name is empty, got %s", fallbackHTML)
	}
}

func TestCleanQuotaLabelsAndNoDuplicates(t *testing.T) {
	// Unit tests for cleanQuotaLabel
	tests := []struct {
		pill     string
		label    string
		expected string
	}{
		{"5h", "5h", "5-Hour Limit"},
		{"5H", "5h", "5-Hour Limit"},
		{"5h", "5h Limit", "Limit"},
		{"5h", "5h - Claude Pro", "Claude Pro"},
		{"5h", "5h: Session", "Session"},
		{"24h", "24h", "Daily Quota"},
		{"7d", "7d", "Weekly Quota"},
		{"30d", "30d", "Monthly Quota"},
		{"Spend", "Spend", "Monthly Spend"},
		{"Tokens", "Tokens", "Token Quota"},
		{"Fast", "Fast", "Fast Quota"},
		{"Burst", "Burst", "Burst Quota"},
		{"RPM", "RPM", "Request Limit"},
		{"5h", "Claude 3.5 Sonnet", "Claude 3.5 Sonnet"},
	}

	for _, tc := range tests {
		got := cleanQuotaLabel(tc.pill, tc.label)
		if got != tc.expected {
			t.Errorf("cleanQuotaLabel(%q, %q) = %q, want %q", tc.pill, tc.label, got, tc.expected)
		}
	}

	// Rendering test: verify no duplicate "5h 5h" in lin-gauge and no duplicate reset caption
	env := Envelope{
		Views: []AccountView{
			{
				Key:          "test-quota-acct",
				ProviderID:   "claude_code",
				ProviderName: "Claude Code",
				AccountID:    "claude-test",
				Status:       "OK",
				StatusBadge:  "OK",
				UsageLines: []UsageLine{
					{Label: "5h", Short: "5h", Value: "80%", Percent: f64(80), ResetIn: "2h 15m"},
				},
			},
		},
	}

	barsHTML := renderFragment(t, env, renderInput{Layout: "bars"})
	if strings.Contains(barsHTML, `<span class="q-pill">5h</span> 5h<`) || strings.Contains(barsHTML, `<span class="q-pill">5h</span> 5h `) {
		t.Error("bars layout rendered duplicate '5h 5h' quota label")
	}
	if !strings.Contains(barsHTML, `<span class="q-pill">5h</span> 5-Hour Limit`) {
		t.Error("bars layout should render human-readable '5-Hour Limit' instead of duplicate 5h")
	}

	// lin-meta should not duplicate reset caption
	if strings.Contains(barsHTML, "Resets in 2h 15mResets in 2h 15m") || strings.Contains(barsHTML, "Resets in 2h 15m Resets in 2h 15m") {
		t.Error("lin-meta contains duplicate reset caption")
	}

	// bento should not duplicate "5h 5h"
	bentoHTML := renderFragment(t, env, renderInput{Layout: "bento"})
	if strings.Contains(bentoHTML, `<span class="q-pill">5h</span> 5h<`) || strings.Contains(bentoHTML, `<span class="q-pill">5h</span> 5h `) {
		t.Error("bento layout rendered duplicate '5h 5h' quota label")
	}
	if !strings.Contains(bentoHTML, `<span class="q-pill">5h</span> 5-Hour Limit`) {
		t.Error("bento layout should render human-readable '5-Hour Limit' instead of duplicate 5h")
	}
}

func fetchEnvelope(t *testing.T, srv *Server) Envelope {
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
	return env
}

func TestCockpit_RealUsageLinesReplaceFabricatedDecks(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	env := fetchEnvelope(t, srv)

	// No fabricated deck content may survive in any cockpit.
	for _, acct := range []string{"codex-cli", "command-code", "cursor-ide", "opencode-pro", "gemini-cli", "antigravity-mohammed"} {
		html := renderInspect(t, env, renderInput{Account: acct})
		for _, banned := range []string{
			"MODEL BURN", "CLIENTS</h2>", "CODE STATISTICS",
			"DAILY SPEND", "MONTHLY CREDITS", "TEAM BUDGET", "CONTAINER FLEET",
			"Primary Model", "command-r-plus",
		} {
			if strings.Contains(html, banned) {
				t.Errorf("%s cockpit still renders fabricated %q", acct, banned)
			}
		}
		if !strings.Contains(html, "Usage &amp; quotas") {
			t.Errorf("%s cockpit missing usage & quotas card", acct)
		}
		if strings.Contains(html, "Next reset:") {
			t.Errorf("%s cockpit has dangling 'Next reset:'", acct)
		}
	}

	// Real windows render as labelled gauges with mode-aware captions.
	ccHTML := renderInspect(t, env, renderInput{Account: "command-code"})
	for _, want := range []string{"Five Hour Limit", "100% left", "Week", "80% left", "Month", "49% left", "Resets in"} {
		if !strings.Contains(ccHTML, want) {
			t.Errorf("command-code cockpit missing real quota text %q", want)
		}
	}

	// Account without quota windows falls back to its real summary.
	fallbackEnv := Envelope{
		Views: []AccountView{
			{
				Key:          "generic-acct",
				ProviderID:   "custom",
				ProviderName: "Custom Agent",
				AccountID:    "custom-agent",
				Status:       "OK",
				StatusBadge:  "OK",
				Summary:      "$5.00 remaining",
				NextReset:    "",
			},
		},
	}
	fallbackHTML := renderInspect(t, fallbackEnv, renderInput{Account: "custom-agent"})
	if !strings.Contains(fallbackHTML, "$5.00 remaining") {
		t.Error("fallback cockpit should surface the real summary")
	}
	if strings.Contains(fallbackHTML, "Rolling window") {
		t.Error("fallback cockpit should not invent a rolling-window label")
	}

	// Unbudgeted account does not render a false 0.0% quota.
	unbudgetedEnv := Envelope{
		Views: []AccountView{
			{
				Key:          "unbudgeted-acct",
				ProviderID:   "custom",
				ProviderName: "Custom Agent",
				AccountID:    "unbudgeted-agent",
				Status:       "OK",
				StatusBadge:  "OK",
				Summary:      "Pay-as-you-go",
				HasGauge:     false,
				GaugePercent: -1,
			},
		},
	}
	unbudgetedHTML := renderInspect(t, unbudgetedEnv, renderInput{Account: "unbudgeted-agent"})
	if strings.Contains(unbudgetedHTML, ">0.0%<") {
		t.Error("unbudgeted account should not render misleading 0.0% quota")
	}
	if !strings.Contains(unbudgetedHTML, "Pay-as-you-go") {
		t.Error("unbudgeted account should render its real summary")
	}
	if strings.Contains(unbudgetedHTML, "ACTIVITY STATUS") || strings.Contains(unbudgetedHTML, "CYCLE STATUS") {
		t.Error("unbudgeted account should not render placeholder status cards")
	}

	// 4. btn-cockpit-refresh uses icon-refresh template, not raw unicode glyph
	if strings.Contains(fallbackHTML, `class="footer-btn btn-cockpit-refresh" hx-get="partial/app?refresh=1&amp;focus=custom-agent" hx-target="#app" hx-swap="outerHTML" title="Refresh account">⟳</button>`) {
		t.Error("btn-cockpit-refresh should use icon-refresh SVG instead of raw glyph ⟳")
	}
	if !strings.Contains(fallbackHTML, `class="footer-btn btn-cockpit-refresh"`) || !strings.Contains(fallbackHTML, `M13.2 8a5.2 5.2 0 1 1-1.7-3.9`) {
		t.Error("btn-cockpit-refresh missing icon-refresh SVG")
	}

	// 5. strip-card space between agent-name and agent-plan
	stripsHTML := renderFragment(t, env, renderInput{Layout: "strips"})
	if strings.Contains(stripsHTML, `</span><span class="agent-plan">`) {
		t.Error("strip-card should have space between agent-name and agent-plan to prevent concatenation")
	}
	if !strings.Contains(stripsHTML, `</span> <span class="agent-plan">`) {
		t.Error("strip-card missing space between agent-name and agent-plan")
	}

	// 6. No fabricated KPI cards survive; the demo cursor snapshot's real
	// attributes still surface through the info card.
	cursorHTML := renderInspect(t, env, renderInput{Account: "cursor-ide"})
	if strings.Contains(cursorHTML, "TEAM BUDGET") || strings.Contains(cursorHTML, "BILLING CYCLE") {
		t.Error("cursor cockpit still renders fabricated KPI cards")
	}
	if strings.Contains(fallbackHTML, "Active tier ·") {
		t.Error("fallback cockpit contains dangling dot 'Active tier ·'")
	}
}

func TestSwissModernistPolishStyles(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	w := getHTML(t, srv, "/app.css")
	if w.Code != http.StatusOK {
		t.Fatalf("app.css status = %d", w.Code)
	}
	css := w.Body.String()

	// 1. Frosted glass header styling on .top-nav .header-main
	for _, want := range []string{
		"backdrop-filter: blur(20px);",
		"border: 1px solid var(--border-soft);",
		"box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("app.css missing frosted glass header style %q", want)
		}
	}
	if strings.Contains(css, "box-shadow: 4px 4px 0px var(--border-strong, #1a1a1c);") {
		t.Error("app.css still contains neo-brutalist header box-shadow 4px 4px 0px")
	}

	// 2. Accent line scoped to max-width 120px
	if !strings.Contains(css, ".accent-line {\n  max-width: 120px;") {
		t.Error("app.css missing max-width: 120px on .accent-line")
	}

	// 3. Apple-style secondary action icon button for .btn-cockpit-refresh
	for _, want := range []string{
		".btn-cockpit-refresh {",
		"width: 28px;",
		"height: 28px;",
		"border-radius: 8px;",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("app.css missing btn-cockpit-refresh style %q", want)
		}
	}

	// 4. Fabricated cockpit deck styles are gone
	for _, banned := range []string{".burn-details {", ".kpi-card {", ".clients-card", ".code-card"} {
		if strings.Contains(css, banned) {
			t.Errorf("app.css still ships dead fabricated-deck rule %q", banned)
		}
	}
}

func TestQuietDockControlSurfaces(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	w := getHTML(t, srv, "/partial/app")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	html := w.Body.String()

	// 1. Presence in rendered /partial/app
	for _, want := range []string{
		`class="footer-actions"`,
		`class="fleet-note"`,
		`class="live-dot"`,
		`id="footer-btn-mode"`,
		`id="footer-btn-refresh"`,
		`id="footer-btn-refresh-all"`,
		`id="footer-btn-layout"`,
		`id="footer-btn-providers"`,
		`id="footer-btn-keys"`,
		`id="footer-theme-select"`,
		`class="header-count"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered fragment missing %q", want)
		}
	}

	// 2. Absence in rendered HTML
	for _, banned := range []string{
		"kbd-hint",
		"footer-sep",
		"footer-btn-segmented",
		"dock-stream",
		"live-label",
		"footer-cadence-tag",
		"footer-mockup-tag",
		"status-tag",
		"header-meta",
		"footer-btn-filter",
		`class="footer-right"`,
	} {
		if strings.Contains(html, banned) {
			t.Errorf("rendered fragment still contains retired token %q", banned)
		}
	}

	// 3. No title= inside <nav class="footer-actions" ...>...</nav>
	navStart := strings.Index(html, `<nav class="footer-actions"`)
	if navStart == -1 {
		t.Fatal("missing <nav class=\"footer-actions\"")
	}
	navEndRel := strings.Index(html[navStart:], "</nav>")
	if navEndRel == -1 {
		t.Fatal("missing closing </nav> for footer-actions")
	}
	navBlock := html[navStart : navStart+navEndRel+len("</nav>")]
	if strings.Contains(navBlock, "title=") {
		t.Errorf("footer-actions nav block should not contain title= attribute: %s", navBlock)
	}

	// 4. CSS styling hooks in /app.css
	cssW := getHTML(t, srv, "/app.css")
	if cssW.Code != http.StatusOK {
		t.Fatalf("GET /app.css = %d", cssW.Code)
	}
	css := cssW.Body.String()

	for _, want := range []string{
		".footer-actions {",
		".footer-btn {",
		".footer-theme-select {",
		".fleet-note {",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("app.css missing footer rule %q", want)
		}
	}

	for _, banned := range []string{
		".footer-sep {",
		".footer-btn-segmented {",
		".dock-stream {",
		".kbd-hint",
		".status-tag {",
		".center-tabs",
	} {
		if strings.Contains(css, banned) {
			t.Errorf("app.css still contains retired rule %q", banned)
		}
	}
}

func TestCockpit_ExhaustedQuotaRendersLimitReached(t *testing.T) {
	// 1. Cursor account at 100% used in used-mode.
	cursorEnv := Envelope{
		UsageMode: "used",
		Views: []AccountView{
			{
				Key:          "cursor-nurulz",
				ProviderID:   "cursor",
				ProviderName: "Cursor",
				AccountID:    "cursor-nurulz",
				Status:       "LIMITED",
				StatusBadge:  "MONTHLY LIMIT",
				Summary:      "100.0% used",
				HasGauge:     true,
				GaugePercent: 100.0,
				UsageLines: []UsageLine{
					{
						Label:   "Plan Usage",
						Short:   "Plan",
						Percent: f64(100.0),
						Value:   "500 / 500 requests",
						Tone:    "crit",
					},
				},
			},
		},
	}
	cursorHTML := renderInspect(t, cursorEnv, renderInput{Account: "cursor-nurulz"})
	if !strings.Contains(cursorHTML, "500 / 500 requests") {
		t.Error("cursor cockpit missing the exhausted window value")
	}
	if strings.Contains(cursorHTML, "QUOTA REMAINING") {
		t.Error("cursor 100% used must not render deceptive 'QUOTA REMAINING'")
	}
	if !strings.Contains(cursorHTML, "tone-crit") {
		t.Error("cursor 100% used must render with tone-crit")
	}
	if !strings.Contains(cursorHTML, "Limit reached") {
		t.Error("cursor 100% used must render the 'Limit reached' caption")
	}

	// 2. Command Code account hit monthly limit with live numbers.
	cmdEnv := Envelope{
		UsageMode: "used",
		Views: []AccountView{
			{
				Key:          "command_code",
				ProviderID:   "command_code",
				ProviderName: "Command Code",
				AccountID:    "command_code",
				Status:       "LIMITED",
				StatusBadge:  "MONTHLY LIMIT",
				Summary:      "Monthly Limit Reached",
				HasGauge:     true,
				GaugePercent: 99.8,
				UsageLines: []UsageLine{
					{
						Label:   "Monthly Subscription",
						Short:   "Month",
						Percent: f64(99.8),
						Value:   "$69.88 / $70.00",
						Hint:    "$0.12 remaining",
						Tone:    "crit",
					},
					{
						Label:   "Weekly Allowance",
						Short:   "Week",
						Percent: f64(20.0),
						Value:   "$7.00 / $35.00",
						Hint:    "$28.00 remaining",
						Tone:    "ok",
					},
				},
			},
		},
	}
	cmdHTML := renderInspect(t, cmdEnv, renderInput{Account: "command_code"})
	if strings.Contains(cmdHTML, "MONTHLY CREDITS") {
		t.Error("command_code cockpit should not render the fabricated MONTHLY CREDITS card")
	}
	if !strings.Contains(cmdHTML, "$69.88 / $70.00") {
		t.Error("command_code should display live $69.88 / $70.00 spending")
	}
	if !strings.Contains(cmdHTML, "tone-crit") {
		t.Error("command_code exhausted monthly limit must render with tone-crit")
	}
	if !strings.Contains(cmdHTML, "Limit reached") {
		t.Error("command_code exhausted month must render the 'Limit reached' caption")
	}
	if !strings.Contains(cmdHTML, "20% used") {
		t.Error("command_code weekly window should keep its used-mode caption")
	}

	// 3. Cursor account at 0% remaining in remaining-mode.
	remEnv := Envelope{
		UsageMode: "remaining",
		Views: []AccountView{
			{
				Key:          "cursor-exhausted",
				ProviderID:   "cursor",
				ProviderName: "Cursor",
				AccountID:    "cursor-exhausted",
				Status:       "LIMITED",
				StatusBadge:  "MONTHLY LIMIT",
				Summary:      "0.0% remaining",
				HasGauge:     true,
				GaugePercent: 0.0,
				UsageLines: []UsageLine{
					{
						Label:   "Plan Usage",
						Short:   "Plan",
						Percent: f64(0.0),
						Value:   "0 / 500 requests",
						Tone:    "crit",
					},
				},
			},
		},
	}
	remHTML := renderInspect(t, remEnv, renderInput{Account: "cursor-exhausted"})
	if strings.Contains(remHTML, "QUOTA USED") {
		t.Error("cursor 0% remaining must not render 'QUOTA USED'")
	}
	if !strings.Contains(remHTML, "tone-crit") {
		t.Error("cursor 0% remaining must render with tone-crit")
	}
	if !strings.Contains(remHTML, "Limit reached") {
		t.Error("cursor 0% remaining must render the 'Limit reached' caption")
	}
}

// TestBentoRows_CollapseDuplicateWindows pins the bento glance contract: one bar per
// quota window. Providers that emit sibling model rows for the same window (antigravity
// Gemini flash + pro "5-Hour Limit") collapse to the tightest bar, and the per-window
// reset chip survives the collapse. Distinct plan buckets (Included/Auto/API) stay
// separate, and at most three bars render.
func TestBentoRows_CollapseDuplicateWindows(t *testing.T) {
	tests := []struct {
		name      string
		used      bool
		lines     []usageLine
		wantRows  int
		wantReset map[int]string
		wantPct   map[int]float64
	}{
		{
			name: "antigravity duplicate 5h collapses to tightest with reset",
			lines: []usageLine{
				{Label: "5-Hour Limit", Short: "5h", Group: "Gemini", Pct: f64(35), ResetIn: "4h12m"},
				{Label: "5-Hour Limit", Short: "5h", Group: "Claude / GPT", Pct: f64(65)},
				{Label: "Week", Short: "wk", Group: "Gemini", Pct: f64(10), ResetIn: "2d03h"},
			},
			used:      true,
			wantRows:  2,
			wantPct:   map[int]float64{0: 65, 1: 10},
			wantReset: map[int]string{0: "4h12m", 1: "2d03h"},
		},
		{
			name: "distinct plan buckets keep three rows",
			lines: []usageLine{
				{Label: "Included", Short: "Included", Pct: f64(100), ResetIn: "9d14h"},
				{Label: "Auto", Short: "Auto", Pct: f64(41), ResetIn: "9d14h"},
				{Label: "API", Short: "API", Pct: f64(0), ResetIn: "9d14h"},
			},
			wantRows: 3,
		},
		{
			name: "duplicate collapse merges urgent and reset across siblings",
			lines: []usageLine{
				{Label: "5-Hour Limit", Short: "5h", Pct: f64(90), Urgent: true, ResetIn: "20m"},
				{Label: "5-Hour Limit", Short: "5h", Pct: f64(10), ResetIn: "4h"},
			},
			used:      true,
			wantRows:  1,
			wantPct:   map[int]float64{0: 90},
			wantReset: map[int]string{0: "20m"},
		},
		{
			name: "duplicate without reset inherits sibling reset",
			lines: []usageLine{
				{Label: "Week", Short: "wk", Pct: f64(0)},
				{Label: "Week", Short: "wk", Pct: f64(5), ResetIn: "1d02h"},
			},
			used:      true,
			wantRows:  1,
			wantPct:   map[int]float64{0: 5},
			wantReset: map[int]string{0: "1d02h"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rows := bentoRows(tc.lines, tc.used)
			if len(rows) != tc.wantRows {
				t.Fatalf("bentoRows len = %d, want %d", len(rows), tc.wantRows)
			}
			for i, want := range tc.wantPct {
				if rows[i].Pct == nil || *rows[i].Pct != want {
					t.Errorf("row %d pct = %v, want %v", i, rows[i].Pct, want)
				}
			}
			for i, want := range tc.wantReset {
				if rows[i].Reset != want {
					t.Errorf("row %d reset = %q, want %q", i, rows[i].Reset, want)
				}
			}
		})
	}
}

// TestBentoTile_RendersPerWindowResetChips proves the bento tile shows one reset
// timer per quota window row, not a single foot-level hint.
func TestBentoTile_RendersPerWindowResetChips(t *testing.T) {
	env := Envelope{
		Views: []AccountView{
			{
				Key: "oc", ProviderID: "opencode", ProviderName: "OpenCode",
				AccountID: "opencode-nurulz", Status: "OK", StatusBadge: "OK",
				UsageLines: []UsageLine{
					{Label: "Five Hour Limit", Short: "5h", Percent: f64(1), Tone: "ok", ResetIn: "4h46m"},
					{Label: "Week", Short: "wk", Percent: f64(7), Tone: "ok", ResetIn: "2d13h"},
				},
			},
		},
	}
	html := renderFragment(t, env, renderInput{Layout: "bento"})
	if got := strings.Count(html, `class="bento-row-reset`); got != 2 {
		t.Errorf("bento tile reset chips = %d, want 2 (one per window)", got)
	}
	if !strings.Contains(html, `>4h46m</span>`) {
		t.Error("bento tile missing 5h window reset chip 4h46m")
	}
	if !strings.Contains(html, `>2d13h</span>`) {
		t.Error("bento tile missing weekly window reset chip 2d13h")
	}
}

// TestTimebandLabel_FiveHourSpelledOut pins the pill for spelled-out window names
// ("Five Hour Limit") so bento tiles label the 5h window correctly.
func TestTimebandLabel_FiveHourSpelledOut(t *testing.T) {
	cases := map[string]string{
		"Five Hour Limit": "5h",
		"5-Hour Limit":    "5h",
		"Weekly Limit":    "7d",
		"Week":            "7d",
		"Monthly Limit":   "30d",
		"Month":           "30d",
	}
	for in, want := range cases {
		if got := timebandLabel(in); got != want {
			t.Errorf("timebandLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestMatrixLines_CollapseDuplicateWindows pins the matrix roster contract: one quota
// column per window (antigravity's sibling "5-Hour Limit" rows collapse to the tightest),
// with the per-window reset carried so the cell chip can render.
func TestMatrixLines_CollapseDuplicateWindows(t *testing.T) {
	v := AccountView{ProviderID: "antigravity", AccountID: "antigravity-x"}
	lines := []usageLine{
		{Label: "5-Hour Limit", Short: "5h", Group: "Gemini", Pct: f64(35), ResetIn: "4h12m"},
		{Label: "5-Hour Limit", Short: "5h", Group: "Claude / GPT", Pct: f64(65)},
		{Label: "Week", Short: "wk", Group: "Gemini", Pct: f64(10), ResetIn: "2d03h"},
	}
	out := matrixLines(v, lines, true)
	if len(out) != 2 {
		t.Fatalf("matrixLines len = %d, want 2 (5h + weekly)", len(out))
	}
	if out[0].Pct == nil || *out[0].Pct != 65 {
		t.Errorf("collapsed 5h pct = %v, want 65 (tightest used)", out[0].Pct)
	}
	if out[0].ResetIn != "4h12m" {
		t.Errorf("collapsed 5h ResetIn = %q, want 4h12m (merged from sibling)", out[0].ResetIn)
	}
	if out[1].Label != "Week" {
		t.Errorf("second row = %q, want Week", out[1].Label)
	}

	cursor := []usageLine{
		{Label: "Included", Short: "Included", Pct: f64(100)},
		{Label: "Auto", Short: "Auto", Pct: f64(41)},
		{Label: "API", Short: "API", Pct: f64(0)},
	}
	if out := matrixLines(v, cursor, true); len(out) != 3 {
		t.Errorf("distinct plan buckets collapsed: got %d rows, want 3", len(out))
	}
}

// TestMatrixRow_ShowsPerWindowResetChips proves each matrix quota cell carries its
// window's reset timer (5h / weekly / monthly), not just the single NEXT RESET column.
func TestMatrixRow_ShowsPerWindowResetChips(t *testing.T) {
	env := Envelope{
		Views: []AccountView{
			{
				Key: "oc", ProviderID: "opencode", ProviderName: "OpenCode",
				AccountID: "opencode-nurulz", Status: "OK", StatusBadge: "OK",
				UsageLines: []UsageLine{
					{Label: "Five Hour Limit", Short: "5h", Percent: f64(1), Tone: "ok", ResetIn: "4h46m"},
					{Label: "Week", Short: "wk", Percent: f64(7), Tone: "ok", ResetIn: "2d13h"},
					{Label: "Month", Short: "mo", Percent: f64(47), Tone: "ok", ResetIn: "6d1h"},
				},
			},
		},
	}
	html := renderFragment(t, env, renderInput{Layout: "matrix"})
	if got := strings.Count(html, `class="matrix-cell-reset`); got != 3 {
		t.Errorf("matrix row reset chips = %d, want 3 (one per window)", got)
	}
	for _, want := range []string{">4h46m<", ">2d13h<", ">6d1h<"} {
		if !strings.Contains(html, want) {
			t.Errorf("matrix missing per-window reset chip %q", want)
		}
	}
}

func TestMatrixLines_SpendLineWithDollarLabelSurvives(t *testing.T) {
	v := AccountView{ProviderID: "command_code", AccountID: "command_code"}
	lines := []usageLine{
		{Label: "Five Hour Limit Used ($0.00 / $14.00 used)", Short: "5h", Pct: f64(0), Value: "$0.00 / $14.00 used", Tone: "ok"},
		{Label: "Weekly Limit Used ($4.96 / $35.00 used)", Short: "7d", Pct: f64(14), ResetIn: "15h54m"},
		{Label: "Monthly Subscription Used ($69.88 / $70.00 used)", Short: "30d", Pct: f64(100), ResetIn: "4d13h", Depleted: true},
	}
	out := matrixLines(v, lines, true)
	if len(out) != 3 {
		t.Fatalf("matrixLines len = %d, want 3; rows: %+v", len(out), out)
	}
	if out[0].Short != "5h" {
		t.Errorf("first row Short = %q, want 5h", out[0].Short)
	}
}

func TestMatrixLines_ExactLiveCommandCodeLines(t *testing.T) {
	v := AccountView{ProviderID: "command_code", AccountID: "command_code"}
	lines := []usageLine{
		{Label: "Five Hour Limit Used ($0.00 / $14.00 used)", Short: "5h", Pct: f64(0), Tone: "ok"},
		{Label: "Weekly Limit Used ($4.96 / $35.00 used)", Short: "Week", Pct: f64(14.16), ResetIn: "15h52m", Hint: "⏱ Resets in 15h 52m"},
		{Label: "Monthly Subscription Used ($69.88 / $70.00 used)", Short: "Month", Pct: f64(99.83), ResetIn: "4d13h", Group: "Subscription", Hint: "⏱ Resets in 4d 13h", Depleted: true},
	}
	out := matrixLines(v, lines, true)
	if len(out) != 3 {
		t.Fatalf("matrixLines len = %d, want 3; got shorts: %v", len(out), shortsOf(out))
	}
	if len(bentoRows(lines, true)) != 3 {
		t.Fatalf("bentoRows len = %d, want 3; got shorts: %v", len(bentoRows(lines, true)), shortsOf(lines))
	}
}

func shortsOf(lines []usageLine) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = l.Short
	}
	return out
}

func TestAntigravity_ExhaustedClaudeDoesNotDepleteGeminiAccount(t *testing.T) {
	for _, mode := range []string{"used", "remaining"} {
		t.Run("mode_"+mode, func(t *testing.T) {
			usedMode := mode == "used"
			var g5, gw, c5, cw float64
			if usedMode {
				g5, gw, c5, cw = 5, 61, 0, 100
			} else {
				g5, gw, c5, cw = 95, 39, 100, 0
			}

			env := Envelope{
				UsageMode: mode,
				Views: []AccountView{
					{
						Key: "antigravity-chaos", ProviderID: "antigravity", ProviderName: "Antigravity",
						AccountID: "antigravity-chaos", Status: "OK", StatusBadge: "OK",
						Detail: "Gemini 3.7 Flash (High)",
						DetailCards: []DetailCard{
							{Rows: []DetailRow{{Label: "Model", Value: "Gemini 3.7 Flash (High)"}}},
						},
						UsageLines: []UsageLine{
							{Label: "Five Hour Limit", Short: "5h", Group: "Gemini", Percent: f64(g5), Tone: "ok", ResetIn: "4h"},
							{Label: "Weekly Limit", Short: "Weekly", Group: "Gemini", Percent: f64(gw), Tone: "peach", ResetIn: "1d 2h"},
							{Label: "Five Hour Limit", Short: "5h", Group: "Claude / GPT", Percent: f64(c5), Tone: "ok", ResetIn: "5h"},
							{Label: "Weekly Limit", Short: "Weekly", Group: "Claude / GPT", Percent: f64(cw), Tone: "crit", ResetIn: "5d 14h"},
						},
					},
				},
			}

			model := buildRenderModel(env, renderInput{Layout: "bento"})
			if len(model.Views) != 1 {
				t.Fatalf("expected 1 view, got %d", len(model.Views))
			}
			v := model.Views[0]

			// 1. Account should NOT be marked depleted or alert
			if v.Depleted {
				t.Errorf("expected v.Depleted to be false, got true")
			}
			if v.IsAlert {
				t.Errorf("expected v.IsAlert to be false, got true")
			}

			// 2. BentoRows must collapse duplicate windows to Gemini (active pool), NOT Claude
			if len(v.BentoRows) != 2 {
				t.Fatalf("expected 2 bento rows, got %d", len(v.BentoRows))
			}
			for _, row := range v.BentoRows {
				if row.Depleted {
					t.Errorf("bento row %s is depleted, expected active pool (Gemini)", row.Label)
				}
			}

			// 3. GlobalStats KPI banner must NOT report the account at limit
			if model.Stats.AtLimit != "0" {
				t.Errorf("expected AtLimit to be '0', got %q", model.Stats.AtLimit)
			}
			if model.Stats.Tightest == "100%" || model.Stats.Tightest == "0%" {
				t.Errorf("expected Tightest window to NOT be depleted, got %q", model.Stats.Tightest)
			}

			// 4. Cockpit graphs must contain both Gemini and Claude groups
			if len(v.Graphs) != 2 {
				t.Fatalf("expected 2 graph groups (Gemini and Claude), got %d", len(v.Graphs))
			}
			if v.Graphs[0].Title != "Gemini" {
				t.Errorf("expected first graph group to be Gemini, got %q", v.Graphs[0].Title)
			}
			if v.Graphs[1].Title != "Claude / GPT" {
				t.Errorf("expected second graph group to be Claude / GPT, got %q", v.Graphs[1].Title)
			}

			// 5. HTML fragment rendering: must not show "Limit reached" in bento
			html := renderFragment(t, env, renderInput{Layout: "bento"})
			if strings.Contains(html, "Limit reached") {
				t.Errorf("bento HTML unexpectedly contains 'Limit reached'")
			}
		})
	}
}

