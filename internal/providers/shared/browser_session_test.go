package shared

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/browsercookies"
	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

func TestLoadOrRefreshBrowserSession_DefaultPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	reader := &browsercookies.FakeReader{
		Cookies: []browsercookies.Cookie{{
			Name:    "sess_id",
			Value:   "val_123",
			Domain:  "test.com",
			Source:  "firefox",
			Expires: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		}},
	}
	acct := core.AccountConfig{
		ID:       "acct_default",
		Provider: "test_prov",
		BrowserCookie: &core.BrowserCookieRef{
			Domain:     "test.com",
			CookieName: "sess_id",
		},
	}

	session, ok, err := LoadOrRefreshBrowserSession(context.Background(), acct, reader)
	if err != nil {
		t.Fatalf("LoadOrRefreshBrowserSession failed: %v", err)
	}
	if !ok || session.Value != "val_123" {
		t.Fatalf("session = %+v, want val_123", session)
	}
}

func TestLoadOrRefreshBrowserSessionFrom_RefreshesStoredSession(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	if err := config.SaveSessionTo(path, "perplexity", config.BrowserSession{
		Domain:        ".perplexity.ai",
		CookieName:    "__Secure-next-auth.session-token",
		Value:         "old-cookie",
		SourceBrowser: "firefox",
		CapturedAt:    "2026-04-30T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}

	reader := &browsercookies.FakeReader{
		Cookies: []browsercookies.Cookie{{
			Name:    "__Secure-next-auth.session-token",
			Value:   "fresh-cookie",
			Domain:  ".perplexity.ai",
			Source:  "chrome",
			Expires: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
		}},
	}
	acct := core.AccountConfig{
		ID:       "perplexity",
		Provider: "perplexity",
		BrowserCookie: &core.BrowserCookieRef{
			Domain:        ".perplexity.ai",
			CookieName:    "__Secure-next-auth.session-token",
			SourceBrowser: "chrome",
		},
	}

	session, ok, err := loadOrRefreshBrowserSessionFrom(path, context.Background(), acct, reader)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("session not found after refresh")
	}
	if session.Value != "fresh-cookie" {
		t.Fatalf("Value = %q, want fresh-cookie", session.Value)
	}
	if session.SourceBrowser != "chrome" {
		t.Fatalf("SourceBrowser = %q, want chrome", session.SourceBrowser)
	}

	saved, ok, err := config.LoadSessionFrom(path, "perplexity")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("session not saved after refresh")
	}
	if saved.Value != "fresh-cookie" {
		t.Fatalf("saved session value = %q, want fresh-cookie", saved.Value)
	}
}

func TestLoadOrRefreshBrowserSessionFrom_FallsBackToStoredOnNoCookie(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	stored := config.BrowserSession{
		Domain:        ".opencode.ai",
		CookieName:    "auth",
		Value:         "stored-cookie",
		SourceBrowser: "chrome",
		CapturedAt:    "2026-04-30T00:00:00Z",
	}
	if err := config.SaveSessionTo(path, "opencode", stored); err != nil {
		t.Fatal(err)
	}

	reader := &browsercookies.FakeReader{Err: browsercookies.ErrNoCookieFound}
	acct := core.AccountConfig{
		ID:       "opencode",
		Provider: "opencode",
		BrowserCookie: &core.BrowserCookieRef{
			Domain:        ".opencode.ai",
			CookieName:    "auth",
			SourceBrowser: "chrome",
		},
	}

	session, ok, err := loadOrRefreshBrowserSessionFrom(path, context.Background(), acct, reader)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected stored fallback session")
	}
	if session != stored {
		t.Fatalf("session = %+v, want %+v", session, stored)
	}
}

func TestLoadOrRefreshBrowserSessionFrom_FallsBackToStoredOnGeneralError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	stored := config.BrowserSession{
		Domain:        ".opencode.ai",
		CookieName:    "auth",
		Value:         "stored-cookie",
		SourceBrowser: "chrome",
		CapturedAt:    "2026-04-30T00:00:00Z",
	}
	if err := config.SaveSessionTo(path, "opencode", stored); err != nil {
		t.Fatal(err)
	}

	reader := &browsercookies.FakeReader{Err: errors.New("browser locked")}
	acct := core.AccountConfig{
		ID:       "opencode",
		Provider: "opencode",
		BrowserCookie: &core.BrowserCookieRef{
			Domain:        ".opencode.ai",
			CookieName:    "auth",
			SourceBrowser: "chrome",
		},
	}

	session, ok, err := loadOrRefreshBrowserSessionFrom(path, context.Background(), acct, reader)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || session.Value != "stored-cookie" {
		t.Fatalf("expected stored fallback session, got %+v", session)
	}
}

func TestLoadOrRefreshBrowserSessionFrom_UsesStoredRefWhenAccountMissingBrowserCookie(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	if err := config.SaveSessionTo(path, "perplexity", config.BrowserSession{
		Domain:        ".perplexity.ai",
		CookieName:    "__Secure-next-auth.session-token",
		Value:         "stored-cookie",
		SourceBrowser: "safari",
	}); err != nil {
		t.Fatal(err)
	}

	reader := &browsercookies.FakeReader{
		Cookies: []browsercookies.Cookie{{
			Name:   "__Secure-next-auth.session-token",
			Value:  "fresh-cookie",
			Domain: ".perplexity.ai",
			Source: "safari",
		}},
	}
	acct := core.AccountConfig{ID: "perplexity", Provider: "perplexity"}

	session, ok, err := loadOrRefreshBrowserSessionFrom(path, context.Background(), acct, reader)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected refreshed session")
	}
	if session.Value != "fresh-cookie" {
		t.Fatalf("Value = %q, want fresh-cookie", session.Value)
	}
}

func TestLoadOrRefreshBrowserSessionFrom_PropagatesReadErrorWithoutStoredFallback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	reader := &browsercookies.FakeReader{Err: errors.New("keychain denied")}
	acct := core.AccountConfig{
		ID:       "perplexity",
		Provider: "perplexity",
		BrowserCookie: &core.BrowserCookieRef{
			Domain:     ".perplexity.ai",
			CookieName: "__Secure-next-auth.session-token",
		},
	}

	_, ok, err := loadOrRefreshBrowserSessionFrom(path, context.Background(), acct, reader)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if ok {
		t.Fatal("ok = true, want false")
	}
}

func TestLoadOrRefreshBrowserSessionFrom_MissingRefAndNoStored(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	acct := core.AccountConfig{ID: "no_ref"}

	session, ok, err := loadOrRefreshBrowserSessionFrom(path, context.Background(), acct, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok || session.Value != "" {
		t.Fatalf("expected ok=false, got ok=%v session=%+v", ok, session)
	}
}

func TestLoadOrRefreshBrowserSessionFrom_EmptyDomainOrCookie(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	acct := core.AccountConfig{
		ID: "empty_domain",
		BrowserCookie: &core.BrowserCookieRef{
			Domain:     "  ",
			CookieName: "cookie",
		},
	}

	session, ok, err := loadOrRefreshBrowserSessionFrom(path, context.Background(), acct, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok || session.Value != "" {
		t.Fatalf("expected ok=false, got ok=%v", ok)
	}
}

func TestLoadOrRefreshBrowserSessionFrom_NoCookieFoundReturnsFalse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	reader := &browsercookies.FakeReader{Err: browsercookies.ErrNoCookieFound}
	acct := core.AccountConfig{
		ID: "no_cookie_found",
		BrowserCookie: &core.BrowserCookieRef{
			Domain:     "example.com",
			CookieName: "auth",
		},
	}

	session, ok, err := loadOrRefreshBrowserSessionFrom(path, context.Background(), acct, reader)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok || session.Value != "" {
		t.Fatalf("expected ok=false, got ok=%v", ok)
	}
}

func TestLoadOrRefreshBrowserSessionFrom_SaveError(t *testing.T) {
	// Use an uncreatable path (e.g. filename under a non-directory)
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "file_as_dir")
	if err := os.WriteFile(filePath, []byte("blocker"), 0o600); err != nil {
		t.Fatal(err)
	}
	invalidPath := filepath.Join(filePath, "invalid", "credentials.json")

	reader := &browsercookies.FakeReader{
		Cookies: []browsercookies.Cookie{{
			Name:   "auth",
			Value:  "cookie_val",
			Domain: "example.com",
		}},
	}
	acct := core.AccountConfig{
		ID: "save_fail",
		BrowserCookie: &core.BrowserCookieRef{
			Domain:     "example.com",
			CookieName: "auth",
		},
	}

	_, ok, err := loadOrRefreshBrowserSessionFrom(invalidPath, context.Background(), acct, reader)
	if err == nil {
		t.Fatal("expected save error, got nil")
	}
	if ok {
		t.Fatal("expected ok=false on save error")
	}
}
