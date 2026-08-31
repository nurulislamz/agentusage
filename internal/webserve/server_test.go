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
	if !strings.Contains(html, "/app.js") {
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
