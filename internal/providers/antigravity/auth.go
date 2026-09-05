package antigravity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

const (
	tokenEndpoint   = "https://oauth2.googleapis.com/token"
	oauthTokenFile  = "antigravity-oauth-token"
	tokenExpirySkew = 60 * time.Second

	oauthClientIDEnv          = "OPENUSAGE_ANTIGRAVITY_CLIENT_ID"
	oauthClientSecretEnv      = "OPENUSAGE_ANTIGRAVITY_CLIENT_SECRET"
	oauthClientIDEnvAlias     = "ANTIGRAVITY_CLIENT_ID"
	oauthClientSecretEnvAlias = "ANTIGRAVITY_CLIENT_SECRET"
)

// AuthError represents an authentication condition requiring user action (non-retryable).
type AuthError struct {
	Reason     string
	StatusCode int
	Message    string
}

func (e *AuthError) Error() string {
	return e.Message
}

// RefreshTransientError represents a temporary, retryable error during token renewal.
type RefreshTransientError struct {
	StatusCode int
	Message    string
}

func (e *RefreshTransientError) Error() string {
	return e.Message
}

type oauthTokenFilePayload struct {
	rawTop   map[string]json.RawMessage
	rawToken map[string]json.RawMessage
	Token    oauthToken `json:"token"`
}

type oauthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	Expiry       string `json:"expiry,omitempty"`
}

var (
	pathLocksMu sync.Mutex
	pathLocks   = make(map[string]*sync.Mutex)
)

func getPathMutex(canonicalPath string) *sync.Mutex {
	pathLocksMu.Lock()
	defer pathLocksMu.Unlock()
	m, ok := pathLocks[canonicalPath]
	if !ok {
		m = &sync.Mutex{}
		pathLocks[canonicalPath] = m
	}
	return m
}

func lockTokenFile(path string) (func(), error) {
	canonical := filepath.Clean(path)
	memLock := getPathMutex(canonical)
	memLock.Lock()

	flockUnlock, err := lockCredentialFile(canonical)
	if err != nil {
		memLock.Unlock()
		return nil, err
	}

	return func() {
		flockUnlock()
		memLock.Unlock()
	}, nil
}

func configDir(acct core.AccountConfig) string {
	if dir := strings.TrimSpace(acct.Path("config_dir", "")); dir != "" {
		return dir
	}
	box := strings.TrimSpace(acct.Hint("box_name", ""))
	if box == "" {
		id := strings.TrimSpace(acct.ID)
		if strings.HasPrefix(id, "antigravity-") {
			box = strings.TrimPrefix(id, "antigravity-")
		}
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	if box != "" {
		return filepath.Join(home, ".agy-containers", box, ".gemini", "antigravity-cli")
	}
	if strings.TrimSpace(acct.ID) == defaultAccountID {
		return filepath.Join(home, ".gemini", "antigravity-cli")
	}
	return ""
}

func tokenFilePath(acct core.AccountConfig) string {
	if override := strings.TrimSpace(acct.Path("oauth_token_file", "")); override != "" {
		return override
	}
	dir := configDir(acct)
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, oauthTokenFile)
}

func boxName(acct core.AccountConfig) string {
	if name := strings.TrimSpace(acct.Hint("box_name", "")); name != "" {
		return name
	}
	id := strings.TrimSpace(acct.ID)
	if strings.HasPrefix(id, "antigravity-") {
		return strings.TrimPrefix(id, "antigravity-")
	}
	dir := configDir(acct)
	if dir == "" {
		return ""
	}
	parts := strings.Split(filepath.ToSlash(dir), "/")
	for i := 0; i+3 < len(parts); i++ {
		if parts[i] == ".agy-containers" {
			return parts[i+1]
		}
	}
	return ""
}

func loadOAuthToken(path string) (oauthTokenFilePayload, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return oauthTokenFilePayload{}, err
	}
	var rawTop map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawTop); err != nil {
		return oauthTokenFilePayload{}, fmt.Errorf("parse oauth token file: %w", err)
	}
	if rawTop == nil {
		return oauthTokenFilePayload{}, fmt.Errorf("empty oauth token file")
	}

	var rawToken map[string]json.RawMessage
	var tok oauthToken
	tokenData, ok := rawTop["token"]
	if !ok {
		tokenData, ok = rawTop["Token"]
	}
	if ok {
		if err := json.Unmarshal(tokenData, &rawToken); err != nil {
			return oauthTokenFilePayload{}, fmt.Errorf("parse token field: %w", err)
		}
		if err := json.Unmarshal(tokenData, &tok); err != nil {
			return oauthTokenFilePayload{}, fmt.Errorf("parse token details: %w", err)
		}
	} else {
		// Fallback for flat token JSON structure
		if err := json.Unmarshal(data, &tok); err == nil && tok.AccessToken != "" {
			_ = json.Unmarshal(data, &rawToken)
		} else {
			return oauthTokenFilePayload{}, fmt.Errorf("missing token in oauth token file")
		}
	}
	if rawToken == nil {
		rawToken = make(map[string]json.RawMessage)
	}

	return oauthTokenFilePayload{
		rawTop:   rawTop,
		rawToken: rawToken,
		Token:    tok,
	}, nil
}

func writeOAuthToken(path string, payload oauthTokenFilePayload) error {
	if payload.rawTop == nil {
		payload.rawTop = make(map[string]json.RawMessage)
	}
	if payload.rawToken == nil {
		payload.rawToken = make(map[string]json.RawMessage)
	}

	// Preserve unknown fields inside token object while updating known token fields
	accessBytes, err := json.Marshal(payload.Token.AccessToken)
	if err != nil {
		return err
	}
	payload.rawToken["access_token"] = accessBytes

	if payload.Token.RefreshToken != "" {
		refreshBytes, err := json.Marshal(payload.Token.RefreshToken)
		if err != nil {
			return err
		}
		payload.rawToken["refresh_token"] = refreshBytes
	}

	if payload.Token.TokenType != "" {
		typeBytes, err := json.Marshal(payload.Token.TokenType)
		if err != nil {
			return err
		}
		payload.rawToken["token_type"] = typeBytes
	}

	if payload.Token.Expiry != "" {
		expiryBytes, err := json.Marshal(payload.Token.Expiry)
		if err != nil {
			return err
		}
		payload.rawToken["expiry"] = expiryBytes
	}

	marshaledToken, err := json.Marshal(payload.rawToken)
	if err != nil {
		return err
	}
	payload.rawTop["token"] = marshaledToken

	data, err := json.MarshalIndent(payload.rawTop, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".antigravity-oauth-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func tokenExpired(tok oauthToken, now time.Time) bool {
	access := strings.TrimSpace(tok.AccessToken)
	if access == "" {
		return true
	}
	expiry := strings.TrimSpace(tok.Expiry)
	if expiry == "" {
		return false
	}
	parsed, err := time.Parse(time.RFC3339Nano, expiry)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, expiry)
	}
	if err != nil {
		return false
	}
	return !parsed.After(now.Add(tokenExpirySkew))
}

func oauthClientCredentials() (clientID, clientSecret string, ok bool) {
	clientID = strings.TrimSpace(os.Getenv(oauthClientIDEnv))
	if clientID == "" {
		clientID = strings.TrimSpace(os.Getenv(oauthClientIDEnvAlias))
	}
	clientSecret = strings.TrimSpace(os.Getenv(oauthClientSecretEnv))
	if clientSecret == "" {
		clientSecret = strings.TrimSpace(os.Getenv(oauthClientSecretEnvAlias))
	}
	if clientID == "" || clientSecret == "" {
		return "", "", false
	}
	return clientID, clientSecret, true
}

func refreshAccessToken(ctx context.Context, refreshToken string, client *http.Client) (oauthToken, error) {
	clientID, clientSecret, ok := oauthClientCredentials()
	if !ok {
		return oauthToken{}, &AuthError{
			Reason:  "unconfigured_client",
			Message: fmt.Sprintf("oauth client not configured (%s / %s)", oauthClientIDEnv, oauthClientSecretEnv),
		}
	}
	return refreshAccessTokenWithBackoff(ctx, refreshToken, client, clientID, clientSecret)
}

func refreshAccessTokenWithBackoff(ctx context.Context, refreshToken string, client *http.Client, clientID, clientSecret string) (oauthToken, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	var lastErr error
	maxAttempts := 2
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return oauthToken{}, err
		}

		tok, err := doTokenRefreshRequest(ctx, refreshToken, client, clientID, clientSecret)
		if err == nil {
			return tok, nil
		}

		var authErr *AuthError
		if errors.As(err, &authErr) {
			// Non-retryable auth error
			return oauthToken{}, err
		}

		lastErr = err
		if attempt < maxAttempts {
			select {
			case <-time.After(100 * time.Millisecond):
			case <-ctx.Done():
				return oauthToken{}, ctx.Err()
			}
		}
	}

	return oauthToken{}, lastErr
}

func doTokenRefreshRequest(ctx context.Context, refreshToken string, client *http.Client, clientID, clientSecret string) (oauthToken, error) {
	form := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return oauthToken{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return oauthToken{}, &RefreshTransientError{
			Message: fmt.Sprintf("token endpoint network error: %v", err),
		}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
			var errPayload struct {
				Error            string `json:"error"`
				ErrorDescription string `json:"error_description"`
			}
			_ = json.Unmarshal(body, &errPayload)
			reason := errPayload.Error
			if reason == "" {
				reason = "invalid_grant"
			}
			return oauthToken{}, &AuthError{
				Reason:     reason,
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("token refresh rejected (HTTP %d: %s)", resp.StatusCode, reason),
			}
		}
		return oauthToken{}, &RefreshTransientError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("token refresh HTTP %d: %s", resp.StatusCode, truncate(string(body), 100)),
		}
	}

	var decoded struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return oauthToken{}, fmt.Errorf("parse token refresh response: %w", err)
	}
	if strings.TrimSpace(decoded.AccessToken) == "" {
		return oauthToken{}, fmt.Errorf("empty access_token in refresh response")
	}
	tok := oauthToken{
		AccessToken:  decoded.AccessToken,
		RefreshToken: strings.TrimSpace(decoded.RefreshToken),
		TokenType:    decoded.TokenType,
	}
	if decoded.ExpiresIn > 0 {
		tok.Expiry = time.Now().UTC().Add(time.Duration(decoded.ExpiresIn) * time.Second).Format(time.RFC3339Nano)
	}
	return tok, nil
}

// ensureAccessToken returns a usable access token, refreshing directly over HTTP when expired.
// If forceRefresh is true, a token renewal is forced even if the local expiry is in the future.
func ensureAccessToken(ctx context.Context, acct core.AccountConfig, client *http.Client, forceRefresh ...bool) (accessToken string, path string, refreshed bool, err error) {
	force := len(forceRefresh) > 0 && forceRefresh[0]
	path = tokenFilePath(acct)
	if path == "" {
		return "", "", false, &AuthError{
			Reason:  "no_path",
			Message: "oauth token path unavailable (set config_dir)",
		}
	}

	payload, err := loadOAuthToken(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", path, false, &AuthError{
				Reason:  "missing_file",
				Message: fmt.Sprintf("oauth token file missing: %s", path),
			}
		}
		return "", path, false, &AuthError{
			Reason:  "invalid_json",
			Message: fmt.Sprintf("parse oauth token file: %v", err),
		}
	}

	now := time.Now().UTC()
	if !force && !tokenExpired(payload.Token, now) {
		return payload.Token.AccessToken, path, false, nil
	}

	// Token expired or forced: coordinate per canonical path
	unlock, lockErr := lockTokenFile(path)
	if lockErr != nil {
		return "", path, false, fmt.Errorf("lock token file: %w", lockErr)
	}
	defer unlock()

	// Re-read under lock in case another caller or process already refreshed it
	reloaded, err := loadOAuthToken(path)
	if err == nil {
		if !force && !tokenExpired(reloaded.Token, time.Now().UTC()) {
			return reloaded.Token.AccessToken, path, false, nil
		}
		if force && reloaded.Token.AccessToken != payload.Token.AccessToken && !tokenExpired(reloaded.Token, time.Now().UTC()) {
			return reloaded.Token.AccessToken, path, true, nil
		}
		payload = reloaded
	}

	refreshTok := strings.TrimSpace(payload.Token.RefreshToken)
	if refreshTok == "" {
		return "", path, false, &AuthError{
			Reason:  "no_refresh_token",
			Message: "token expired and no refresh token available; please sign in",
		}
	}

	clientID, clientSecret, ok := oauthClientCredentials()
	if !ok {
		return "", path, false, &AuthError{
			Reason:  "unconfigured_client",
			Message: fmt.Sprintf("oauth client not configured (%s / %s); please configure credentials or sign in", oauthClientIDEnv, oauthClientSecretEnv),
		}
	}

	newTok, refreshErr := refreshAccessTokenWithBackoff(ctx, refreshTok, client, clientID, clientSecret)
	if refreshErr != nil {
		return "", path, false, refreshErr
	}

	if newTok.RefreshToken == "" {
		newTok.RefreshToken = refreshTok
	}
	payload.Token = newTok

	// Detect if an external process updated the file with a valid token while we refreshed
	if curPayload, curErr := loadOAuthToken(path); curErr == nil {
		if curPayload.Token.AccessToken != reloaded.Token.AccessToken && !tokenExpired(curPayload.Token, time.Now().UTC()) {
			return curPayload.Token.AccessToken, path, true, nil
		}
	}

	if writeErr := writeOAuthToken(path, payload); writeErr != nil {
		return "", path, false, fmt.Errorf("persist refreshed token: %w", writeErr)
	}

	return newTok.AccessToken, path, true, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
