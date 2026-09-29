package boxes

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateBoxName_RejectsTraversal(t *testing.T) {
	for _, name := range []string{"", ".", "..", "../.ssh", "..\\foo", "a/b", "-flag", "has space", "a:b"} {
		if err := validateBoxName(name); err == nil {
			t.Errorf("validateBoxName(%q) = nil, want error", name)
		}
	}
	if err := validateBoxName("work-box"); err != nil {
		t.Errorf("validateBoxName(work-box) = %v, want nil", err)
	}
}

func TestResolveBoxProfileDir_StaysUnderRoot(t *testing.T) {
	root := t.TempDir()
	got, err := resolveBoxProfileDir(root, "safe-box")
	if err != nil {
		t.Fatalf("resolveBoxProfileDir: %v", err)
	}
	want := filepath.Join(root, "safe-box")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if _, err := resolveBoxProfileDir(root, ".."); err == nil {
		t.Fatal("expected error for ..")
	}
	if _, err := resolveBoxProfileDir(root, "."); err == nil {
		t.Fatal("expected error for .")
	}
}

func TestDeleteBox_RejectsPathTraversal(t *testing.T) {
	home := t.TempDir()
	root := filepath.Join(home, "containers")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(home, "do-not-delete")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := DeleteBox(context.Background(), "..", root); err == nil {
		t.Fatal("DeleteBox(..) expected error")
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("sentinel deleted by traversal: %v", err)
	}
	if err := DeleteBox(context.Background(), ".", root); err == nil {
		t.Fatal("DeleteBox(.) expected error")
	}
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("containers root deleted by '.': %v", err)
	}
}

func TestDeleteCodexBox_RejectsPathTraversal(t *testing.T) {
	home := t.TempDir()
	root := filepath.Join(home, "codex-containers")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(home, "keep-me")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := DeleteCodexBox(context.Background(), "..", root); err == nil {
		t.Fatal("DeleteCodexBox(..) expected error")
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("sentinel deleted by traversal: %v", err)
	}
}

func TestCreateBox_RejectsDotDot(t *testing.T) {
	dir := t.TempDir()
	if _, err := CreateBox(context.Background(), "..", dir); err == nil {
		t.Fatal("CreateBox(..) expected error")
	}
	if _, err := CreateCodexBox(context.Background(), ".", dir); err == nil {
		t.Fatal("CreateCodexBox(.) expected error")
	}
}
