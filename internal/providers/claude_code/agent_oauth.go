package claude_code

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
)

// agentusage's own Claude sign-in. Borrowing the Claude Code CLI's token from
// ~/.claude/.credentials.json only works while Claude Code keeps refreshing it,
// and refreshing it ourselves would rotate Claude Code's single-use refresh
// token out from under it. So `agentusage login claude` runs its own OAuth
// (PKCE) sign-in with the public Claude Code client ID, asks only for the
// user:profile scope the usage endpoint needs, and keeps the token pair in
// agentusage's credentials.json where it can be renewed independently.

const (
	// AgentOAuthAccountID keys the sign-in in credentials.json.
	AgentOAuthAccountID = "claude-code"

	claudeOAuthClientID    = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
	claudeOAuthScope       = "user:profile"
	claudeOAuthRedirectURI = "https://platform.claude.com/oauth/code/callback"

	// agentOAuthRefreshMargin renews the access token this long before it
	// expires, so a poll never sends a token that dies mid-request.
	agentOAuthRefreshMargin = 5 * time.Minute
	tokenExchangeTimeout    = 15 * time.Second
)

// tokenExchangeContext is independent of the caller's deadline. A cancelled
// Fetch must not abort a refresh whose POST may already have been applied.
func tokenExchangeContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(context.WithoutCancel(ctx), tokenExchangeTimeout)
}

// Package variables rather than constants so tests can point them at an
// httptest server and a temp credentials file.
var (
	claudeOAuthAuthorizeURL = "https://claude.ai/oauth/authorize"
	claudeOAuthTokenURL     = "https://platform.claude.com/v1/oauth/token"
	agentOAuthCredsPath     = config.CredentialsPath
)

// ErrAgentOAuthNotSignedIn means no agentusage Claude sign-in is stored.
var ErrAgentOAuthNotSignedIn = errors.New("not signed in (run `agentusage login claude`)")

// ErrAgentOAuthSignInExpired means Anthropic refused the stored refresh token;
// only a new `agentusage login claude` can recover.
var ErrAgentOAuthSignInExpired = errors.New("Claude sign-in expired (run `agentusage login claude`)")

// ClaudeLogin is one in-progress sign-in: the URL to open and the PKCE
// verifier/state needed to redeem the code the user pastes back.
type ClaudeLogin struct {
	AuthorizeURL string
	state        string
	verifier     string
}

// NewClaudeLogin starts a sign-in with a fresh PKCE pair and state.
func NewClaudeLogin() (*ClaudeLogin, error) {
	verifier, err := randomURLToken(32)
	if err != nil {
		return nil, fmt.Errorf("claude_code: generating PKCE verifier: %w", err)
	}
	state, err := randomURLToken(32)
	if err != nil {
		return nil, fmt.Errorf("claude_code: generating state: %w", err)
	}
	sum := sha256.Sum256([]byte(verifier))
	q := url.Values{
		"code":                  {"true"},
		"client_id":             {claudeOAuthClientID},
		"response_type":         {"code"},
		"redirect_uri":          {claudeOAuthRedirectURI},
		"scope":                 {claudeOAuthScope},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(sum[:])},
		"code_challenge_method": {"S256"},
		"state":                 {state},
	}
	return &ClaudeLogin{
		AuthorizeURL: claudeOAuthAuthorizeURL + "?" + q.Encode(),
		state:        state,
		verifier:     verifier,
	}, nil
}

// Complete redeems the pasted authorization code and stores the resulting
// token pair. The callback page shows the code as "code#state"; a bare code or
// the full callback URL is accepted too.
func (l *ClaudeLogin) Complete(ctx context.Context, pasted string) (config.OAuthToken, error) {
	code, state, err := parsePastedAuthCode(pasted)
	if err != nil {
		return config.OAuthToken{}, err
	}
	if state != "" && state != l.state {
		return config.OAuthToken{}, fmt.Errorf("claude_code: state mismatch; start the sign-in again")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	// Auth codes are single-use; finish the exchange even if the CLI's
	// 30s timeout is already firing.
	exchangeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	token, err := postClaudeToken(exchangeCtx, map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     claudeOAuthClientID,
		"code":          code,
		"state":         l.state,
		"redirect_uri":  claudeOAuthRedirectURI,
		"code_verifier": l.verifier,
	})
	if err != nil {
		return config.OAuthToken{}, err
	}
	path := agentOAuthCredsPath()
	agentOAuthMu.Lock()
	defer agentOAuthMu.Unlock()
	unlock, err := lockAgentOAuthFile(path)
	if err != nil {
		return config.OAuthToken{}, fmt.Errorf("claude_code: locking sign-in: %w", err)
	}
	defer unlock()
	if err := config.SaveOAuthTokenTo(path, AgentOAuthAccountID, token); err != nil {
		return config.OAuthToken{}, fmt.Errorf("claude_code: saving sign-in: %w", err)
	}
	return token, nil
}

func parsePastedAuthCode(pasted string) (code, state string, err error) {
	s := strings.TrimSpace(pasted)
	if u, perr := url.Parse(s); perr == nil && u.Scheme != "" && u.Host != "" {
		code, state = u.Query().Get("code"), u.Query().Get("state")
	} else if c, st, ok := strings.Cut(s, "#"); ok {
		code, state = c, st
	} else {
		code = s
	}
	if code == "" {
		return "", "", fmt.Errorf("claude_code: no authorization code found in input")
	}
	return code, state, nil
}

func randomURLToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type claudeTokenResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresIn    float64 `json:"expires_in"`
	Scope        string  `json:"scope"`
}

// signInExpiredCodes are the token-endpoint errors that mean the refresh token
// itself is dead. Anything else (outage, rate limit, garbage) fails only this
// attempt, and the next poll tries again.
var signInExpiredCodes = map[string]bool{
	"invalid_grant":             true,
	"refresh_token_expired":     true,
	"refresh_token_reused":      true,
	"refresh_token_invalidated": true,
}

func postClaudeToken(ctx context.Context, fields map[string]string) (config.OAuthToken, error) {
	body, err := json.Marshal(fields)
	if err != nil {
		return config.OAuthToken{}, fmt.Errorf("claude_code: encoding token request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claudeOAuthTokenURL, bytes.NewReader(body))
	if err != nil {
		return config.OAuthToken{}, fmt.Errorf("claude_code: creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: tokenExchangeTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return config.OAuthToken{}, fmt.Errorf("claude_code: token request failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))

	if resp.StatusCode != http.StatusOK {
		var oauthErr struct {
			Error any `json:"error"`
		}
		_ = json.Unmarshal(respBody, &oauthErr)
		code, _ := oauthErr.Error.(string)
		if m, ok := oauthErr.Error.(map[string]any); ok {
			code, _ = m["type"].(string)
			if c, ok := m["code"].(string); ok {
				code = c
			}
		}
		if fields["grant_type"] == "refresh_token" && signInExpiredCodes[code] {
			return config.OAuthToken{}, ErrAgentOAuthSignInExpired
		}
		return config.OAuthToken{}, fmt.Errorf("claude_code: token endpoint returned %d: %s", resp.StatusCode, truncateBody(respBody, 300))
	}

	var tr claudeTokenResponse
	if err := json.Unmarshal(respBody, &tr); err != nil || tr.AccessToken == "" {
		return config.OAuthToken{}, fmt.Errorf("claude_code: unexpected token response")
	}
	token := config.OAuthToken{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		Scope:        tr.Scope,
	}
	if tr.ExpiresIn > 0 {
		token.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn * float64(time.Second))).UTC().Format(time.RFC3339)
	}
	return token, nil
}

func truncateBody(b []byte, n int) string {
	s := strings.TrimSpace(string(b))
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

// agentOAuthNeedsRefresh reports whether the token is expired or about to be.
// A token with no recorded expiry is used until the API refuses it.
func agentOAuthConfigured() bool {
	tok, ok, err := config.LoadOAuthTokenFrom(agentOAuthCredsPath(), AgentOAuthAccountID)
	return err == nil && ok && tok.AccessToken != ""
}

func agentOAuthNeedsRefresh(t config.OAuthToken, now time.Time) bool {
	if t.ExpiresAt == "" {
		return false
	}
	exp, err := time.Parse(time.RFC3339, t.ExpiresAt)
	if err != nil {
		return true
	}
	return !exp.After(now.Add(agentOAuthRefreshMargin))
}

// agentOAuthMu serializes renewals within this process; lockAgentOAuthFile
// does the same across processes (daemon, `agentusage get`, the TUI), since
// a refresh token is single-use and two concurrent renewals would kill it.
var agentOAuthMu sync.Mutex

// agentOAuthAccessToken returns a usable access token from agentusage's own
// sign-in, renewing it first when it is about to expire or when force is set
// (the API just refused it with 401).
func agentOAuthAccessToken(ctx context.Context, force bool) (string, error) {
	path := agentOAuthCredsPath()
	tok, ok, err := config.LoadOAuthTokenFrom(path, AgentOAuthAccountID)
	if err != nil {
		return "", fmt.Errorf("claude_code: reading sign-in: %w", err)
	}
	if !ok || tok.AccessToken == "" {
		return "", ErrAgentOAuthNotSignedIn
	}
	if !force && !agentOAuthNeedsRefresh(tok, time.Now()) {
		return tok.AccessToken, nil
	}

	agentOAuthMu.Lock()
	defer agentOAuthMu.Unlock()
	unlock, err := lockAgentOAuthFile(path)
	if err != nil {
		return "", fmt.Errorf("claude_code: locking sign-in: %w", err)
	}
	defer unlock()

	// Re-read under the lock: another process may have renewed it already.
	current, ok, err := config.LoadOAuthTokenFrom(path, AgentOAuthAccountID)
	if err != nil {
		return "", fmt.Errorf("claude_code: reading sign-in: %w", err)
	}
	if !ok || current.AccessToken == "" {
		return "", ErrAgentOAuthNotSignedIn
	}
	if current.AccessToken != tok.AccessToken && !agentOAuthNeedsRefresh(current, time.Now()) {
		return current.AccessToken, nil
	}
	if !force && !agentOAuthNeedsRefresh(current, time.Now()) {
		return current.AccessToken, nil
	}
	if current.RefreshToken == "" {
		return "", ErrAgentOAuthSignInExpired
	}

	// Detach from the caller (daemon Fetch is 8s, hydrate is 3s, and JSONL
	// parsing often spends most of that). The refresh token is single-use:
	// if Anthropic rotates it and the parent ctx then cancels, we must still
	// read the body and persist or the sign-in is permanently dead.
	exchangeCtx, cancel := tokenExchangeContext(ctx)
	defer cancel()
	renewed, err := postClaudeToken(exchangeCtx, map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": current.RefreshToken,
		"client_id":     claudeOAuthClientID,
		"scope":         claudeOAuthScope,
	})
	if err != nil {
		return "", err
	}
	if renewed.RefreshToken == "" {
		renewed.RefreshToken = current.RefreshToken
	}
	if renewed.Scope == "" {
		renewed.Scope = current.Scope
	}
	if err := config.SaveOAuthTokenTo(path, AgentOAuthAccountID, renewed); err != nil {
		return "", fmt.Errorf("claude_code: saving renewed sign-in: %w", err)
	}
	return renewed.AccessToken, nil
}
