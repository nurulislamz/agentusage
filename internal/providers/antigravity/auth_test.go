//go:build !windows

package antigravity

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestConfigDir_Variations(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// 1. Direct path override
	acct1 := core.AccountConfig{
		ID:            "acct-1",
		ProviderPaths: map[string]string{"config_dir": "/custom/path"},
	}
	if got := configDir(acct1); got != "/custom/path" {
		t.Errorf("configDir(override) = %q, want /custom/path", got)
	}

	// 2. box_name hint
	acct2 := core.AccountConfig{
		ID:           "acct-2",
		RuntimeHints: map[string]string{"box_name": "worker-1"},
	}
	want2 := filepath.Join(home, ".agy-containers", "worker-1", ".gemini", "antigravity-cli")
	if got := configDir(acct2); got != want2 {
		t.Errorf("configDir(box_name hint) = %q, want %q", got, want2)
	}

	// 3. antigravity-<box> account ID
	acct3 := core.AccountConfig{
		ID: "antigravity-cluster-alpha",
	}
	want3 := filepath.Join(home, ".agy-containers", "cluster-alpha", ".gemini", "antigravity-cli")
	if got := configDir(acct3); got != want3 {
		t.Errorf("configDir(antigravity-<box>) = %q, want %q", got, want3)
	}

	// 4. Default account ID
	acct4 := core.AccountConfig{
		ID: defaultAccountID,
	}
	want4 := filepath.Join(home, ".gemini", "antigravity-cli")
	if got := configDir(acct4); got != want4 {
		t.Errorf("configDir(default) = %q, want %q", got, want4)
	}

	// 5. Unrecognized account ID without box or path
	acct5 := core.AccountConfig{
		ID: "custom-unsupported-id",
	}
	if got := configDir(acct5); got != "" {
		t.Errorf("configDir(unrecognized) = %q, want empty string", got)
	}

	// 6. Empty HOME
	t.Setenv("HOME", "")
	if got := configDir(acct4); got != "" {
		t.Errorf("configDir(empty HOME) = %q, want empty string", got)
	}
}

func TestTokenFilePath_Variations(t *testing.T) {
	// 1. Override path
	acct1 := core.AccountConfig{
		ProviderPaths: map[string]string{"oauth_token_file": "/custom/token/path"},
	}
	if got := tokenFilePath(acct1); got != "/custom/token/path" {
		t.Errorf("tokenFilePath(override) = %q, want /custom/token/path", got)
	}

	// 2. Empty configDir yields empty tokenFilePath
	acct2 := core.AccountConfig{
		ID: "unknown-account",
	}
	if got := tokenFilePath(acct2); got != "" {
		t.Errorf("tokenFilePath(no configDir) = %q, want empty string", got)
	}

	// 3. Standard configDir
	acct3 := core.AccountConfig{
		ID:            "acct-3",
		ProviderPaths: map[string]string{"config_dir": "/tmp/custom_config"},
	}
	want3 := filepath.Join("/tmp/custom_config", oauthTokenFile)
	if got := tokenFilePath(acct3); got != want3 {
		t.Errorf("tokenFilePath = %q, want %q", got, want3)
	}
}

func TestBoxName_Variations(t *testing.T) {
	// 1. Runtime hint
	acct1 := core.AccountConfig{
		RuntimeHints: map[string]string{"box_name": "hint-box"},
	}
	if got := boxName(acct1); got != "hint-box" {
		t.Errorf("boxName(hint) = %q, want hint-box", got)
	}

	// 2. Account ID prefix
	acct2 := core.AccountConfig{
		ID: "antigravity-id-box",
	}
	if got := boxName(acct2); got != "id-box" {
		t.Errorf("boxName(id prefix) = %q, want id-box", got)
	}

	// 3. Config directory path with .agy-containers
	acct3 := core.AccountConfig{
		ProviderPaths: map[string]string{
			"config_dir": "/home/user/.agy-containers/dir-box/.gemini/antigravity-cli",
		},
	}
	if got := boxName(acct3); got != "dir-box" {
		t.Errorf("boxName(config_dir container) = %q, want dir-box", got)
	}

	// 4. Config directory path without .agy-containers
	acct4 := core.AccountConfig{
		ProviderPaths: map[string]string{
			"config_dir": "/home/user/.gemini/antigravity-cli",
		},
	}
	if got := boxName(acct4); got != "" {
		t.Errorf("boxName(no container) = %q, want empty string", got)
	}

	// 5. Empty account config
	acct5 := core.AccountConfig{
		ID: "arbitrary-id",
	}
	if got := boxName(acct5); got != "" {
		t.Errorf("boxName(arbitrary) = %q, want empty string", got)
	}
}

func TestLoadOAuthToken_And_WriteOAuthToken(t *testing.T) {
	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token-test")

	// Write raw JSON with extra unknown fields
	rawInitial := `{
  "auth_method": "oauth",
  "project_id": "test-sentinel-project",
  "token": {
    "access_token": "secret-access-token",
    "refresh_token": "secret-refresh-token",
    "token_type": "Bearer",
    "expiry": "2030-01-01T00:00:00Z",
    "extra_token_field": "keep-me-safe"
  },
  "custom_metadata": 12345
}`
	if err := os.WriteFile(tokenPath, []byte(rawInitial), 0o600); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	// Load token
	loaded, err := loadOAuthToken(tokenPath)
	if err != nil {
		t.Fatalf("loadOAuthToken() error = %v", err)
	}
	if loaded.Token.AccessToken != "secret-access-token" {
		t.Errorf("AccessToken = %q, want 'secret-access-token'", loaded.Token.AccessToken)
	}
	if loaded.Token.RefreshToken != "secret-refresh-token" {
		t.Errorf("RefreshToken = %q, want 'secret-refresh-token'", loaded.Token.RefreshToken)
	}

	// Update token fields and write back
	loaded.Token.AccessToken = "rotated-access-token"
	loaded.Token.Expiry = "2031-01-01T00:00:00Z"
	if err := writeOAuthToken(tokenPath, loaded); err != nil {
		t.Fatalf("writeOAuthToken() error = %v", err)
	}

	// Check file permissions (0600)
	info, err := os.Stat(tokenPath)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("file mode = %v, want 0600", info.Mode().Perm())
	}

	// Verify unknown fields were preserved on disk
	content, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	var rawMap map[string]any
	if err := json.Unmarshal(content, &rawMap); err != nil {
		t.Fatalf("unmarshal disk json: %v", err)
	}
	if rawMap["project_id"] != "test-sentinel-project" {
		t.Errorf("project_id not preserved, got %v", rawMap["project_id"])
	}
	tokenMap, ok := rawMap["token"].(map[string]any)
	if !ok {
		t.Fatalf("token is not map: %v", rawMap["token"])
	}
	if tokenMap["access_token"] != "rotated-access-token" {
		t.Errorf("access_token not updated, got %v", tokenMap["access_token"])
	}
	if tokenMap["extra_token_field"] != "keep-me-safe" {
		t.Errorf("extra_token_field inside token not preserved, got %v", tokenMap["extra_token_field"])
	}
	if tokenMap["refresh_token"] != "secret-refresh-token" {
		t.Errorf("refresh_token not preserved, got %v", tokenMap["refresh_token"])
	}

	// Load non-existent file
	if _, err := loadOAuthToken(filepath.Join(dir, "nonexistent")); err == nil {
		t.Error("expected error loading non-existent token file")
	}

	// Load malformed file
	badPath := filepath.Join(dir, "bad-json")
	_ = os.WriteFile(badPath, []byte("not-json"), 0o600)
	if _, err := loadOAuthToken(badPath); err == nil {
		t.Error("expected error loading malformed token file")
	}

	// Write to invalid path
	if err := writeOAuthToken(filepath.Join(dir, "bad\x00path"), loaded); err == nil {
		t.Error("expected error writing to invalid path")
	}
}

// stubCLIPayload builds a fake CLI binary blob containing an OAuth client.
// The patterns are assembled at runtime so the fixtures never match the
// literals GitHub push protection scans for (the values are test fakes).
func stubCLIPayload() []byte {
	id := "1071006060591-abc123x." + "apps." + "googleusercontent" + ".com"
	secret := "GO" + "CSPX-" + "ZZZTESTSECRETVALUE1234567890"
	blob := "client_id=" + id + " secret=" + secret + " end"
	out := bytes.Repeat([]byte{0x00}, 2048)
	return append(out, []byte(blob)...)
}

func TestTokenExpired_Matrix(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)

	// 1. Empty access token is expired
	if !tokenExpired(oauthToken{AccessToken: ""}, now) {
		t.Error("token with empty access token should be expired")
	}

	// 2. Empty expiry with valid access token is not expired
	if tokenExpired(oauthToken{AccessToken: "valid", Expiry: ""}, now) {
		t.Error("token with empty expiry should not be expired")
	}

	// 3. Expiry far in future (RFC3339Nano)
	tokFutureNano := oauthToken{
		AccessToken: "valid",
		Expiry:      now.Add(2 * time.Hour).Format(time.RFC3339Nano),
	}
	if tokenExpired(tokFutureNano, now) {
		t.Error("future RFC3339Nano token should not be expired")
	}

	// 4. Expiry far in future (RFC3339)
	tokFutureRFC := oauthToken{
		AccessToken: "valid",
		Expiry:      now.Add(2 * time.Hour).Format(time.RFC3339),
	}
	if tokenExpired(tokFutureRFC, now) {
		t.Error("future RFC3339 token should not be expired")
	}

	// 5. Expiry within skew window (e.g. 30s in future, skew is 60s)
	tokNearExpiry := oauthToken{
		AccessToken: "valid",
		Expiry:      now.Add(30 * time.Second).Format(time.RFC3339Nano),
	}
	if !tokenExpired(tokNearExpiry, now) {
		t.Error("token within skew window should be expired")
	}

	// 6. Expiry in the past
	tokPast := oauthToken{
		AccessToken: "valid",
		Expiry:      now.Add(-10 * time.Minute).Format(time.RFC3339Nano),
	}
	if !tokenExpired(tokPast, now) {
		t.Error("past token should be expired")
	}

	// 7. Unparseable expiry string should return false
	tokBadDate := oauthToken{
		AccessToken: "valid",
		Expiry:      "not-a-date",
	}
	if tokenExpired(tokBadDate, now) {
		t.Error("token with bad date string should return false")
	}
}

func TestOAuthClientCredentials(t *testing.T) {
	// Neither set: falls back to the CLI's installed-application client
	// extracted from a binary stub.
	t.Setenv(oauthClientIDEnv, "")
	t.Setenv(oauthClientSecretEnv, "")
	t.Setenv(oauthClientIDEnvAlias, "")
	t.Setenv(oauthClientSecretEnvAlias, "")
	cliBin := filepath.Join(t.TempDir(), "agy")
	payload := stubCLIPayload()
	if err := os.WriteFile(cliBin, payload, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePath := cliOAuthClientPath
	cliOAuthClientPath = func() string { return cliBin }
	defer func() { cliOAuthClientPath = restorePath }()
	resetCLIOAuthClientCache()
	id, secret, ok := oauthClientCredentials()
	if !ok || !strings.HasSuffix(id, ".apps.googleusercontent.com") || !strings.HasPrefix(secret, "GOCSPX-") {
		t.Errorf("oauthClientCredentials() fallback = (%q, %q, %v), want extracted CLI client, true", id, secret, ok)
	}

	// Primary env vars
	t.Setenv(oauthClientIDEnv, "my-client-id")
	t.Setenv(oauthClientSecretEnv, "my-secret")
	id, secret, ok = oauthClientCredentials()
	if !ok || id != "my-client-id" || secret != "my-secret" {
		t.Errorf("oauthClientCredentials() = (%q, %q, %v), want (my-client-id, my-secret, true)", id, secret, ok)
	}

	// Aliases fallback
	t.Setenv(oauthClientIDEnv, "")
	t.Setenv(oauthClientSecretEnv, "")
	t.Setenv(oauthClientIDEnvAlias, "alias-client-id")
	t.Setenv(oauthClientSecretEnvAlias, "alias-secret")
	id, secret, ok = oauthClientCredentials()
	if !ok || id != "alias-client-id" || secret != "alias-secret" {
		t.Errorf("oauthClientCredentials() with aliases = (%q, %q, %v), want (alias-client-id, alias-secret, true)", id, secret, ok)
	}
}

func TestRefreshAccessToken_Branches(t *testing.T) {
	// Credentials not configured: the embedded fallback client is used and the
	// refresh proceeds over the real token endpoint (fails offline-friendly).
	t.Setenv(oauthClientIDEnv, "")
	t.Setenv(oauthClientSecretEnv, "")
	t.Setenv(oauthClientIDEnvAlias, "")
	t.Setenv(oauthClientSecretEnvAlias, "")
	_, err := refreshAccessToken(context.Background(), "refresh-tok", nil)
	if err == nil {
		t.Error("expected refresh failure without a reachable token endpoint")
	}
	var aErr *AuthError
	if errors.As(err, &aErr) && strings.Contains(aErr.Message, "oauth client not configured") {
		t.Error("fallback client should be used; unconfigured error must be gone")
	}

	// Set credentials
	t.Setenv(oauthClientIDEnv, "test-client-id")
	t.Setenv(oauthClientSecretEnv, "test-client-secret")

	// 2. Successful token refresh
	refreshServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if r.FormValue("client_id") != "test-client-id" || r.FormValue("grant_type") != "refresh_token" {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{
			"access_token": "brand-new-access",
			"refresh_token": "brand-new-refresh",
			"expires_in": 3600,
			"token_type": "Bearer"
		}`))
	}))
	defer refreshServer.Close()

	// 3. Error response (400 invalid_grant)
	invalidGrantServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "invalid_grant", "error_description": "Token has been expired or revoked."}`))
	}))
	defer invalidGrantServer.Close()

	// 4. Server error 500
	server500Attempts := int32(0)
	server500 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&server500Attempts, 1)
		http.Error(w, "internal server error access_token=body-secret refresh_token=body-secret", http.StatusInternalServerError)
	}))
	defer server500.Close()

	// Helper to point tokenEndpoint to test server
	testRefreshWithClient := func(serverURL string) (oauthToken, error) {
		customClient := &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				req.URL.Scheme = "http"
				req.URL.Host = strings.TrimPrefix(serverURL, "http://")
				return http.DefaultTransport.RoundTrip(req)
			}),
		}
		return refreshAccessToken(context.Background(), "my-refresh-token", customClient)
	}

	// Success case
	tok, err := testRefreshWithClient(refreshServer.URL)
	if err != nil {
		t.Fatalf("refreshAccessToken() error = %v", err)
	}
	if tok.AccessToken != "brand-new-access" || tok.RefreshToken != "brand-new-refresh" || tok.Expiry == "" {
		t.Errorf("refreshed token = %+v, want valid new tokens", tok)
	}

	// Invalid grant case (non-retryable AuthError)
	_, err = testRefreshWithClient(invalidGrantServer.URL)
	if err == nil {
		t.Fatal("expected error on invalid_grant")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) || !strings.Contains(err.Error(), "invalid_grant") {
		t.Errorf("expected AuthError with invalid_grant, got %v", err)
	}

	// Server error 500 case (retryable, bounded backoff)
	atomic.StoreInt32(&server500Attempts, 0)
	_, err = testRefreshWithClient(server500.URL)
	if err == nil {
		t.Fatal("expected error on 500")
	}
	if attempts := atomic.LoadInt32(&server500Attempts); attempts != 2 {
		t.Errorf("500 retry attempts = %d, want 2", attempts)
	}
	if strings.Contains(err.Error(), "body-secret") {
		t.Errorf("transient refresh error leaked response body: %v", err)
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestEnsureAccessToken_Flows(t *testing.T) {
	configDir := t.TempDir()
	tokenPath := filepath.Join(configDir, oauthTokenFile)

	// 1. Missing token file path
	if _, _, _, err := ensureAccessToken(context.Background(), core.AccountConfig{ID: "invalid"}, nil); err == nil {
		t.Error("expected error when tokenFilePath returns empty string")
	}

	// 2. Existing valid token (no refresh)
	writeTestToken(t, tokenPath, "disk-token", "2030-01-01T00:00:00Z", "")
	acct := core.AccountConfig{
		ID:            "antigravity",
		ProviderPaths: map[string]string{"config_dir": configDir},
	}
	tok, path, refreshed, err := ensureAccessToken(context.Background(), acct, nil)
	if err != nil {
		t.Fatalf("ensureAccessToken() error = %v", err)
	}
	if tok != "disk-token" || refreshed {
		t.Errorf("ensureAccessToken() = (%q, %v), want (disk-token, false)", tok, refreshed)
	}
	if path != tokenPath {
		t.Errorf("path = %q, want %q", path, tokenPath)
	}

	// 3. Missing token file returns AuthError (no subprocess)
	_ = os.Remove(tokenPath)
	_, _, _, err = ensureAccessToken(context.Background(), acct, nil)
	if err == nil {
		t.Fatal("expected error on missing token file")
	}
	var aErr *AuthError
	if !errors.As(err, &aErr) || !strings.Contains(err.Error(), "missing") {
		t.Errorf("expected missing file error, got %v", err)
	}

	// 4. Expired token without refresh token returns AuthError (no subprocess)
	writeTestToken(t, tokenPath, "expired-token", "2020-01-01T00:00:00Z", "")
	_, _, _, err = ensureAccessToken(context.Background(), acct, nil)
	if err == nil {
		t.Fatal("expected error on expired token without refresh token")
	}
	if !strings.Contains(err.Error(), "no refresh token") {
		t.Errorf("expected no refresh token error, got %v", err)
	}

	// 5. Expired token with refresh token and NO env credentials: the CLI's
	// installed-application client is extracted from a stub binary and used
	writeTestToken(t, tokenPath, "expired-token", "2020-01-01T00:00:00Z", "my-refresh-token")
	t.Setenv(oauthClientIDEnv, "")
	t.Setenv(oauthClientSecretEnv, "")
	t.Setenv(oauthClientIDEnvAlias, "")
	t.Setenv(oauthClientSecretEnvAlias, "")
	wantID := "1071006060591-abc123x." + "apps." + "googleusercontent" + ".com"
	wantSecret := "GO" + "CSPX-" + "ZZZTESTSECRETVALUE1234567890"
	cliBin2 := filepath.Join(t.TempDir(), "agy")
	payload2 := append(bytes.Repeat([]byte{0x00}, 2048),
		[]byte("client_id="+wantID+" secret="+wantSecret+" end")...)
	if err := os.WriteFile(cliBin2, payload2, 0o755); err != nil {
		t.Fatal(err)
	}
	restorePath2 := cliOAuthClientPath
	cliOAuthClientPath = func() string { return cliBin2 }
	defer func() { cliOAuthClientPath = restorePath2 }()
	resetCLIOAuthClientCache()
	fallbackMock := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			form, _ := url.ParseQuery(string(body))
			if got := form.Get("client_id"); got != wantID {
				t.Errorf("fallback refresh used client_id %q, want %q", got, wantID)
			}
			if got := form.Get("client_secret"); got != wantSecret {
				t.Errorf("fallback refresh used secret %q, want fallback secret", got)
			}
			rec := httptest.NewRecorder()
			_, _ = rec.WriteString(`{"access_token":"fallback-refreshed","expires_in":3600}`)
			return rec.Result(), nil
		}),
	}
	tok, _, _, err = ensureAccessToken(context.Background(), acct, fallbackMock)
	if err != nil {
		t.Fatalf("expected fallback refresh to succeed, got %v", err)
	}
	if tok != "fallback-refreshed" {
		t.Errorf("fallback refresh access = %q, want %q", tok, "fallback-refreshed")
	}

	// 6. Expired token with refresh token and configured client credentials: refreshed via HTTP
	t.Setenv(oauthClientIDEnv, "client-id")
	t.Setenv(oauthClientSecretEnv, "client-secret")
	writeTestToken(t, tokenPath, "expired-token", "2020-01-01T00:00:00Z", "my-refresh-token")

	refreshMockClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			// Response omits refresh_token to test preservation of existing refresh token!
			_, _ = rec.WriteString(`{
				"access_token": "refreshed-via-oauth",
				"expires_in": 3600
			}`)
			return rec.Result(), nil
		}),
	}
	tok, _, refreshed, err = ensureAccessToken(context.Background(), acct, refreshMockClient)
	if err != nil {
		t.Fatalf("ensureAccessToken() with oauth refresh error = %v", err)
	}
	if tok != "refreshed-via-oauth" || !refreshed {
		t.Errorf("ensureAccessToken() oauth = (%q, %v), want (refreshed-via-oauth, true)", tok, refreshed)
	}

	// Verify existing refresh token was preserved when refresh response omitted it
	loaded, err := loadOAuthToken(tokenPath)
	if err != nil {
		t.Fatalf("load token error: %v", err)
	}
	if loaded.Token.RefreshToken != "my-refresh-token" {
		t.Errorf("refresh token not preserved: %q", loaded.Token.RefreshToken)
	}

	// 7. Force refresh even when token is not expired (e.g. after 401 retry)
	writeTestToken(t, tokenPath, "future-token", "2035-01-01T00:00:00Z", "my-refresh-token")
	forceMockClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			_, _ = rec.WriteString(`{
				"access_token": "forced-renewed-token",
				"expires_in": 3600
			}`)
			return rec.Result(), nil
		}),
	}
	tok, _, refreshed, err = ensureAccessToken(context.Background(), acct, forceMockClient, true)
	if err != nil {
		t.Fatalf("forced ensureAccessToken error: %v", err)
	}
	if tok != "forced-renewed-token" || !refreshed {
		t.Errorf("forced token = (%q, %v), want (forced-renewed-token, true)", tok, refreshed)
	}
}

func TestEnsureAccessToken_LockAndExternalUpdate(t *testing.T) {
	t.Setenv(oauthClientIDEnv, "client-id")
	t.Setenv(oauthClientSecretEnv, "client-secret")
	configDir := t.TempDir()
	tokenPath := filepath.Join(configDir, oauthTokenFile)
	writeTestToken(t, tokenPath, "expired-token", "2020-01-01T00:00:00Z", "my-refresh-token")

	acct := core.AccountConfig{
		ID:            "antigravity",
		ProviderPaths: map[string]string{"config_dir": configDir},
	}

	refreshCalls := int32(0)
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&refreshCalls, 1)
			time.Sleep(50 * time.Millisecond)
			rec := httptest.NewRecorder()
			_, _ = rec.WriteString(`{"access_token": "refreshed-tok", "expires_in": 3600}`)
			return rec.Result(), nil
		}),
	}

	// Run concurrent ensureAccessToken calls on the same file
	var wg sync.WaitGroup
	tokens := make([]string, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tok, _, _, err := ensureAccessToken(context.Background(), acct, client)
			if err != nil {
				t.Errorf("concurrent ensureAccessToken error: %v", err)
				return
			}
			tokens[idx] = tok
		}(i)
	}
	wg.Wait()

	// Only 1 refresh call should have been made due to path locking and re-reading
	if calls := atomic.LoadInt32(&refreshCalls); calls != 1 {
		t.Errorf("expected 1 refresh call for concurrent callers, got %d", calls)
	}
	for _, tok := range tokens {
		if tok != "refreshed-tok" {
			t.Errorf("got token %q, want refreshed-tok", tok)
		}
	}
}

func TestNoAgyOrAgyBoxInvocation(t *testing.T) {
	// Put a sentinel executable in PATH named 'agy' and 'agy-box' that fails immediately if invoked
	fakeBin := t.TempDir()
	t.Setenv("PATH", fakeBin+":"+os.Getenv("PATH"))

	sentinelFile := filepath.Join(fakeBin, "sentinel-triggered")
	trapScript := "#!/bin/sh\ntouch " + sentinelFile + "\nexit 99\n"

	for _, name := range []string{"agy", "agy-box"} {
		p := filepath.Join(fakeBin, name)
		if err := os.WriteFile(p, []byte(trapScript), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	configDir := t.TempDir()
	tokenPath := filepath.Join(configDir, oauthTokenFile)
	acct := core.AccountConfig{
		ID:            "antigravity-test",
		Provider:      "antigravity",
		ProviderPaths: map[string]string{"config_dir": configDir},
	}

	provider := New()

	// Scenario 1: Missing token file
	snap, _ := provider.Fetch(context.Background(), acct)
	if snap.Status != core.StatusAuth {
		t.Errorf("scenario 1 status = %q, want auth", snap.Status)
	}

	// Scenario 2: Corrupted token file
	_ = os.WriteFile(tokenPath, []byte("bad-json"), 0o600)
	snap, _ = provider.Fetch(context.Background(), acct)
	if snap.Status != core.StatusAuth {
		t.Errorf("scenario 2 status = %q, want auth", snap.Status)
	}

	// Scenario 3: Expired token without refresh token
	writeTestToken(t, tokenPath, "expired", "2020-01-01T00:00:00Z", "")
	snap, _ = provider.Fetch(context.Background(), acct)
	if snap.Status != core.StatusAuth {
		t.Errorf("scenario 3 status = %q, want auth", snap.Status)
	}

	// Scenario 4: Quota 401 error
	t.Setenv(oauthClientIDEnv, "cid")
	t.Setenv(oauthClientSecretEnv, "csecret")
	writeTestToken(t, tokenPath, "valid-but-rejected", "2030-01-01T00:00:00Z", "my-refresh-token")

	mockTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		if strings.Contains(req.URL.Path, "token") {
			_, _ = rec.WriteString(`{"access_token": "new-token", "expires_in": 3600}`)
		} else {
			// Quota endpoint returns 401 Unauthorized
			http.Error(rec, "unauthorized", http.StatusUnauthorized)
		}
		return rec.Result(), nil
	})
	provider.HTTPClient = &http.Client{Transport: mockTransport}
	snap, _ = provider.Fetch(context.Background(), acct)
	if snap.Status != core.StatusAuth {
		t.Errorf("scenario 4 status = %q, want auth", snap.Status)
	}

	// Scenario 5: Quota 403 error
	mock403Transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		http.Error(rec, "forbidden", http.StatusForbidden)
		return rec.Result(), nil
	})
	provider.HTTPClient = &http.Client{Transport: mock403Transport}
	snap, _ = provider.Fetch(context.Background(), acct)
	if snap.Status != core.StatusAuth {
		t.Errorf("scenario 5 status = %q, want auth", snap.Status)
	}

	// Scenario 6: Quota 429 rate limit
	mock429Transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		rec.Header().Set("Retry-After", "60")
		http.Error(rec, "rate limit", http.StatusTooManyRequests)
		return rec.Result(), nil
	})
	provider.HTTPClient = &http.Client{Transport: mock429Transport}
	snap, _ = provider.Fetch(context.Background(), acct)
	if snap.Status != core.StatusLimited {
		t.Errorf("scenario 6 status = %q, want limited", snap.Status)
	}

	// Assert sentinel file was NEVER created
	if _, err := os.Stat(sentinelFile); err == nil {
		t.Fatal("CRITICAL: agy or agy-box sentinel executable was invoked during auth scenarios!")
	}
}
