package browsercookies

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/browserutils/kooky"
)

type mockBrowserInfo struct {
	browser  string
	filePath string
}

func (m mockBrowserInfo) Browser() string        { return m.browser }
func (m mockBrowserInfo) Profile() string        { return "Default" }
func (m mockBrowserInfo) IsDefaultProfile() bool { return true }
func (m mockBrowserInfo) FilePath() string       { return m.filePath }

type mockCookieStore struct {
	mockBrowserInfo
	cookies []*kooky.Cookie
	closed  bool
}

func (m *mockCookieStore) SetCookies(_ *url.URL, _ []*http.Cookie) {}
func (m *mockCookieStore) Cookies(_ *url.URL) []*http.Cookie       { return nil }
func (m *mockCookieStore) SubJar(_ context.Context, _ ...kooky.Filter) (http.CookieJar, error) {
	return nil, nil
}
func (m *mockCookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	return func(yield func(*kooky.Cookie, error) bool) {
		for _, c := range m.cookies {
			if c != nil {
				passes := true
				for _, f := range filters {
					if f != nil && !f.Filter(c) {
						passes = false
						break
					}
				}
				if !passes {
					continue
				}
			}
			if !yield(c, nil) {
				return
			}
		}
	}
}
func (m *mockCookieStore) Close() error {
	m.closed = true
	return nil
}

func TestNormalizeDomain(t *testing.T) {
	cases := map[string]string{
		"example.com":   "example.com",
		".example.com":  "example.com",
		"  Example.Com": "example.com",
		"":              "",
	}
	for in, want := range cases {
		if got := normalizeDomain(in); got != want {
			t.Errorf("normalizeDomain(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMatches(t *testing.T) {
	cases := []struct {
		cookieDomain string
		lookupDomain string
		want         bool
	}{
		// Bare domain, no leading dot — exact match only.
		{"opencode.ai", "opencode.ai", true},
		{"opencode.ai", "console.opencode.ai", false},
		{"opencode.ai", ".opencode.ai", true},

		// Leading-dot domain — covers the bare host and any subdomain.
		{".opencode.ai", "opencode.ai", true},
		{".opencode.ai", "console.opencode.ai", true},
		{".opencode.ai", "deep.console.opencode.ai", true},
		{".opencode.ai", "evil-opencode.ai", false},

		// Non-matching.
		{"opencode.ai", "perplexity.ai", false},

		// Case-insensitive.
		{".OpenCode.AI", "opencode.ai", true},

		// Empty inputs.
		{"", "opencode.ai", false},
		{".opencode.ai", "", false},
	}
	for _, tc := range cases {
		if got := matches(tc.cookieDomain, tc.lookupDomain); got != tc.want {
			t.Errorf("matches(%q, %q) = %v, want %v", tc.cookieDomain, tc.lookupDomain, got, tc.want)
		}
	}
}

func TestCanonicalBrowser(t *testing.T) {
	cases := map[string]string{
		"chromium":           "chrome",
		"google-chrome":      "chrome",
		"Chrome":             "chrome",
		"firefox":            "firefox",
		"Mozilla Firefox":    "firefox",
		"safari":             "safari",
		"Microsoft Edge":     "edge",
		"Brave Browser":      "brave",
		"vivaldi":            "vivaldi",
		"opera":              "opera",
		"":                   "",
		"unknown-browser-99": "unknown-browser-99",
	}
	for in, want := range cases {
		if got := canonicalBrowser(in); got != want {
			t.Errorf("canonicalBrowser(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsKeychainProtected(t *testing.T) {
	tests := []struct {
		browser string
		want    bool
	}{
		{"chrome", true},
		{"google-chrome", true},
		{"chromium", true},
		{"edge", true},
		{"brave", true},
		{"vivaldi", true},
		{"opera", true},
		{"firefox", false},
		{"safari", false},
		{"unknown", false},
	}
	for _, tt := range tests {
		if got := IsKeychainProtected(tt.browser); got != tt.want {
			t.Errorf("IsKeychainProtected(%q) = %v, want %v", tt.browser, got, tt.want)
		}
	}
}

func TestCookie_IsExpired(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name string
		c    Cookie
		want bool
	}{
		{"future", Cookie{Expires: now.Add(time.Hour)}, false},
		{"past", Cookie{Expires: now.Add(-time.Hour)}, true},
		{"zero is session-cookie, not expired", Cookie{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.c.IsExpired(); got != tc.want {
				t.Errorf("IsExpired = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPickStoresForBrowser(t *testing.T) {
	sChrome := &mockCookieStore{mockBrowserInfo: mockBrowserInfo{browser: "chrome"}}
	sFirefox := &mockCookieStore{mockBrowserInfo: mockBrowserInfo{browser: "firefox"}}
	sSafari := &mockCookieStore{mockBrowserInfo: mockBrowserInfo{browser: "safari"}}
	sEdge := &mockCookieStore{mockBrowserInfo: mockBrowserInfo{browser: "edge"}}

	allStores := []kooky.CookieStore{sChrome, nil, sFirefox, sSafari, sEdge}

	// 1. Auto browser (empty string) -> picks only non-keychain stores (firefox, safari)
	autoStores := pickStoresForBrowser(allStores, "")
	if len(autoStores) != 2 {
		t.Fatalf("pickStoresForBrowser(all, \"\") got %d stores, want 2", len(autoStores))
	}
	for _, s := range autoStores {
		bn := canonicalBrowser(s.Browser())
		if bn != "firefox" && bn != "safari" {
			t.Errorf("unexpected store in auto list: %q", bn)
		}
	}

	// 2. Specific browser ("chrome") -> picks only chrome
	chromeStores := pickStoresForBrowser(allStores, "chrome")
	if len(chromeStores) != 1 || canonicalBrowser(chromeStores[0].Browser()) != "chrome" {
		t.Errorf("pickStoresForBrowser(all, 'chrome') = %v, want 1 chrome store", chromeStores)
	}

	// 3. Specific browser ("firefox") -> picks only firefox
	ffStores := pickStoresForBrowser(allStores, "firefox")
	if len(ffStores) != 1 || canonicalBrowser(ffStores[0].Browser()) != "firefox" {
		t.Errorf("pickStoresForBrowser(all, 'firefox') = %v, want 1 firefox store", ffStores)
	}

	// 4. Non-matching browser -> returns empty slice
	noneStores := pickStoresForBrowser(allStores, "opera")
	if len(noneStores) != 0 {
		t.Errorf("pickStoresForBrowser(all, 'opera') = %v, want empty", noneStores)
	}
}

func TestReadFromStores(t *testing.T) {
	now := time.Now()
	olderExp := now.Add(time.Hour)
	fresherExp := now.Add(24 * time.Hour)

	bInfo := mockBrowserInfo{browser: "google-chrome", filePath: "/path/to/Cookies"}

	s1 := &mockCookieStore{
		mockBrowserInfo: bInfo,
		cookies: []*kooky.Cookie{
			{
				Cookie: http.Cookie{
					Name:     "session",
					Value:    "older_val",
					Domain:   ".example.com",
					Path:     "/",
					Expires:  olderExp,
					HttpOnly: true,
					Secure:   true,
				},
				Browser: bInfo,
			},
			nil, // Nil cookie in sequence should be handled safely
			{
				Cookie: http.Cookie{
					Name:    "unrelated",
					Value:   "foo",
					Domain:  ".example.com",
					Expires: fresherExp,
				},
				Browser: bInfo,
			},
		},
	}

	s2 := &mockCookieStore{
		mockBrowserInfo: bInfo,
		cookies: []*kooky.Cookie{
			{
				Cookie: http.Cookie{
					Name:     "session",
					Value:    "fresher_val",
					Domain:   ".example.com",
					Path:     "/",
					Expires:  fresherExp,
					HttpOnly: true,
					Secure:   true,
				},
				Browser: bInfo,
			},
			{
				Cookie: http.Cookie{
					Name:    "session",
					Value:   "wrong_domain",
					Domain:  ".different.com",
					Expires: fresherExp.Add(time.Hour),
				},
				Browser: bInfo,
			},
		},
	}

	stores := []kooky.CookieStore{nil, s1, s2}

	cookie, found := readFromStores(stores, "example.com", "session")
	if !found {
		t.Fatal("readFromStores failed to find matching cookie")
	}
	if cookie.Value != "fresher_val" {
		t.Errorf("cookie.Value = %q, want fresher 'fresher_val'", cookie.Value)
	}
	if cookie.Source != "chrome" {
		t.Errorf("cookie.Source = %q, want canonical 'chrome'", cookie.Source)
	}
	if cookie.StorePath != "/path/to/Cookies" {
		t.Errorf("cookie.StorePath = %q, want '/path/to/Cookies'", cookie.StorePath)
	}
	if !cookie.HTTPOnly || !cookie.Secure {
		t.Errorf("HTTPOnly=%v Secure=%v, want both true", cookie.HTTPOnly, cookie.Secure)
	}

	// Non-matching lookup
	_, foundMissing := readFromStores(stores, "example.com", "nonexistent")
	if foundMissing {
		t.Error("expected found=false for missing cookie name")
	}
}

func TestFakeReader_FindsCookieByDomainAndName(t *testing.T) {
	exp := time.Now().Add(time.Hour)
	f := &FakeReader{
		Cookies: []Cookie{
			{Name: "noise", Value: "x", Domain: ".opencode.ai", Source: "chrome"},
			{Name: "auth", Value: "wanted", Domain: ".opencode.ai", Expires: exp, Source: "chrome"},
			{Name: "auth", Value: "different-domain", Domain: ".perplexity.ai", Source: "firefox"},
		},
	}
	got, err := f.ReadCookie(context.Background(), "opencode.ai", "auth", "")
	if err != nil {
		t.Fatalf("ReadCookie error: %v", err)
	}
	if got.Value != "wanted" {
		t.Errorf("Value = %q, want 'wanted'", got.Value)
	}
	if got.Source != "chrome" {
		t.Errorf("Source = %q, want chrome", got.Source)
	}
	if got.Expires != exp {
		t.Errorf("Expires = %v, want %v", got.Expires, exp)
	}
	if f.Calls() != 1 {
		t.Errorf("Calls = %d, want 1", f.Calls())
	}
}

func TestFakeReader_NoCookieReturnsErrNoCookieFound(t *testing.T) {
	f := &FakeReader{Cookies: []Cookie{
		{Name: "auth", Value: "wrong-domain", Domain: ".other.com", Source: "chrome"},
	}}
	_, err := f.ReadCookie(context.Background(), "opencode.ai", "auth", "")
	if !errors.Is(err, ErrNoCookieFound) {
		t.Errorf("err = %v, want ErrNoCookieFound", err)
	}
}

func TestFakeReader_PropagatesError(t *testing.T) {
	want := errors.New("simulated keychain failure")
	f := &FakeReader{Err: want}
	_, err := f.ReadCookie(context.Background(), "opencode.ai", "auth", "")
	if err != want {
		t.Errorf("err = %v, want %v", err, want)
	}
	browsers, err := f.AvailableBrowsers(context.Background())
	if err != want {
		t.Errorf("err = %v, want %v", err, want)
	}
	if browsers != nil {
		t.Errorf("AvailableBrowsers = %v, want nil on error", browsers)
	}
}

func TestFakeReader_AvailableBrowsersDistinct(t *testing.T) {
	f := &FakeReader{Cookies: []Cookie{
		{Name: "a", Domain: "x.com", Source: "chrome"},
		{Name: "b", Domain: "x.com", Source: "chrome"},
		{Name: "c", Domain: "x.com", Source: "firefox"},
		{Name: "d", Domain: "x.com", Source: ""},
	}}
	browsers, err := f.AvailableBrowsers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"chrome": true, "firefox": true}
	if len(browsers) != len(want) {
		t.Fatalf("got %d, want %d: %v", len(browsers), len(want), browsers)
	}
	for _, b := range browsers {
		if !want[b] {
			t.Errorf("unexpected browser %q in result", b)
		}
	}
}

// New() returns a non-nil reader (this is a smoke test — we don't want the
// real kooky scan to run during unit tests because it triggers keychain
// prompts on macOS, but we do verify the constructor doesn't panic).
func TestNew_ReturnsReader(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatal("New returned nil")
	}
	r2 := NewWithTimeout(time.Second)
	if r2 == nil {
		t.Fatal("NewWithTimeout returned nil")
	}
}

func TestKookyReader_LiveMethods(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	r := NewWithTimeout(500 * time.Millisecond)

	// AvailableBrowsers on real reader
	browsers, err := r.AvailableBrowsers(ctx)
	if err != nil {
		t.Logf("AvailableBrowsers returned err: %v", err)
	}
	t.Logf("AvailableBrowsers: %v", browsers)

	// ReadCookie on real reader with non-existent cookie
	_, err = r.ReadCookie(ctx, "nonexistent-domain-test.local", "auth_token", "firefox")
	if !errors.Is(err, ErrNoCookieFound) && err != context.DeadlineExceeded {
		t.Logf("ReadCookie returned unexpected error: %v", err)
	}

	// ReadCookie with auto browser ("")
	_, err = r.ReadCookie(ctx, "nonexistent-domain-test.local", "auth_token", "")
	if !errors.Is(err, ErrNoCookieFound) && err != context.DeadlineExceeded {
		t.Logf("ReadCookie (auto) returned unexpected error: %v", err)
	}
}

func TestReadCookieWSL_KeyValidation(t *testing.T) {
	// Create a temp directory to simulate home dir with agentusage config
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tempDir, ".config", "agentusage")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}

	r := &kookyReader{}
	ctx := context.Background()

	// 1. Invalid base64 in key file -> ErrNoCookieFound
	keyFile := filepath.Join(configDir, "chrome_key")
	if err := os.WriteFile(keyFile, []byte("invalid-base64!@#$"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := r.readCookieWSL(ctx, "example.com", "token", "chrome")
	if !errors.Is(err, ErrNoCookieFound) {
		t.Errorf("readCookieWSL with invalid base64 got %v, want ErrNoCookieFound", err)
	}

	// 2. Base64 with wrong key length (!= 32 bytes) -> ErrNoCookieFound
	shortKey := base64.StdEncoding.EncodeToString([]byte("too-short"))
	if err := os.WriteFile(keyFile, []byte(shortKey), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = r.readCookieWSL(ctx, "example.com", "token", "chrome")
	if !errors.Is(err, ErrNoCookieFound) {
		t.Errorf("readCookieWSL with short key got %v, want ErrNoCookieFound", err)
	}

	// 3. Valid 32-byte key base64 -> attempts WSL scan and returns ErrNoCookieFound (unless /mnt/c/Users exists with matching cookie)
	valid32ByteKey := make([]byte, 32)
	for i := range valid32ByteKey {
		valid32ByteKey[i] = byte(i + 1)
	}
	validKeyB64 := base64.StdEncoding.EncodeToString(valid32ByteKey)
	if err := os.WriteFile(keyFile, []byte(validKeyB64), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = r.readCookieWSL(ctx, "nonexistent-domain-test.local", "token", "chrome")
	if !errors.Is(err, ErrNoCookieFound) {
		t.Logf("readCookieWSL with valid key format returned: %v", err)
	}
}

