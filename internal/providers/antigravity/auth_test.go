//go:build !windows

package antigravity

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/nurulislamz/agentusage/internal/core"
)

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
