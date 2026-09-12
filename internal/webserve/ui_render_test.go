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
		{"bento", []string{`class="bento-tiles-grid"`, `class="bento-tile`, `class="bento-quota-bar"`}},
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
			if tc.layout == "bento" || tc.layout == "bars" || tc.layout == "dials" {
				if strings.Contains(html, "provider-group-box") {
					t.Errorf("layout %s should not use provider-group-box boxing in glance views", tc.layout)
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
	} {
		if !strings.Contains(css, want) {
			t.Errorf("app.css missing %q", want)
		}
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
