package webserve

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func testServer(t *testing.T, opts Options) *Server {
	t.Helper()
	if opts.ListenAddr == "" {
		opts.ListenAddr = "127.0.0.1:0"
	}
	if opts.Collect == nil && !opts.Demo {
		opts.Demo = true
	}
	opts.RefreshSeconds = 30
	opts.Version = "test"
	opts.TimeWindow = "7d"
	opts.Theme = "Gruvbox"
	srv, err := NewServer(opts)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	return srv
}

func TestHealthz(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q", body["status"])
	}
	if body["source"] != "demo" {
		t.Errorf("source = %q, want demo", body["source"])
	}
}

func TestSnapshotsDemoStripsRaw(t *testing.T) {
	srv := testServer(t, Options{
		Collect: func() (Envelope, error) {
			snap := core.NewUsageSnapshot("openai", "personal")
			snap.Status = core.StatusOK
			snap.Raw = map[string]string{"authorization": "sk-secret"}
			snap.Metrics["today_cost"] = core.Metric{Used: f64(1.25), Unit: "USD"}
			return Envelope{Source: "demo", Snapshots: []core.UsageSnapshot{snap}}, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body %s", w.Code, w.Body.String())
	}
	var env Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.SchemaVersion != schemaVersion {
		t.Errorf("schema = %q", env.SchemaVersion)
	}
	if env.AgentUsageVersion != "test" {
		t.Errorf("version = %q", env.AgentUsageVersion)
	}
	if len(env.Snapshots) != 1 {
		t.Fatalf("snapshots = %d", len(env.Snapshots))
	}
	if env.Snapshots[0].Raw != nil {
		t.Errorf("raw map should be stripped, got %#v", env.Snapshots[0].Raw)
	}
	if env.Snapshots[0].Metrics["today_cost"].Used == nil {
		t.Error("expected today_cost to survive sanitization")
	}
	if len(env.Catalog) == 0 {
		t.Error("expected provider catalog")
	}
}

func TestSnapshotsRequiresAuth(t *testing.T) {
	srv := testServer(t, Options{Demo: true, AuthToken: "s3cret"})
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want 401", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	req.Header.Set("Authorization", "Bearer s3cret")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("authed status = %d, want 200", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("healthz should stay open, got %d", w.Code)
	}
}

func TestIndexServed(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	html := string(body)
	if !strings.Contains(html, "agentUsage") {
		t.Error("index.html should mention agentUsage")
	}
	if !strings.Contains(html, `id="nav"`) || !strings.Contains(html, `id="panel"`) {
		t.Error("index.html should use native nav/panel chrome")
	}
	if strings.Contains(html, "tui-frame") {
		t.Error("index.html should not paint a TUI frame")
	}
	if strings.Contains(html, `src="/app.js"`) || strings.Contains(html, `href="/app.css"`) {
		t.Error("index.html should not load assets with root-absolute URLs")
	}
	if !strings.Contains(html, `src="app.js"`) {
		t.Error("index.html should load app.js")
	}
}

func TestStaticAssets(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	for _, path := range []string{"/app.css", "/app.js"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("%s status = %d", path, w.Code)
		}
	}
}

func TestAppJSUsageModeKeyHandler(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("/app.js status = %d, want 200", w.Code)
	}
	body := w.Body.String()

	checks := []struct {
		desc    string
		pattern string
	}{
		{"modifier key guard", "ev.ctrlKey || ev.metaKey || ev.altKey"},
		{"form element guard", `matches("input, textarea, select")`},
		{"keydown 'u' case", `case "u":`},
		{"keydown 'U' case", `case "U":`},
		{"usage mode toggle call", "cycleUsageMode()"},
		{"footer button usage mode", `id="footer-btn-mode"`},
	}

	for _, tc := range checks {
		if !strings.Contains(body, tc.pattern) {
			t.Errorf("/app.js missing %s (expected pattern %q)", tc.desc, tc.pattern)
		}
	}
}

func TestAppJSRefreshKeyHandlers(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("/app.js status = %d, want 200", w.Code)
	}
	body := w.Body.String()

	checks := []struct {
		desc    string
		pattern string
	}{
		{"load opts.accountID query param", "account_id"},
		{"keydown 'r' focused account refresh", `accountID: filteredViews()[state.selected]?.account_id`},
		{"keydown 'R' refresh all", `case "R":`},
		{"footer button refresh title", `title="Refresh focused account (r) / all (R)"`},
		{"footer button focused refresh call", `accountID: filteredViews()[state.selected]?.account_id`},
	}

	for _, tc := range checks {
		if !strings.Contains(body, tc.pattern) {
			t.Errorf("/app.js missing %s (expected pattern %q)", tc.desc, tc.pattern)
		}
	}
}

func TestSnapshotsAccountIDParam(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots?refresh=1&account_id=cursor-main", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestNewServerRejectsPublicBind(t *testing.T) {
	_, err := NewServer(Options{ListenAddr: ":8080", Demo: true})
	if err == nil {
		t.Fatal("expected public bind without token to fail")
	}
}

func TestDemoSnapshotsShape(t *testing.T) {
	snaps := demoSnapshots(time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	if len(snaps) < 4 {
		t.Fatalf("expected several demo providers, got %d", len(snaps))
	}
	seen := map[string]bool{}
	for _, snap := range snaps {
		if snap.ProviderID == "" || snap.AccountID == "" {
			t.Fatalf("incomplete snapshot: %+v", snap)
		}
		if len(snap.Raw) != 0 {
			t.Errorf("%s demo snapshot should not carry raw maps", snap.ProviderID)
		}
		seen[snap.ProviderID] = true
	}
	for _, id := range []string{"claude_code", "cursor", "openrouter", "copilot"} {
		if !seen[id] {
			t.Errorf("missing demo provider %s", id)
		}
	}
}

func TestMetaEndpoint(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/meta", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestListenAndServeLoopback(t *testing.T) {
	srv := testServer(t, Options{Demo: true, ListenAddr: "127.0.0.1:0"})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe(ctx) }()

	deadline := time.Now().Add(2 * time.Second)
	var addr string
	for time.Now().Before(deadline) {
		addr = srv.Addr()
		host, port, err := net.SplitHostPort(addr)
		if err == nil && host == "127.0.0.1" && port != "0" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	resp, err := http.Get("http://" + addr + "/healthz")
	if err != nil {
		t.Fatalf("healthz: %v (addr=%s)", err, addr)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("healthz status = %d", resp.StatusCode)
	}
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("ListenAndServe: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ListenAndServe did not return after cancel")
	}
}

func TestUsageModePOSTReprojectsViews(t *testing.T) {
	srv := testServer(t, Options{Demo: true, UsageMode: "remaining"})
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var remaining Envelope
	if err := json.NewDecoder(w.Body).Decode(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining.UsageMode != "remaining" {
		t.Fatalf("usage_mode = %q, want remaining", remaining.UsageMode)
	}
	if len(remaining.Views) == 0 {
		t.Fatal("expected views")
	}
	remPct := remaining.Views[0].GaugePercent

	req = httptest.NewRequest(http.MethodPost, "/api/v1/usage-mode", strings.NewReader(`{"usage_mode":"used"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("usage-mode status = %d body %s", w.Code, w.Body.String())
	}
	var used Envelope
	if err := json.NewDecoder(w.Body).Decode(&used); err != nil {
		t.Fatal(err)
	}
	if used.UsageMode != "used" {
		t.Fatalf("usage_mode = %q, want used", used.UsageMode)
	}
	if len(used.Views) == 0 {
		t.Fatal("expected views after toggle")
	}
	if used.Views[0].AccountID != remaining.Views[0].AccountID {
		t.Fatalf("account changed %q → %q", remaining.Views[0].AccountID, used.Views[0].AccountID)
	}
	if used.Views[0].GaugePercent == remPct {
		t.Fatalf("gauge_percent stayed %v after switching to used", remPct)
	}
}

func TestUsageMode_CSRF_OriginProtection(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	handler := srv.Handler()

	// 1. Cross-site Sec-Fetch-Site should be blocked
	req := httptest.NewRequest(http.MethodPost, "/api/v1/usage-mode", strings.NewReader(`{"usage_mode":"used"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for Sec-Fetch-Site: cross-site, got %d", w.Code)
	}

	// 2. Untrusted external Origin should be blocked
	req = httptest.NewRequest(http.MethodPost, "/api/v1/usage-mode", strings.NewReader(`{"usage_mode":"used"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://evil.com")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for Origin: http://evil.com, got %d", w.Code)
	}

	// 3. Localhost origin should be allowed
	req = httptest.NewRequest(http.MethodPost, "/api/v1/usage-mode", strings.NewReader(`{"usage_mode":"used"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:8080")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for Origin: http://localhost:8080, got %d", w.Code)
	}

	// 4. 127.0.0.1 origin should be allowed
	req = httptest.NewRequest(http.MethodPost, "/api/v1/usage-mode", strings.NewReader(`{"usage_mode":"remaining"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://127.0.0.1:8080")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for Origin: http://127.0.0.1:8080, got %d", w.Code)
	}
}

func TestSnapshotsWrongMethod(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", w.Code)
	}
}

func TestSnapshotsRefreshWithAccountIDQueryParam(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots?refresh=1&account_id=test-acc", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestFetchingUXAssetsContract(t *testing.T) {
	srv := testServer(t, Options{Demo: true})

	// Test index.html structure
	{
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("GET / status = %d", w.Code)
		}
		html := w.Body.String()
		// #fetching-header inside .header-main
		if !strings.Contains(html, `class="header-main"`) || !strings.Contains(html, `id="fetching-header"`) {
			t.Error("index.html missing header-main or fetching-header")
		}
		headerMainIdx := strings.Index(html, `class="header-main"`)
		fetchingHeaderIdx := strings.Index(html, `id="fetching-header"`)
		if headerMainIdx == -1 || fetchingHeaderIdx == -1 || fetchingHeaderIdx < headerMainIdx {
			t.Errorf("expected fetching-header inside/after header-main, got headerMainIdx=%d, fetchingHeaderIdx=%d", headerMainIdx, fetchingHeaderIdx)
		}

		// #fetching-footer inside .footer-main
		footerMainIdx := strings.Index(html, `class="footer-main"`)
		fetchingFooterIdx := strings.Index(html, `id="fetching-footer"`)
		if footerMainIdx == -1 || fetchingFooterIdx == -1 || fetchingFooterIdx < footerMainIdx {
			t.Errorf("expected fetching-footer inside/after footer-main, got footerMainIdx=%d, fetchingFooterIdx=%d", footerMainIdx, fetchingFooterIdx)
		}

		// #fetching-detail should be inside #panel or inline to avoid layout jump
		panelIdx := strings.Index(html, `id="panel"`)
		fetchingDetailIdx := strings.Index(html, `id="fetching-detail"`)
		if panelIdx == -1 || fetchingDetailIdx == -1 || fetchingDetailIdx < panelIdx {
			t.Errorf("expected fetching-detail inside/after panel, got panelIdx=%d, fetchingDetailIdx=%d", panelIdx, fetchingDetailIdx)
		}
		if strings.Contains(html, `<div id="fetching-detail" class="fetching" hidden><span class="spin" aria-hidden="true">⠋</span> Fetching...</div>`+"\n"+`        <div id="panel"></div>`) {
			t.Error("fetching-detail should not be placed above #panel causing layout shift")
		}
	}

	// Test app.css rules
	{
		req := httptest.NewRequest(http.MethodGet, "/app.css", nil)
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("GET /app.css status = %d", w.Code)
		}
		css := w.Body.String()
		if strings.Contains(css, ".shell.refreshing .footer-main { display: none; }") ||
			strings.Contains(css, ".shell.refreshing .footer-main{display:none}") ||
			strings.Contains(css, ".shell.refreshing .footer-main { display: none") {
			t.Error("app.css should NOT hide .footer-main when refreshing")
		}
		if !strings.Contains(css, ".item.refreshing") {
			t.Error("app.css should contain styles for .item.refreshing")
		}
	}

	// Test app.js rules
	{
		req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("GET /app.js status = %d", w.Code)
		}
		js := w.Body.String()
		if !strings.Contains(js, "Fetching (") {
			t.Error("app.js should format dynamic fetching message with account ID")
		}
		if !strings.Contains(js, "Fetching all...") {
			t.Error("app.js should format dynamic fetching message for all accounts")
		}
		if !strings.Contains(js, "account_id=") {
			t.Error("app.js should pass account_id in query string for focused refresh")
		}
		if !strings.Contains(js, "350") {
			t.Error("app.js should enforce minimum visual display time (350ms)")
		}
	}
}

func TestAppJSFetchingIndicatorsPreserved(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /app.js status = %d", w.Code)
	}
	js := w.Body.String()

	// 1. Assert renderHeader defines #fetching-header
	renderHeaderIdx := strings.Index(js, "function renderHeader()")
	if renderHeaderIdx == -1 {
		t.Fatal("renderHeader not found in app.js")
	}
	renderNavIdx := strings.Index(js, "function renderNav()")
	if renderNavIdx == -1 || renderNavIdx < renderHeaderIdx {
		t.Fatal("renderNav not found after renderHeader in app.js")
	}
	renderHeaderBody := js[renderHeaderIdx:renderNavIdx]
	if !strings.Contains(renderHeaderBody, `id="fetching-header"`) {
		t.Error("renderHeader must define id=\"fetching-header\"")
	}

	// 2. Assert renderDetail defines #fetching-detail in both empty view and normal view
	renderDetailIdx := strings.Index(js, "function renderDetail()")
	if renderDetailIdx == -1 {
		t.Fatal("renderDetail not found in app.js")
	}
	renderFooterIdx := strings.Index(js, "function renderFooter()")
	if renderFooterIdx == -1 || renderFooterIdx < renderDetailIdx {
		t.Fatal("renderFooter not found after renderDetail in app.js")
	}
	renderDetailBody := js[renderDetailIdx:renderFooterIdx]

	emptyViewIdx := strings.Index(renderDetailBody, "if (!views.length)")
	if emptyViewIdx == -1 {
		t.Fatal("renderDetail missing 'if (!views.length)' check")
	}
	emptyReturnIdx := strings.Index(renderDetailBody[emptyViewIdx:], "return;")
	if emptyReturnIdx == -1 {
		t.Fatal("renderDetail missing return in empty view branch")
	}
	emptyViewBranch := renderDetailBody[emptyViewIdx : emptyViewIdx+emptyReturnIdx]
	if !strings.Contains(emptyViewBranch, `id="fetching-detail"`) {
		t.Error("renderDetail empty view (!views.length) must include id=\"fetching-detail\"")
	}

	normalViewBranch := renderDetailBody[emptyViewIdx+emptyReturnIdx:]
	if !strings.Contains(normalViewBranch, `id="fetching-detail"`) {
		t.Error("renderDetail normal view must include id=\"fetching-detail\"")
	}

	// 3. Assert renderFooter defines #fetching-footer
	renderFuncIdx := strings.Index(js, "function render()")
	if renderFuncIdx == -1 || renderFuncIdx < renderFooterIdx {
		t.Fatal("render not found after renderFooter in app.js")
	}
	renderFooterBody := js[renderFooterIdx:renderFuncIdx]
	if !strings.Contains(renderFooterBody, `id="fetching-footer"`) {
		t.Error("renderFooter must define id=\"fetching-footer\"")
	}
}

func TestBasePathServesUnderPrefix(t *testing.T) {
	srv := testServer(t, Options{Demo: true, BasePath: "/agentusage"})
	handler := srv.Handler()

	root := httptest.NewRequest(http.MethodGet, "/", nil)
	rootW := httptest.NewRecorder()
	handler.ServeHTTP(rootW, root)
	if rootW.Code != http.StatusNotFound {
		t.Fatalf("GET / status = %d, want 404 when base path is set", rootW.Code)
	}

	redir := httptest.NewRequest(http.MethodGet, "/agentusage", nil)
	redirW := httptest.NewRecorder()
	handler.ServeHTTP(redirW, redir)
	if redirW.Code != http.StatusMovedPermanently {
		t.Fatalf("GET /agentusage status = %d, want 301", redirW.Code)
	}
	if loc := redirW.Header().Get("Location"); loc != "/agentusage/" {
		t.Fatalf("Location = %q, want /agentusage/", loc)
	}

	for _, path := range []string{
		"/agentusage/",
		"/agentusage/healthz",
		"/agentusage/app.js",
		"/agentusage/app.css",
		"/agentusage/api/v1/snapshots",
		"/agentusage/api/v1/meta",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("%s status = %d, want 200", path, w.Code)
		}
	}

	idx := httptest.NewRequest(http.MethodGet, "/agentusage/", nil)
	idxW := httptest.NewRecorder()
	handler.ServeHTTP(idxW, idx)
	html := idxW.Body.String()
	if strings.Contains(html, `href="/app.css"`) || strings.Contains(html, `src="/app.js"`) {
		t.Error("index.html must not use root-absolute /app.css or /app.js (breaks Tailscale Serve subpaths)")
	}
	if !strings.Contains(html, `href="app.css"`) || !strings.Contains(html, `src="app.js"`) {
		t.Error("index.html should load app.css and app.js with relative URLs")
	}

	jsReq := httptest.NewRequest(http.MethodGet, "/agentusage/app.js", nil)
	jsW := httptest.NewRecorder()
	handler.ServeHTTP(jsW, jsReq)
	js := jsW.Body.String()
	if strings.Contains(js, `fetch("/api/v1/snapshots`) || strings.Contains(js, `fetch("/api/v1/usage-mode"`) {
		t.Error("app.js must not fetch root-absolute /api/v1/* (breaks Tailscale Serve subpaths)")
	}
	if !strings.Contains(js, `fetch("api/v1/snapshots`) || !strings.Contains(js, `fetch("api/v1/usage-mode"`) {
		t.Error("app.js should fetch api/v1/* with relative URLs")
	}
}

func TestNewServerRejectsInvalidBasePath(t *testing.T) {
	_, err := NewServer(Options{Demo: true, ListenAddr: "127.0.0.1:0", BasePath: "/../secret"})
	if err == nil {
		t.Fatal("expected invalid base path to fail")
	}
}
