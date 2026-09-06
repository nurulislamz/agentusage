package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSaveAndLoadCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")

	if err := SaveCredentialTo(path, "openai-personal", "sk-test-key-123"); err != nil {
		t.Fatalf("SaveCredentialTo error: %v", err)
	}
	if err := SaveCredentialTo(path, "anthropic-work", "sk-ant-456"); err != nil {
		t.Fatalf("SaveCredentialTo error: %v", err)
	}

	creds, err := LoadCredentialsFrom(path)
	if err != nil {
		t.Fatalf("LoadCredentialsFrom error: %v", err)
	}

	if len(creds.Keys) != 2 {
		t.Fatalf("keys count = %d, want 2", len(creds.Keys))
	}
	if creds.Keys["openai-personal"] != "sk-test-key-123" {
		t.Errorf("openai key = %q, want sk-test-key-123", creds.Keys["openai-personal"])
	}
	if creds.Keys["anthropic-work"] != "sk-ant-456" {
		t.Errorf("anthropic key = %q, want sk-ant-456", creds.Keys["anthropic-work"])
	}
}

func TestDeleteCredential(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")

	if err := SaveCredentialTo(path, "openai-personal", "sk-test-key-123"); err != nil {
		t.Fatal(err)
	}
	if err := SaveCredentialTo(path, "anthropic-work", "sk-ant-456"); err != nil {
		t.Fatal(err)
	}

	if err := DeleteCredentialFrom(path, "openai-personal"); err != nil {
		t.Fatalf("DeleteCredentialFrom error: %v", err)
	}

	creds, err := LoadCredentialsFrom(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(creds.Keys) != 1 {
		t.Fatalf("keys count = %d, want 1", len(creds.Keys))
	}
	if _, ok := creds.Keys["openai-personal"]; ok {
		t.Error("openai-personal should have been deleted")
	}
	if creds.Keys["anthropic-work"] != "sk-ant-456" {
		t.Errorf("anthropic key = %q, want sk-ant-456", creds.Keys["anthropic-work"])
	}
}

func TestLoadCredentials_FileNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent", "credentials.json")

	creds, err := LoadCredentialsFrom(path)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if creds.Keys == nil {
		t.Fatal("expected non-nil Keys map")
	}
	if len(creds.Keys) != 0 {
		t.Errorf("expected empty keys, got %d", len(creds.Keys))
	}
}

func TestSaveCredential_CreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "deep", "dir")
	path := filepath.Join(dir, "credentials.json")

	if err := SaveCredentialTo(path, "test-acct", "sk-key-789"); err != nil {
		t.Fatalf("SaveCredentialTo error: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("credentials file was not created")
	}

	creds, err := LoadCredentialsFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if creds.Keys["test-acct"] != "sk-key-789" {
		t.Errorf("key = %q, want sk-key-789", creds.Keys["test-acct"])
	}
}

func TestCredentialFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission test not applicable on Windows")
	}

	path := filepath.Join(t.TempDir(), "credentials.json")

	if err := SaveCredentialTo(path, "test-acct", "sk-secret"); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}

func TestSaveCredential_OverwritesExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")

	if err := SaveCredentialTo(path, "openai-personal", "sk-old-key"); err != nil {
		t.Fatal(err)
	}
	if err := SaveCredentialTo(path, "openai-personal", "sk-new-key"); err != nil {
		t.Fatal(err)
	}

	creds, err := LoadCredentialsFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if creds.Keys["openai-personal"] != "sk-new-key" {
		t.Errorf("key = %q, want sk-new-key", creds.Keys["openai-personal"])
	}
}

func TestLoadCredentialsFrom_PreservesAccountIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	content := `{
  "keys": {
    "openai-auto": "sk-old",
    "openai": "sk-new",
    "gemini-api-auto": "g1",
    "copilot-auto": "c1"
  }
}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	creds, err := LoadCredentialsFrom(path)
	if err != nil {
		t.Fatalf("LoadCredentialsFrom error: %v", err)
	}

	if len(creds.Keys) != 4 {
		t.Fatalf("keys count = %d, want 4", len(creds.Keys))
	}
	if creds.Keys["openai-auto"] != "sk-old" {
		t.Errorf("openai-auto key = %q, want sk-old", creds.Keys["openai-auto"])
	}
	if creds.Keys["openai"] != "sk-new" {
		t.Errorf("openai key = %q, want sk-new", creds.Keys["openai"])
	}
	if creds.Keys["gemini-api-auto"] != "g1" {
		t.Errorf("gemini-api-auto key = %q, want g1", creds.Keys["gemini-api-auto"])
	}
	if creds.Keys["copilot-auto"] != "c1" {
		t.Errorf("copilot-auto key = %q, want c1", creds.Keys["copilot-auto"])
	}
}

func TestDeleteCredentialFrom_RequiresExactAccountID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")

	if err := SaveCredentialTo(path, "openai-auto", "sk-test-key"); err != nil {
		t.Fatalf("SaveCredentialTo error: %v", err)
	}
	if err := DeleteCredentialFrom(path, "openai"); err != nil {
		t.Fatalf("DeleteCredentialFrom error: %v", err)
	}

	creds, err := LoadCredentialsFrom(path)
	if err != nil {
		t.Fatalf("LoadCredentialsFrom error: %v", err)
	}
	if len(creds.Keys) != 1 {
		t.Fatalf("keys count = %d, want 1", len(creds.Keys))
	}
	if got := creds.Keys["openai-auto"]; got != "sk-test-key" {
		t.Fatalf("openai-auto key = %q, want preserved", got)
	}
}

func TestSaveSession_PermissionsAndAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")

	session := BrowserSession{
		Domain:     "claude.ai",
		CookieName: "sessionKey",
		Value:      "sk-ant-session-secret",
	}

	if err := SaveSessionTo(path, "claude_code-personal", session); err != nil {
		t.Fatalf("SaveSessionTo: %v", err)
	}

	loaded, found, err := LoadSessionFrom(path, "claude_code-personal")
	if err != nil {
		t.Fatalf("LoadSessionFrom: %v", err)
	}
	if !found || loaded.Value != session.Value {
		t.Errorf("loaded session = %+v, want %+v", loaded, session)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("os.Stat: %v", err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("credentials.json permissions = %o, want 0600", perm)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "credentials.json" {
		t.Errorf("expected only credentials.json in directory, found: %v", entries)
	}
}

func TestSaveCredentialTo_CorruptFileDoesNotWipe(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	if err := os.WriteFile(path, []byte(`{"keys":{"keep-me":"sk-original"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	// Truncate mid-object so LoadCredentialsFrom fails to parse.
	if err := os.WriteFile(path, []byte(`{"keys":{"keep-me":"sk-original","other":`), 0o600); err != nil {
		t.Fatal(err)
	}

	err := SaveCredentialTo(path, "new-acct", "sk-new")
	if err == nil {
		t.Fatal("SaveCredentialTo on corrupt file: want error, got nil")
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile: %v", readErr)
	}
	if string(data) == "" || !strings.Contains(string(data), "keep-me") {
		t.Fatalf("corrupt credentials file was overwritten; contents=%q", data)
	}
}

func TestSaveSessionTo_CorruptFileDoesNotWipe(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	if err := os.WriteFile(path, []byte(`{"keys":{"keep-me":"sk-original"},"sessions":{`), 0o600); err != nil {
		t.Fatal(err)
	}

	err := SaveSessionTo(path, "sess-acct", BrowserSession{
		Domain:     "example.com",
		CookieName: "session",
		Value:      "cookie-value",
	})
	if err == nil {
		t.Fatal("SaveSessionTo on corrupt file: want error, got nil")
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile: %v", readErr)
	}
	if !strings.Contains(string(data), "keep-me") {
		t.Fatalf("corrupt credentials file was overwritten; contents=%q", data)
	}
}
