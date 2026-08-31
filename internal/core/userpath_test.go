package core

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestUserLocalBinDir(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	got := UserLocalBinDir()
	want := filepath.Join(home, ".local", "bin")
	if got != want {
		t.Fatalf("UserLocalBinDir() = %q, want %q", got, want)
	}
}

func TestPrependUserLocalBin_PrependsOnce(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	dir := filepath.Join(home, ".local", "bin")

	got, ok := prependUserLocalBin("/usr/bin")
	if !ok {
		t.Fatal("expected PATH to change")
	}
	want := dir + string(os.PathListSeparator) + "/usr/bin"
	if got != want {
		t.Fatalf("prependUserLocalBin() = %q, want %q", got, want)
	}

	again, ok := prependUserLocalBin(got)
	if ok {
		t.Fatal("second prepend should be a no-op")
	}
	if again != got {
		t.Fatalf("second prependUserLocalBin() = %q, want %q", again, got)
	}
}

func TestEnsureUserLocalBinOnPATH(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	t.Setenv("PATH", "/usr/bin")

	EnsureUserLocalBinOnPATH()
	dir := filepath.Join(home, ".local", "bin")
	path := os.Getenv("PATH")
	if !strings.HasPrefix(path, dir+string(os.PathListSeparator)) {
		t.Fatalf("PATH = %q, want prefix %q", path, dir)
	}

	EnsureUserLocalBinOnPATH()
	if strings.Count(os.Getenv("PATH"), dir) != 1 {
		t.Fatalf("PATH prepended more than once: %q", os.Getenv("PATH"))
	}
}

func TestEnvironWithUserLocalBin_ReplacesPATH(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	dir := filepath.Join(home, ".local", "bin")

	got := EnvironWithUserLocalBin([]string{"FOO=bar", "PATH=/usr/bin"})
	found := false
	for _, entry := range got {
		if strings.HasPrefix(entry, "PATH=") {
			found = true
			if entry != "PATH="+dir+string(os.PathListSeparator)+"/usr/bin" {
				t.Fatalf("PATH entry = %q", entry)
			}
		}
	}
	if !found {
		t.Fatal("missing PATH in env")
	}
}

func setHome(t *testing.T, home string) {
	t.Helper()
	t.Setenv("HOME", home)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
	}
}
