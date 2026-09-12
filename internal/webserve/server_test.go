package webserve

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/providers/antigravity"
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

func TestStaticAssets(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	for _, path := range []string{"/app.css", "/app.js", "/htmx.min.js"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("%s status = %d", path, w.Code)
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

func TestThemePOSTReprojectsViews(t *testing.T) {
	srv := testServer(t, Options{Demo: true, Theme: "Gruvbox"})
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var initEnv Envelope
	if err := json.NewDecoder(w.Body).Decode(&initEnv); err != nil {
		t.Fatal(err)
	}
	if initEnv.Theme != "Gruvbox" {
		t.Fatalf("theme = %q, want Gruvbox", initEnv.Theme)
	}

	// Explicit theme switch
	req = httptest.NewRequest(http.MethodPost, "/api/v1/theme", strings.NewReader(`{"theme":"Nord"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("theme status = %d body %s", w.Code, w.Body.String())
	}
	var nordEnv Envelope
	if err := json.NewDecoder(w.Body).Decode(&nordEnv); err != nil {
		t.Fatal(err)
	}
	if nordEnv.Theme != "Nord" {
		t.Fatalf("theme = %q, want Nord", nordEnv.Theme)
	}
	if nordEnv.ThemeTokens.Name != "Nord" {
		t.Fatalf("theme_tokens.name = %q, want Nord", nordEnv.ThemeTokens.Name)
	}

	// Empty body cycles to next theme
	req = httptest.NewRequest(http.MethodPost, "/api/v1/theme", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("theme cycle status = %d body %s", w.Code, w.Body.String())
	}
	var cycledEnv Envelope
	if err := json.NewDecoder(w.Body).Decode(&cycledEnv); err != nil {
		t.Fatal(err)
	}
	if cycledEnv.Theme == "Nord" {
		t.Fatalf("expected theme to cycle from Nord, got %q", cycledEnv.Theme)
	}

	// Backward cycling with {"backward": true} returns back to Nord
	req = httptest.NewRequest(http.MethodPost, "/api/v1/theme", strings.NewReader(`{"backward":true}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("theme backward status = %d body %s", w.Code, w.Body.String())
	}
	var backEnv Envelope
	if err := json.NewDecoder(w.Body).Decode(&backEnv); err != nil {
		t.Fatal(err)
	}
	if backEnv.Theme != "Nord" {
		t.Fatalf("expected theme to cycle back to Nord, got %q", backEnv.Theme)
	}

	// Backward cycling with query param ?direction=prev
	req = httptest.NewRequest(http.MethodPost, "/api/v1/theme?direction=prev", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("theme backward query status = %d body %s", w.Code, w.Body.String())
	}
	var backQueryEnv Envelope
	if err := json.NewDecoder(w.Body).Decode(&backQueryEnv); err != nil {
		t.Fatal(err)
	}
	if backQueryEnv.Theme == "Nord" {
		t.Fatalf("expected theme to cycle back from Nord, got %q", backQueryEnv.Theme)
	}
}

func TestTheme_CSRF_OriginProtection(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	handler := srv.Handler()

	// Cross-site Sec-Fetch-Site should be blocked
	req := httptest.NewRequest(http.MethodPost, "/api/v1/theme", strings.NewReader(`{"theme":"Nord"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for Sec-Fetch-Site: cross-site, got %d", w.Code)
	}

	// External Origin should be blocked
	req = httptest.NewRequest(http.MethodPost, "/api/v1/theme", strings.NewReader(`{"theme":"Nord"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://evil.com")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for Origin: http://evil.com, got %d", w.Code)
	}

	// Localhost origin should be allowed
	req = httptest.NewRequest(http.MethodPost, "/api/v1/theme", strings.NewReader(`{"theme":"Nord"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:8080")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for Origin: http://localhost:8080, got %d", w.Code)
	}
}

func TestGetThemesEndpoint(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	for _, path := range []string{"/api/v1/theme", "/api/v1/themes"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want 200", path, w.Code)
		}
		var res struct {
			Active string   `json:"active"`
			Themes []string `json:"themes"`
		}
		if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
			t.Fatalf("%s json decode failed: %v", path, err)
		}
		if res.Active == "" {
			t.Fatalf("%s active theme is empty", path)
		}
		if len(res.Themes) < 10 {
			t.Fatalf("%s themes count = %d, want at least 10", path, len(res.Themes))
		}
	}
}

func TestEnvelopeIncludesAvailableThemes(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if len(env.AvailableThemes) < 10 {
		t.Fatalf("available_themes count = %d, want at least 10", len(env.AvailableThemes))
	}
}

func TestNoDataReportsErrorMessageAndNoFakeData(t *testing.T) {
	srv := testServer(t, Options{
		Demo: false,
		Collect: func() (Envelope, error) {
			return Envelope{
				Source:    "direct",
				Snapshots: []core.UsageSnapshot{},
			}, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if len(env.Snapshots) != 0 {
		t.Fatalf("expected 0 snapshots (no fake data), got %d", len(env.Snapshots))
	}
	if len(env.Views) != 0 {
		t.Fatalf("expected 0 views (no fake data), got %d", len(env.Views))
	}
	if env.Error == "" {
		t.Fatal("expected error message in envelope when no data is available")
	}
	if !strings.Contains(strings.ToLower(env.Error), "no usage snapshots") {
		t.Fatalf("expected descriptive error message, got %q", env.Error)
	}
}

func TestEmptyStateNoFakeDataOrDemoSuggestionContract(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("/ status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if strings.Contains(body, "agentusage serve --demo") {
		t.Error("index.html should not suggest running agentusage serve --demo when there is no data")
	}
	if !strings.Contains(body, `id="empty-error"`) {
		t.Error("index.html missing #empty-error container for reporting errors")
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
		"/agentusage/metrics",
		"/agentusage/app.js",
		"/agentusage/app.css",
		"/agentusage/htmx.min.js",
		"/agentusage/partial/app",
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
	if strings.Contains(html, `href="/app.css"`) || strings.Contains(html, `src="/app.js"`) || strings.Contains(html, `src="/htmx.min.js"`) {
		t.Error("shell must not use root-absolute asset URLs (breaks Tailscale Serve subpaths)")
	}
	if !strings.Contains(html, `href="app.css"`) || !strings.Contains(html, `src="app.js"`) || !strings.Contains(html, `src="htmx.min.js"`) {
		t.Error("shell should load app.css, app.js and htmx.min.js with relative URLs")
	}
	if !strings.Contains(html, `hx-get="partial/app"`) {
		t.Error("shell should request the app fragment with a relative URL")
	}

	jsReq := httptest.NewRequest(http.MethodGet, "/agentusage/app.js", nil)
	jsW := httptest.NewRecorder()
	handler.ServeHTTP(jsW, jsReq)
	js := jsW.Body.String()
	if strings.Contains(js, `"/api/v1/`) || strings.Contains(js, `"/partial/`) {
		t.Error("app.js must not use root-absolute API/fragment URLs")
	}
	if !strings.Contains(js, `htmx.ajax("GET", "partial/app"`) {
		t.Error("app.js should request the app fragment with a relative URL")
	}
}

func TestNewServerRejectsInvalidBasePath(t *testing.T) {
	_, err := NewServer(Options{Demo: true, ListenAddr: "127.0.0.1:0", BasePath: "/../secret"})
	if err == nil {
		t.Fatal("expected invalid base path to fail")
	}
}

func TestMetricsEndpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	antigravity.ResetMetricsForTesting()
	antigravity.RecordBoxPing("chaos", "antigravity-chaos", "missing_token", 120*time.Millisecond, nil)

	srv := testServer(t, Options{Demo: true})
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /metrics code = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "agy_box_pings_total") {
		t.Errorf("expected agy_box_pings_total in /metrics, got: %s", body)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain", ct)
	}
}
