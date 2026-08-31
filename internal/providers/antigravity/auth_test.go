//go:build !windows

package antigravity

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

	payload := oauthTokenFilePayload{
		AuthMethod: "oauth",
		Token: oauthToken{
			AccessToken:  "secret-access-token",
			RefreshToken: "secret-refresh-token",
			TokenType:    "Bearer",
			Expiry:       "2030-01-01T00:00:00Z",
		},
	}

	// Write token atomically
	if err := writeOAuthToken(tokenPath, payload); err != nil {
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

	// Write to non-existent parent directory
	if err := writeOAuthToken(filepath.Join(dir, "no-such-dir", "token"), payload); err == nil {
		t.Error("expected error writing token to non-existent directory")
	}
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

	// 7. Unparseable expiry string should return false (optimistic non-expiration)
	tokBadDate := oauthToken{
		AccessToken: "valid",
		Expiry:      "not-a-date",
	}
	if tokenExpired(tokBadDate, now) {
		t.Error("token with bad date string should return false")
	}
}

func TestOAuthClientCredentials(t *testing.T) {
	// Neither set
	t.Setenv(oauthClientIDEnv, "")
	t.Setenv(oauthClientSecretEnv, "")
	if _, _, ok := oauthClientCredentials(); ok {
		t.Error("expected false when envs unset")
	}

	// Only client ID
	t.Setenv(oauthClientIDEnv, "my-client-id")
	t.Setenv(oauthClientSecretEnv, "")
	if _, _, ok := oauthClientCredentials(); ok {
		t.Error("expected false when secret unset")
	}

	// Only client Secret
	t.Setenv(oauthClientIDEnv, "")
	t.Setenv(oauthClientSecretEnv, "my-secret")
	if _, _, ok := oauthClientCredentials(); ok {
		t.Error("expected false when client ID unset")
	}

	// Both set
	t.Setenv(oauthClientIDEnv, "my-client-id")
	t.Setenv(oauthClientSecretEnv, "my-secret")
	id, secret, ok := oauthClientCredentials()
	if !ok || id != "my-client-id" || secret != "my-secret" {
		t.Errorf("oauthClientCredentials() = (%q, %q, %v), want (my-client-id, my-secret, true)", id, secret, ok)
	}
}

func TestRefreshAccessToken_Branches(t *testing.T) {
	// 1. Credentials not configured
	t.Setenv(oauthClientIDEnv, "")
	t.Setenv(oauthClientSecretEnv, "")
	if _, err := refreshAccessToken(context.Background(), "refresh-tok", nil); err == nil {
		t.Error("expected error when credentials not configured")
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

	// 3. Error response from server
	errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid grant", http.StatusBadRequest)
	}))
	defer errServer.Close()

	// 4. Malformed JSON response
	badJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{invalid-json`))
	}))
	defer badJSONServer.Close()

	// 5. Empty access token response
	emptyTokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token": "", "expires_in": 3600}`))
	}))
	defer emptyTokenServer.Close()

	// Test helper that redirects tokenEndpoint
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

	// HTTP error case
	_, err = testRefreshWithClient(errServer.URL)
	if err == nil || !strings.Contains(err.Error(), "token refresh HTTP 400") {
		t.Errorf("expected HTTP 400 error, got %v", err)
	}

	// Bad JSON case
	_, err = testRefreshWithClient(badJSONServer.URL)
	if err == nil || !strings.Contains(err.Error(), "parse token refresh response") {
		t.Errorf("expected parse error, got %v", err)
	}

	// Empty access token case
	_, err = testRefreshWithClient(emptyTokenServer.URL)
	if err == nil || !strings.Contains(err.Error(), "empty access_token") {
		t.Errorf("expected empty access_token error, got %v", err)
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestEnsureAccessToken_Flows(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := t.TempDir()
	tokenPath := filepath.Join(configDir, oauthTokenFile)

	// Mock agy CLI
	binDir := filepath.Join(home, ".local", "bin")
	_ = os.MkdirAll(binDir, 0o755)
	mockAgy := filepath.Join(binDir, "agy")
	mockScript := fmt.Sprintf("#!/bin/sh\ncat << 'EOF' > %q\n{\n  \"token\": {\n    \"access_token\": \"ping-access-token\",\n    \"expiry\": \"2030-01-01T00:00:00Z\"\n  }\n}\nEOF\nexit 0\n", tokenPath)
	if err := os.WriteFile(mockAgy, []byte(mockScript), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

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

	// 3. Missing token file: triggers pingBoxForToken and reloads
	_ = os.Remove(tokenPath)
	tok, _, refreshed, err = ensureAccessToken(context.Background(), acct, nil)
	if err != nil {
		t.Fatalf("ensureAccessToken() after ping error = %v", err)
	}
	if tok != "ping-access-token" || !refreshed {
		t.Errorf("ensureAccessToken() = (%q, %v), want (ping-access-token, true)", tok, refreshed)
	}

	// 4. Missing token file with ping failure
	_ = os.Remove(tokenPath)
	acctFail := core.AccountConfig{
		ID:            "antigravity",
		Binary:        "/no/such/agy",
		ProviderPaths: map[string]string{"config_dir": configDir},
	}
	if _, _, _, err := ensureAccessToken(context.Background(), acctFail, nil); err == nil {
		t.Error("expected error when ping fails on missing token file")
	}

	// 5. Expired token without refresh token: triggers pingBoxForToken
	writeTestToken(t, tokenPath, "expired-token", "2020-01-01T00:00:00Z", "")
	tok, _, refreshed, err = ensureAccessToken(context.Background(), acct, nil)
	if err != nil {
		t.Fatalf("ensureAccessToken() expired error = %v", err)
	}
	if tok != "ping-access-token" || !refreshed {
		t.Errorf("ensureAccessToken() expired = (%q, %v), want (ping-access-token, true)", tok, refreshed)
	}

	// 6. Expired token with refresh token and OAuth credentials
	t.Setenv(oauthClientIDEnv, "client-id")
	t.Setenv(oauthClientSecretEnv, "client-secret")
	writeTestToken(t, tokenPath, "expired-token", "2020-01-01T00:00:00Z", "my-refresh-token")

	refreshMockClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			_, _ = rec.WriteString(`{
				"access_token": "refreshed-via-oauth",
				"refresh_token": "new-refresh-token",
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

	// 7. Expired token with refresh token where OAuth refresh fails -> falls back to ping
	writeTestToken(t, tokenPath, "expired-token", "2020-01-01T00:00:00Z", "my-refresh-token")
	refreshFailClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			http.Error(rec, "invalid_grant", http.StatusBadRequest)
			return rec.Result(), nil
		}),
	}
	tok, _, refreshed, err = ensureAccessToken(context.Background(), acct, refreshFailClient)
	if err != nil {
		t.Fatalf("ensureAccessToken() oauth fail fallback error = %v", err)
	}
	if tok != "ping-access-token" || !refreshed {
		t.Errorf("ensureAccessToken() oauth fail fallback = (%q, %v), want (ping-access-token, true)", tok, refreshed)
	}

	// 8. Expired token where ping fails with refresh token
	writeTestToken(t, tokenPath, "expired-token", "2020-01-01T00:00:00Z", "my-refresh-token")
	if _, _, _, err := ensureAccessToken(context.Background(), acctFail, refreshFailClient); err == nil {
		t.Error("expected error when both refresh and ping fail")
	}

	// 9. Existing token file is unreadable/corrupted (non-exist error)
	_ = os.WriteFile(tokenPath, []byte("broken json"), 0o600)
	if _, _, _, err := ensureAccessToken(context.Background(), acct, nil); err == nil {
		t.Error("expected error when existing token file is corrupted")
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("short", 10); got != "short" {
		t.Errorf("truncate(short, 10) = %q, want 'short'", got)
	}
	if got := truncate("exact", 5); got != "exact" {
		t.Errorf("truncate(exact, 5) = %q, want 'exact'", got)
	}
	if got := truncate("longer-string", 6); got != "longer…" {
		t.Errorf("truncate(longer-string, 6) = %q, want 'longer…'", got)
	}
}

func TestResolveCLI_Branches(t *testing.T) {
	// Empty name returns empty string
	if got := resolveCLI(""); got != "" {
		t.Errorf("resolveCLI('') = %q, want empty string", got)
	}

	// Absolute path returns as-is
	if got := resolveCLI("/usr/bin/agy"); got != "/usr/bin/agy" {
		t.Errorf("resolveCLI(abs) = %q, want /usr/bin/agy", got)
	}

	// Non-existent command returns name as-is
	if got := resolveCLI("definitely-nonexistent-command-xyz"); got != "definitely-nonexistent-command-xyz" {
		t.Errorf("resolveCLI(nonexistent) = %q, want name", got)
	}
}

func TestResolveCLI_FindsHomeLocalBinWhenNotOnPATH(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", t.TempDir())

	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(binDir, "agy-box")
	if err := os.WriteFile(want, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := resolveCLI("agy-box")
	if got != want {
		t.Fatalf("resolveCLI() = %q, want %q", got, want)
	}
}

func TestPingBoxForToken_DefaultAgy(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	binDir := filepath.Join(home, ".local", "bin")
	_ = os.MkdirAll(binDir, 0o755)

	script := filepath.Join(binDir, "agy")
	contents := "#!/bin/sh\n" +
		"test \"$1\" = -p || exit 2\n" +
		"test \"$2\" = ping || exit 3\n" +
		"exit 0\n"
	if err := os.WriteFile(script, []byte(contents), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	err := pingBoxForToken(context.Background(), core.AccountConfig{
		ID: "antigravity",
	})
	if err != nil {
		t.Fatalf("pingBoxForToken() default agy = %v", err)
	}
}

func TestPingBoxForToken_FindsAgyBoxOffPATH(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", filepath.Join(t.TempDir(), "missing"))

	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(binDir, "agy-box")
	contents := "#!/bin/sh\n" +
		"test \"$1\" = chaos || exit 2\n" +
		"test \"$2\" = -p || exit 3\n" +
		"test \"$3\" = ping || exit 4\n" +
		"exit 0\n"
	if err := os.WriteFile(script, []byte(contents), 0o755); err != nil {
		t.Fatal(err)
	}

	err := pingBoxForToken(context.Background(), core.AccountConfig{
		ID: "antigravity-chaos",
		RuntimeHints: map[string]string{
			"box_name": "chaos",
		},
	})
	if err != nil {
		t.Fatalf("pingBoxForToken() = %v", err)
	}
}

func TestPingBoxForToken_MissingBinaryIsError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", filepath.Join(t.TempDir(), "missing"))

	err := pingBoxForToken(context.Background(), core.AccountConfig{
		ID: "antigravity-chaos",
		RuntimeHints: map[string]string{
			"box_name": "chaos",
		},
	})
	if err == nil {
		t.Fatal("expected ping error when agy-box is missing")
	}
}

func TestAuth_ConcurrencyUnderRace(t *testing.T) {
	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token-concurrent")
	payload := oauthTokenFilePayload{
		Token: oauthToken{
			AccessToken:  "concurrent-token",
			RefreshToken: "concurrent-refresh",
			TokenType:    "Bearer",
			Expiry:       "2030-01-01T00:00:00Z",
		},
	}
	if err := writeOAuthToken(tokenPath, payload); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			loaded, err := loadOAuthToken(tokenPath)
			if err != nil {
				t.Errorf("loadOAuthToken() concurrent error = %v", err)
				return
			}
			if loaded.Token.AccessToken != "concurrent-token" {
				t.Errorf("loaded token = %q", loaded.Token.AccessToken)
			}
			if tokenExpired(loaded.Token, time.Now().UTC()) {
				t.Error("expected non-expired token")
			}
		}(i)
	}
	wg.Wait()
}
