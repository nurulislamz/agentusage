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
			if tc.layout == "bento" || tc.layout == "bars" || tc.layout == "dials" {
				if strings.Contains(html, "provider-group-box") {
					t.Errorf("layout %s should not use provider-group-box boxing in glance views", tc.layout)
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
	})
}

func TestLayoutBtnActiveContrast(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	w := getHTML(t, srv, "/app.css")
	if w.Code != http.StatusOK {
		t.Fatalf("app.css status = %d", w.Code)
	}
	css := w.Body.String()

	if !strings.Contains(css, ".layout-btn.active") {
		t.Fatal("app.css missing .layout-btn.active")
	}

	for _, want := range []string{
		"var(--fg",
		"var(--bg",
		"font-weight: 600",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("app.css .layout-btn.active missing expected high-contrast token/property %q", want)
		}
	}

	if strings.Contains(css, ".layout-btn.active { background: var(--surface2); color: #ffffff;") {
		t.Errorf("app.css should not use low-contrast #ffffff on var(--surface2) for .layout-btn.active")
	}
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

func TestCockpitDashboard_NoDanglingNextResetAndCodexCommandHeroStats(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	env := fetchEnvelope(t, srv)

	// 1. Inspect codex-cli
	codexHTML := renderInspect(t, env, renderInput{Account: "codex-cli"})
	if !strings.Contains(codexHTML, "openai · OpenAI Codex CLI · dev@acme-corp.dev") {
		t.Errorf("codex-cli missing hero meta, got: %s", codexHTML[:300])
	}
	if !strings.Contains(codexHTML, "DAILY SPEND") {
		t.Error("codex-cli should render DAILY SPEND card instead of generic USAGE QUOTA")
	}
	if !strings.Contains(codexHTML, "gpt-5.1-codex") {
		t.Error("codex-cli should render gpt-5.1-codex in model burn")
	}
	if strings.Contains(codexHTML, "Next reset:") {
		t.Error("codex-cli should not have dangling 'Next reset:'")
	}

	// 2. Inspect command-code
	cmdHTML := renderInspect(t, env, renderInput{Account: "command-code"})
	if !strings.Contains(cmdHTML, "command-code · GOAT · dev@acme-corp.dev") {
		t.Errorf("command-code missing hero meta, got: %s", cmdHTML[:300])
	}
	if !strings.Contains(cmdHTML, "MONTHLY CREDITS") {
		t.Error("command-code should render MONTHLY CREDITS card instead of generic USAGE QUOTA")
	}
	if !strings.Contains(cmdHTML, "command-r-plus") {
		t.Error("command-code should render command-r-plus in model burn")
	}

	// 3. Fallback account with no CycleSchedule and empty NextDisplay
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
	if strings.Contains(fallbackHTML, "Next reset:") {
		t.Error("fallback account must not render dangling 'Next reset:'")
	}
	if !strings.Contains(fallbackHTML, "Rolling window") {
		t.Error("fallback account with empty NextDisplay should render 'Rolling window'")
	}

	// 4. btn-cockpit-refresh uses icon-refresh template, not raw unicode glyph
	if strings.Contains(fallbackHTML, `class="footer-btn btn-cockpit-refresh" hx-get="partial/app?refresh=1&amp;focus=custom-agent" hx-target="#app" hx-swap="outerHTML" title="Refresh account">⟳</button>`) {
		t.Error("btn-cockpit-refresh should use icon-refresh SVG instead of raw glyph ⟳")
	}
	if !strings.Contains(fallbackHTML, `class="footer-btn btn-cockpit-refresh"`) || !strings.Contains(fallbackHTML, `M13.2 8a5.2 5.2 0 1 1-1.7-3.9`) {
		t.Error("btn-cockpit-refresh missing icon-refresh SVG")
	}

	// 5. burn-card collapsible details
	cursorHTML := renderInspect(t, env, renderInput{Account: "cursor-ide"})
	if !strings.Contains(cursorHTML, `<details open class="burn-details">`) || !strings.Contains(cursorHTML, `<summary class="burn-summary"><h2><span class="burn-chevron">▸</span> MODEL BURN</h2></summary>`) {
		t.Error("cursor-ide burn card missing collapsible burn-details and burn-summary elements")
	}

	// 6. strip-card space between agent-name and agent-plan
	stripsHTML := renderFragment(t, env, renderInput{Layout: "strips"})
	if strings.Contains(stripsHTML, `</span><span class="agent-plan">`) {
		t.Error("strip-card should have space between agent-name and agent-plan to prevent concatenation")
	}
	if !strings.Contains(stripsHTML, `</span> <span class="agent-plan">`) {
		t.Error("strip-card missing space between agent-name and agent-plan")
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

	// 2. Accent line scoped to max-width 240px
	if !strings.Contains(css, ".accent-line {\n  max-width: 240px;") {
		t.Error("app.css missing max-width: 240px on .accent-line")
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

	// 4. Collapsible toggle affordances for burn-details and burn-summary
	for _, want := range []string{
		".burn-details {",
		".burn-summary {",
		".burn-chevron {",
		".burn-details[open] .burn-chevron {",
		"transform: rotate(90deg);",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("app.css missing burn card collapsible style %q", want)
		}
	}
}

