package webserve

import "testing"

func TestANSIToHTML_TrueColorBold(t *testing.T) {
	in := "\x1b[1m\x1b[38;2;126;184;247mOpenUsage\x1b[0m"
	got := ANSIToHTML(in)
	if !containsAll(got, "font-weight:700", "#7eb8f7", "OpenUsage") {
		t.Fatalf("unexpected html: %s", got)
	}
	if containsAll(got, "<script") {
		t.Fatal("should not contain script tags")
	}
}

func TestANSIToHTML_EscapesHTML(t *testing.T) {
	in := "<b>x</b> & y"
	got := ANSIToHTML(in)
	if got != "&lt;b&gt;x&lt;/b&gt; &amp; y" {
		t.Fatalf("got %q", got)
	}
}

func TestANSIToHTML_ResetClearsStyle(t *testing.T) {
	in := "\x1b[31mred\x1b[0m plain"
	got := ANSIToHTML(in)
	if !containsAll(got, "red", "plain") {
		t.Fatalf("got %q", got)
	}
	if !containsAll(got, "</span>") {
		t.Fatalf("expected closed span: %q", got)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
