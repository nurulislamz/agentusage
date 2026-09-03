package boxes

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateBox_SuccessAndValidation(t *testing.T) {
	dir := t.TempDir()

	// 1. Valid creation
	createdPath, err := CreateBox(context.Background(), "work-box", dir)
	if err != nil {
		t.Fatalf("CreateBox failed: %v", err)
	}
	expectedPath := filepath.Join(dir, "work-box")
	if createdPath != expectedPath {
		t.Errorf("createdPath = %q, want %q", createdPath, expectedPath)
	}

	settingsPath := filepath.Join(expectedPath, ".gemini", "antigravity-cli", "settings.json")
	if !fileExists(settingsPath) {
		t.Errorf("settings.json missing in created box")
	}

	// 2. Duplicate rejection
	if _, err := CreateBox(context.Background(), "work-box", dir); err == nil {
		t.Errorf("expected error creating duplicate box")
	}

	// 3. Validation
	if _, err := CreateBox(context.Background(), "", dir); err == nil {
		t.Errorf("expected error for empty name")
	}
	if _, err := CreateBox(context.Background(), "-invalid", dir); err == nil {
		t.Errorf("expected error for name starting with -")
	}
	if _, err := CreateBox(context.Background(), "invalid/slash", dir); err == nil {
		t.Errorf("expected error for name containing slash")
	}
}

func TestDeleteBox(t *testing.T) {
	dir := t.TempDir()

	_, err := CreateBox(context.Background(), "temp-box", dir)
	if err != nil {
		t.Fatalf("CreateBox failed: %v", err)
	}

	if err := DeleteBox(context.Background(), "temp-box", dir); err != nil {
		t.Fatalf("DeleteBox failed: %v", err)
	}

	if fileExists(filepath.Join(dir, "temp-box")) {
		t.Errorf("box directory still exists after delete")
	}

	// Deleting non-existent box should error
	if err := DeleteBox(context.Background(), "non-existent", dir); err == nil {
		t.Errorf("expected error deleting non-existent box")
	}
}

func TestListBoxes_SortingAndStatus(t *testing.T) {
	dir := t.TempDir()

	// Box 1: Initialized (no token)
	_, _ = CreateBox(context.Background(), "zebra", dir)

	// Box 2: Ready (with valid token)
	_, _ = CreateBox(context.Background(), "alpha", dir)
	alphaTokenDir := filepath.Join(dir, "alpha", ".gemini", "antigravity-cli")
	validToken := map[string]interface{}{
		"token": map[string]string{
			"access_token": "ya29.mock_valid_token",
		},
	}
	tokenData, _ := json.Marshal(validToken)
	_ = os.WriteFile(filepath.Join(alphaTokenDir, "antigravity-oauth-token"), tokenData, 0o600)

	// Box 3: Authenticated (legacy oauth file)
	_, _ = CreateBox(context.Background(), "beta", dir)
	betaGeminiDir := filepath.Join(dir, "beta", ".gemini")
	_ = os.WriteFile(filepath.Join(betaGeminiDir, "oauth_creds.json"), []byte("{}"), 0o600)

	boxes, err := ListBoxes(context.Background(), dir)
	if err != nil {
		t.Fatalf("ListBoxes failed: %v", err)
	}

	if len(boxes) != 3 {
		t.Fatalf("expected 3 boxes, got %d", len(boxes))
	}

	// Test alphabetical order
	if boxes[0].Name != "alpha" || boxes[1].Name != "beta" || boxes[2].Name != "zebra" {
		t.Errorf("unexpected box ordering: %v, %v, %v", boxes[0].Name, boxes[1].Name, boxes[2].Name)
	}

	// Test statuses
	if boxes[0].Status != StatusReady {
		t.Errorf("alpha status = %v, want %v", boxes[0].Status, StatusReady)
	}
	if boxes[1].Status != StatusAuthenticated {
		t.Errorf("beta status = %v, want %v", boxes[1].Status, StatusAuthenticated)
	}
	if boxes[2].Status != StatusInitialized {
		t.Errorf("zebra status = %v, want %v", boxes[2].Status, StatusInitialized)
	}
}

func TestLoginBoxSession_SuccessWithURLAndToken(t *testing.T) {
	dir := t.TempDir()
	_, _ = CreateBox(context.Background(), "login-box", dir)

	mockAuthURL := "https://accounts.google.com/o/oauth2/auth?client_id=123.apps.googleusercontent.com&redirect_uri=http://localhost:8085"
	simulatedOutput := "Starting Antigravity CLI in box login-box...\nPlease authenticate by visiting:\n" + mockAuthURL + "\nWaiting for authentication...\n"

	var capturedURL string
	var openedURL string
	tokenSavedCalled := false

	mockRunner := func(ctx context.Context, box string, args ...string) (io.ReadCloser, func(), error) {
		rc := io.NopCloser(bytes.NewBufferString(simulatedOutput))
		cancel := func() {}

		// Simulate user completing OAuth after 100ms
		go func() {
			time.Sleep(100 * time.Millisecond)
			tokenPath := filepath.Join(dir, box, ".gemini", "antigravity-cli", "antigravity-oauth-token")
			payload := map[string]interface{}{
				"token": map[string]string{
					"access_token": "ya29.simulated_access_token",
				},
			}
			data, _ := json.Marshal(payload)
			_ = os.WriteFile(tokenPath, data, 0o600)
		}()

		return rc, cancel, nil
	}

	opts := LoginOptions{
		BaseDir: dir,
		Runner:  mockRunner,
		BrowserOpener: func(u string) error {
			openedURL = u
			return nil
		},
		OnAuthURL: func(u string) {
			capturedURL = u
		},
		OnTokenSaved: func() {
			tokenSavedCalled = true
		},
		PollInterval: 50 * time.Millisecond,
		Timeout:      2 * time.Second,
	}

	err := LoginBoxSession(context.Background(), "login-box", opts)
	if err != nil {
		t.Fatalf("LoginBoxSession failed: %v", err)
	}

	if capturedURL != mockAuthURL {
		t.Errorf("capturedURL = %q, want %q", capturedURL, mockAuthURL)
	}
	if openedURL != mockAuthURL {
		t.Errorf("openedURL = %q, want %q", openedURL, mockAuthURL)
	}
	if !tokenSavedCalled {
		t.Errorf("expected OnTokenSaved callback to be called")
	}
}

func TestLoginBoxSession_Timeout(t *testing.T) {
	dir := t.TempDir()
	_, _ = CreateBox(context.Background(), "timeout-box", dir)

	mockRunner := func(ctx context.Context, box string, args ...string) (io.ReadCloser, func(), error) {
		rc := io.NopCloser(bytes.NewBufferString("Running without output..."))
		return rc, func() {}, nil
	}

	opts := LoginOptions{
		BaseDir:      dir,
		Runner:       mockRunner,
		PollInterval: 20 * time.Millisecond,
		Timeout:      100 * time.Millisecond,
	}

	err := LoginBoxSession(context.Background(), "timeout-box", opts)
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
}
