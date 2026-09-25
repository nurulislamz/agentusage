package boxes

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateCodexBox_SuccessAndValidation(t *testing.T) {
	dir := t.TempDir()

	// 1. Valid creation
	createdPath, err := CreateCodexBox(context.Background(), "work-box", dir)
	if err != nil {
		t.Fatalf("CreateCodexBox failed: %v", err)
	}
	expectedPath := filepath.Join(dir, "work-box")
	if createdPath != expectedPath {
		t.Errorf("createdPath = %q, want %q", createdPath, expectedPath)
	}

	codexDir := filepath.Join(expectedPath, ".codex")
	if !dirExists(codexDir) {
		t.Errorf(".codex directory missing in created box")
	}

	// 2. Duplicate rejection
	if _, err := CreateCodexBox(context.Background(), "work-box", dir); err == nil {
		t.Errorf("expected error creating duplicate box")
	}

	// 3. Validation
	if _, err := CreateCodexBox(context.Background(), "", dir); err == nil {
		t.Errorf("expected error for empty name")
	}
	if _, err := CreateCodexBox(context.Background(), "-invalid", dir); err == nil {
		t.Errorf("expected error for name starting with -")
	}
	if _, err := CreateCodexBox(context.Background(), "invalid/slash", dir); err == nil {
		t.Errorf("expected error for name containing slash")
	}
}

func TestDeleteCodexBox(t *testing.T) {
	dir := t.TempDir()

	_, err := CreateCodexBox(context.Background(), "temp-box", dir)
	if err != nil {
		t.Fatalf("CreateCodexBox failed: %v", err)
	}

	if err := DeleteCodexBox(context.Background(), "temp-box", dir); err != nil {
		t.Fatalf("DeleteCodexBox failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "temp-box")); !os.IsNotExist(err) {
		t.Errorf("expected temp-box to be removed")
	}

	// Deleting non-existent box returns error
	if err := DeleteCodexBox(context.Background(), "temp-box", dir); err == nil {
		t.Errorf("expected error deleting nonexistent box")
	}
}

func TestListCodexBoxes(t *testing.T) {
	dir := t.TempDir()

	// Empty list
	boxes, err := ListCodexBoxes(context.Background(), dir)
	if err != nil {
		t.Fatalf("ListCodexBoxes on empty dir failed: %v", err)
	}
	if len(boxes) != 0 {
		t.Errorf("expected 0 boxes, got %d", len(boxes))
	}

	// Create boxes
	_, err = CreateCodexBox(context.Background(), "b-box", dir)
	if err != nil {
		t.Fatalf("CreateCodexBox b-box failed: %v", err)
	}
	_, err = CreateCodexBox(context.Background(), "a-box", dir)
	if err != nil {
		t.Fatalf("CreateCodexBox a-box failed: %v", err)
	}

	// Mark a-box as authenticated
	_ = os.WriteFile(filepath.Join(dir, "a-box", ".codex", "auth.json"), []byte("{}"), 0o600)

	boxes, err = ListCodexBoxes(context.Background(), dir)
	if err != nil {
		t.Fatalf("ListCodexBoxes failed: %v", err)
	}
	if len(boxes) != 2 {
		t.Fatalf("expected 2 boxes, got %d", len(boxes))
	}

	// Alphabetically sorted: a-box then b-box
	if boxes[0].Name != "a-box" || boxes[0].Status != CodexStatusAuthenticated {
		t.Errorf("boxes[0] = %+v, want a-box Authenticated", boxes[0])
	}
	if boxes[1].Name != "b-box" || boxes[1].Status != CodexStatusInitialized {
		t.Errorf("boxes[1] = %+v, want b-box Initialized", boxes[1])
	}
}
