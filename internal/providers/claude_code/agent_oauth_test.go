package claude_code

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

// useAgentOAuthFixture points the sign-in store at a fresh temp file and the
// token endpoint at handler.
func useAgentOAuthFixture(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "credentials.json")
	oldPath, oldURL := agentOAuthCredsPath, claudeOAuthTokenURL
	agentOAuthCredsPath = func() string { return path }
	t.Cleanup(func() { agentOAuthCredsPath, claudeOAuthTokenURL = oldPath, oldURL })
	if handler != nil {
		srv := httptest.NewServer(handler)
		t.Cleanup(srv.Close)
		claudeOAuthTokenURL = srv.URL
	}
	return path
}

func saveAgentToken(t *testing.T, path string, tok config.OAuthToken) {
	t.Helper()
	if err := config.SaveOAuthTokenTo(path, AgentOAuthAccountID, tok); err != nil {
		t.Fatalf("SaveOAuthTokenTo: %v", err)
	}
}

func TestParsePastedAuthCode(t *testing.T) {
	tests := []struct {
		in, code, state string
		wantErr         bool
	}{
		{in: "abc#xyz", code: "abc", state: "xyz"},
		{in: "  abc#xyz\n", code: "abc", state: "xyz"},
		{in: "abc", code: "abc"},
		{in: "https://platform.claude.com/oauth/code/callback?code=abc&state=xyz", code: "abc", state: "xyz"},
		{in: "", wantErr: true},
		{in: "#xyz", wantErr: true},
	}
	for _, tt := range tests {
		code, state, err := parsePastedAuthCode(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("parsePastedAuthCode(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			continue
		}
		if code != tt.code || state != tt.state {
			t.Errorf("parsePastedAuthCode(%q) = (%q, %q), want (%q, %q)", tt.in, code, state, tt.code, tt.state)
		}
	}
}

func TestClaudeLogin_AuthorizeURL(t *testing.T) {
	l, err := NewClaudeLogin()
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(l.AuthorizeURL)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("client_id") != claudeOAuthClientID || q.Get("scope") != "user:profile" ||
		q.Get("redirect_uri") != claudeOAuthRedirectURI || q.Get("code_challenge_method") != "S256" ||
		q.Get("state") != l.state || q.Get("code_challenge") == "" {
		t.Errorf("unexpected authorize query: %v", q)
	}
}

func TestClaudeLogin_CompleteStoresToken(t *testing.T) {
	var got map[string]string
	path := useAgentOAuthFixture(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"access_token":"acc-1","refresh_token":"ref-1","expires_in":3600,"scope":"user:profile"}`))
	})
	l, err := NewClaudeLogin()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := l.Complete(context.Background(), "code-1#wrong-state"); err == nil {
		t.Fatal("Complete with mismatched state: want error")
	}

	tok, err := l.Complete(context.Background(), "code-1#"+l.state)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if got["grant_type"] != "authorization_code" || got["code"] != "code-1" || got["code_verifier"] != l.verifier {
		t.Errorf("token request = %v", got)
	}
	stored, ok, _ := config.LoadOAuthTokenFrom(path, AgentOAuthAccountID)
	if !ok || stored.AccessToken != "acc-1" || stored.RefreshToken != "ref-1" || stored != tok {
		t.Errorf("stored = %+v, ok=%v", stored, ok)
	}
}

func TestAgentOAuthAccessToken_NotSignedIn(t *testing.T) {
	useAgentOAuthFixture(t, nil)
	if _, err := agentOAuthAccessToken(context.Background(), false); !errors.Is(err, ErrAgentOAuthNotSignedIn) {
		t.Fatalf("err = %v, want ErrAgentOAuthNotSignedIn", err)
	}
}

func TestAgentOAuthAccessToken_FreshTokenNotRenewed(t *testing.T) {
	var calls int32
	path := useAgentOAuthFixture(t, func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
	})
	saveAgentToken(t, path, config.OAuthToken{
		AccessToken: "acc", RefreshToken: "ref",
		ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	})
	tok, err := agentOAuthAccessToken(context.Background(), false)
	if err != nil || tok != "acc" {
		t.Fatalf("token = %q, err = %v", tok, err)
	}
	if calls != 0 {
		t.Errorf("token endpoint called %d times, want 0", calls)
	}
}

func TestAgentOAuthAccessToken_RenewsExpiredAndPersists(t *testing.T) {
	var got map[string]string
	path := useAgentOAuthFixture(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"access_token":"acc-2","refresh_token":"ref-2","expires_in":28800}`))
	})
	saveAgentToken(t, path, config.OAuthToken{
		AccessToken: "acc-1", RefreshToken: "ref-1", Scope: "user:profile",
		ExpiresAt: time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
	})

	tok, err := agentOAuthAccessToken(context.Background(), false)
	if err != nil || tok != "acc-2" {
		t.Fatalf("token = %q, err = %v", tok, err)
	}
	if got["grant_type"] != "refresh_token" || got["refresh_token"] != "ref-1" || got["client_id"] != claudeOAuthClientID {
		t.Errorf("refresh request = %v", got)
	}
	stored, _, _ := config.LoadOAuthTokenFrom(path, AgentOAuthAccountID)
	if stored.AccessToken != "acc-2" || stored.RefreshToken != "ref-2" || stored.Scope != "user:profile" {
		t.Errorf("stored = %+v", stored)
	}
	if agentOAuthNeedsRefresh(stored, time.Now()) {
		t.Errorf("renewed token already needs refresh: %+v", stored)
	}
}

func TestAgentOAuthAccessToken_InvalidGrantIsSignInExpired(t *testing.T) {
	path := useAgentOAuthFixture(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"Refresh token not found or invalid"}`))
	})
	saveAgentToken(t, path, config.OAuthToken{
		AccessToken: "acc", RefreshToken: "ref",
		ExpiresAt: time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
	})
	if _, err := agentOAuthAccessToken(context.Background(), false); !errors.Is(err, ErrAgentOAuthSignInExpired) {
		t.Fatalf("err = %v, want ErrAgentOAuthSignInExpired", err)
	}
}

func TestAgentOAuthAccessToken_OutageIsNotSignInExpired(t *testing.T) {
	path := useAgentOAuthFixture(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	saveAgentToken(t, path, config.OAuthToken{
		AccessToken: "acc", RefreshToken: "ref",
		ExpiresAt: time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
	})
	_, err := agentOAuthAccessToken(context.Background(), false)
	if err == nil || errors.Is(err, ErrAgentOAuthSignInExpired) {
		t.Fatalf("err = %v, want a transient error", err)
	}
}

// TestReadUsageAPI_AgentOAuthRenewsOn401 covers a token the API refuses
// before its recorded expiry: the source renews once and retries.
func TestReadUsageAPI_AgentOAuthRenewsOn401(t *testing.T) {
	path := useAgentOAuthFixture(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"acc-new","refresh_token":"ref-new","expires_in":28800}`))
	})
	saveAgentToken(t, path, config.OAuthToken{
		AccessToken: "acc-old", RefreshToken: "ref-old",
		ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	})

	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		seen = append(seen, auth)
		if auth != "Bearer acc-new" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"five_hour":{"utilization":42,"resets_at":"2999-01-01T00:00:00Z"}}`))
	}))
	defer srv.Close()
	old := oauthUsageURL
	oauthUsageURL = srv.URL
	defer func() { oauthUsageURL = old }()
	setTempHome(t) // no Claude Code credentials to fall back to

	p := &Provider{}
	snap := core.NewUsageSnapshot("claude_code", "acct")
	if err := p.readUsageAPI(context.Background(), "org-uuid", &snap); err != nil {
		t.Fatalf("readUsageAPI: %v", err)
	}
	if strings.Join(seen, ",") != "Bearer acc-old,Bearer acc-new" {
		t.Errorf("usage requests = %v", seen)
	}
	if snap.Raw["usage_api_source"] != "agent-oauth" {
		t.Errorf("usage_api_source = %q, want agent-oauth", snap.Raw["usage_api_source"])
	}
	if m := snap.Metrics["usage_five_hour"]; m.Used == nil || *m.Used != 42 {
		t.Errorf("usage_five_hour = %+v", m)
	}
}

func TestAgentOAuthAccessToken_CancelledContextStillSaves(t *testing.T) {
	path := useAgentOAuthFixture(t, func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(30 * time.Millisecond)
		_, _ = w.Write([]byte(`{"access_token":"acc-2","refresh_token":"ref-2","expires_in":28800}`))
	})
	saveAgentToken(t, path, config.OAuthToken{
		AccessToken: "acc-1", RefreshToken: "ref-1",
		ExpiresAt: time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tok, err := agentOAuthAccessToken(ctx, false)
	if err != nil || tok != "acc-2" {
		t.Fatalf("token = %q, err = %v (cancelled Fetch must not abort persist)", tok, err)
	}
	stored, _, _ := config.LoadOAuthTokenFrom(path, AgentOAuthAccountID)
	if stored.RefreshToken != "ref-2" {
		t.Errorf("stored refresh = %q, want ref-2", stored.RefreshToken)
	}
}

func TestAgentOAuthAccessToken_ConcurrentRefreshOnce(t *testing.T) {
	var calls int32
	path := useAgentOAuthFixture(t, func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(40 * time.Millisecond)
		_, _ = w.Write([]byte(`{"access_token":"acc-2","refresh_token":"ref-2","expires_in":28800}`))
	})
	saveAgentToken(t, path, config.OAuthToken{
		AccessToken: "acc-1", RefreshToken: "ref-1",
		ExpiresAt: time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
	})

	var wg sync.WaitGroup
	errCh := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tok, err := agentOAuthAccessToken(context.Background(), false)
			if err != nil {
				errCh <- err
				return
			}
			if tok != "acc-2" {
				errCh <- fmtError("token", tok)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("token endpoint called %d times, want 1", got)
	}
}

func fmtError(label, got string) error {
	return errors.New(label + " = " + got)
}

func TestProvider_Fetch_AgentOAuthWithoutOrgUUID(t *testing.T) {
	path := useAgentOAuthFixture(t, nil)
	saveAgentToken(t, path, config.OAuthToken{
		AccessToken: "acc", RefreshToken: "ref",
		ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer acc" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"five_hour":{"utilization":7,"resets_at":"2999-01-01T00:00:00Z"}}`))
	}))
	defer srv.Close()
	old := oauthUsageURL
	oauthUsageURL = srv.URL
	defer func() { oauthUsageURL = old }()
	setTempHome(t)

	tmpDir := t.TempDir()
	accountPath := filepath.Join(tmpDir, ".claude.json")
	if err := os.WriteFile(accountPath, []byte(`{"hasAvailableSubscription":true,"oauthAccount":{"emailAddress":"a@b.c"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	p := New()
	snap, err := p.Fetch(context.Background(), testClaudeAccount("test-claude", filepath.Join(tmpDir, "no-stats"), accountPath))
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if snap.Raw["usage_api_source"] != "agent-oauth" {
		t.Errorf("usage_api_source = %q, want agent-oauth (no org UUID); raw=%v", snap.Raw["usage_api_source"], snap.Raw)
	}
	if m := snap.Metrics["usage_five_hour"]; m.Used == nil || *m.Used != 7 {
		t.Errorf("usage_five_hour = %+v", m)
	}
}
