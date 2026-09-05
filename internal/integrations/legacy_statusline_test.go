package integrations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCleanLegacyStatusline_NoConfig(t *testing.T) {
	nonexistent := filepath.Join(t.TempDir(), "settings.json")
	res := CleanLegacyStatuslineFile(nonexistent, "antigravity")
	if res.Outcome != OutcomeNotFound {
		t.Fatalf("res.Outcome = %v, want %v", res.Outcome, OutcomeNotFound)
	}
	if res.Backup != "" {
		t.Fatalf("expected no backup, got %q", res.Backup)
	}
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
}

func TestCleanLegacyStatusline_OwnedCommand(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	content := `{
  "customSetting": "preserved",
  "statusLine": {
    "command": "openusage antigravity statusline"
  }
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	res := CleanLegacyStatuslineFile(path, "antigravity")
	if res.Outcome != OutcomeCleaned {
		t.Fatalf("res.Outcome = %v, want %v (err: %v)", res.Outcome, OutcomeCleaned, res.Err)
	}
	if res.Backup != path+".bak" {
		t.Fatalf("res.Backup = %q, want %q", res.Backup, path+".bak")
	}

	// Verify backup file exists and contains original content
	bakData, err := os.ReadFile(res.Backup)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(bakData) != content {
		t.Fatalf("backup content mismatch:\n%s", string(bakData))
	}

	// Verify mutated file no longer has statusLine but preserved customSetting
	updatedData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read updated: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(updatedData, &parsed); err != nil {
		t.Fatalf("parse updated JSON: %v", err)
	}
	if _, exists := parsed["statusLine"]; exists {
		t.Fatalf("statusLine still exists in updated file: %+v", parsed)
	}
	if parsed["customSetting"] != "preserved" {
		t.Fatalf("customSetting not preserved: %+v", parsed)
	}
}

func TestCleanLegacyStatusline_QuotedExecutablePath(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "cli-config.json")
	content := `{
  "key": 12345,
  "statusLine": {
    "type": "command",
    "command": "\"/usr/local/bin/agentusage\" cursor statusline"
  }
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	res := CleanLegacyStatuslineFile(path, "cursor")
	if res.Outcome != OutcomeCleaned {
		t.Fatalf("res.Outcome = %v, want %v (err: %v)", res.Outcome, OutcomeCleaned, res.Err)
	}

	// Check backup and updated file
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("expected backup file: %v", err)
	}

	updatedData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read updated: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(updatedData, &parsed); err != nil {
		t.Fatalf("parse updated: %v", err)
	}
	if _, exists := parsed["statusLine"]; exists {
		t.Fatalf("statusLine not removed: %+v", parsed)
	}
	if val, ok := parsed["key"].(float64); !ok || val != 12345 {
		t.Fatalf("key not preserved: %+v", parsed)
	}
}

func TestCleanLegacyStatusline_CustomCommand(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	content := `{
  "statusLine": {
    "command": "my-custom-tool antigravity statusline"
  },
  "other": true
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	res := CleanLegacyStatuslineFile(path, "antigravity")
	if res.Outcome != OutcomeSkippedCustom {
		t.Fatalf("res.Outcome = %v, want %v", res.Outcome, OutcomeSkippedCustom)
	}
	if res.Backup != "" {
		t.Fatalf("expected no backup for skipped file, got %q", res.Backup)
	}

	// File content must be byte-exact identical
	afterData, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(afterData) != content {
		t.Fatalf("custom file was mutated:\n%s", string(afterData))
	}
	if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
		t.Fatal("backup should not be created for custom command")
	}
}

func TestCleanLegacyStatusline_MalformedJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	content := `{ malformed json: not valid`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	res := CleanLegacyStatuslineFile(path, "antigravity")
	if res.Outcome != OutcomeError {
		t.Fatalf("res.Outcome = %v, want %v", res.Outcome, OutcomeError)
	}
	if res.Err == nil {
		t.Fatal("expected per-file error for malformed JSON, got nil")
	}
	if res.Backup != "" {
		t.Fatalf("expected no backup for malformed file, got %q", res.Backup)
	}

	// File content must be untouched
	afterData, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(afterData) != content {
		t.Fatalf("malformed file was altered:\n%s", string(afterData))
	}
	if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
		t.Fatal("backup should not be created for malformed file")
	}
}

func TestCleanLegacyStatusline_MultipleBoxProfiles(t *testing.T) {
	home := t.TempDir()
	dirs := Dirs{
		Home:       home,
		ConfigRoot: filepath.Join(home, ".config"),
		HooksDir:   filepath.Join(home, ".config", "agentusage", "hooks"),
	}

	// 1. Antigravity host
	hostAgy := filepath.Join(home, ".gemini", "antigravity-cli", "settings.json")
	_ = os.MkdirAll(filepath.Dir(hostAgy), 0o755)
	_ = os.WriteFile(hostAgy, []byte(`{"statusLine":{"command":"openusage antigravity statusline"}}`), 0o644)

	// 2. Antigravity boxes
	box1Agy := filepath.Join(home, ".agy-containers", "dev1", ".gemini", "antigravity-cli", "settings.json")
	box2Agy := filepath.Join(home, ".agy-containers", "dev2", ".gemini", "antigravity-cli", "settings.json")
	_ = os.MkdirAll(filepath.Dir(box1Agy), 0o755)
	_ = os.MkdirAll(filepath.Dir(box2Agy), 0o755)
	_ = os.WriteFile(box1Agy, []byte(`{"statusLine":{"command":"agentusage antigravity statusline"}}`), 0o644)
	_ = os.WriteFile(box2Agy, []byte(`{"statusLine":{"command":"\"/opt/agentusage\" antigravity statusline"}}`), 0o644)

	// 3. Cursor host
	hostCursor := filepath.Join(home, ".cursor", "cli-config.json")
	_ = os.MkdirAll(filepath.Dir(hostCursor), 0o755)
	_ = os.WriteFile(hostCursor, []byte(`{"statusLine":{"type":"command","command":"agentusage cursor statusline"}}`), 0o644)

	// 4. Cursor boxes (.agent-containers and .cursor-containers)
	box1Cursor := filepath.Join(home, ".agent-containers", "agentbox", ".cursor", "cli-config.json")
	box2Cursor := filepath.Join(home, ".cursor-containers", "cursorbox", ".cursor", "cli-config.json")
	_ = os.MkdirAll(filepath.Dir(box1Cursor), 0o755)
	_ = os.MkdirAll(filepath.Dir(box2Cursor), 0o755)
	_ = os.WriteFile(box1Cursor, []byte(`{"statusLine":{"command":"openusage cursor statusline"}}`), 0o644)
	_ = os.WriteFile(box2Cursor, []byte(`{"statusLine":{"command":"openusage cursor statusline"}}`), 0o644)

	results := CleanLegacyStatusLines(dirs)

	// Check that all 6 files were cleaned
	cleanedCount := 0
	for _, res := range results {
		if res.Outcome == OutcomeCleaned {
			cleanedCount++
		}
	}
	if cleanedCount != 6 {
		t.Fatalf("expected 6 cleaned files, got %d. Results: %+v", cleanedCount, results)
	}
}

func TestCleanLegacyStatusline_StaleStatusFileOnly(t *testing.T) {
	home := t.TempDir()
	dirs := Dirs{
		Home:       home,
		ConfigRoot: filepath.Join(home, ".config"),
		HooksDir:   filepath.Join(home, ".config", "agentusage", "hooks"),
	}

	// Create only stale status files in ~/.local/state/agentusage
	stateDir := filepath.Join(home, ".local", "state", "agentusage")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "antigravity-status.json"), []byte(`{"quota":100}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "cursor-status.json"), []byte(`{"quota":100}`), 0o644); err != nil {
		t.Fatal(err)
	}

	results := CleanLegacyStatusLines(dirs)
	for _, res := range results {
		if res.Outcome == OutcomeCleaned {
			t.Errorf("expected no clean operations when configs do not exist, got %v for %s", res.Outcome, res.Path)
		}
	}

	// Stale status files must be untouched
	if _, err := os.Stat(filepath.Join(stateDir, "antigravity-status.json")); err != nil {
		t.Fatalf("stale status file removed or missing: %v", err)
	}
}

func TestCleanLegacyStatusline_RepeatedCleanup(t *testing.T) {
	home := t.TempDir()
	dirs := Dirs{
		Home:       home,
		ConfigRoot: filepath.Join(home, ".config"),
		HooksDir:   filepath.Join(home, ".config", "agentusage", "hooks"),
	}

	settingsPath := filepath.Join(home, ".gemini", "antigravity-cli", "settings.json")
	_ = os.MkdirAll(filepath.Dir(settingsPath), 0o755)
	initialContent := `{"statusLine":{"command":"openusage antigravity statusline"},"user":"test"}`
	_ = os.WriteFile(settingsPath, []byte(initialContent), 0o644)

	// First cleanup pass
	firstResults := CleanLegacyStatusLines(dirs)
	cleaned := false
	for _, r := range firstResults {
		if r.Path == settingsPath && r.Outcome == OutcomeCleaned {
			cleaned = true
		}
	}
	if !cleaned {
		t.Fatalf("first pass did not clean %s: %+v", settingsPath, firstResults)
	}

	bakStat1, err := os.Stat(settingsPath + ".bak")
	if err != nil {
		t.Fatalf("backup missing after first pass: %v", err)
	}

	// Second cleanup pass (must be no-op)
	secondResults := CleanLegacyStatusLines(dirs)
	for _, r := range secondResults {
		if r.Path == settingsPath {
			if r.Outcome != OutcomeNoStatusline {
				t.Fatalf("second pass outcome = %v, want %v", r.Outcome, OutcomeNoStatusline)
			}
			if r.Backup != "" {
				t.Fatalf("second pass created backup: %q", r.Backup)
			}
		}
	}

	// Backup modtime and contents must be unchanged
	bakStat2, err := os.Stat(settingsPath + ".bak")
	if err != nil {
		t.Fatalf("backup missing after second pass: %v", err)
	}
	if bakStat1.ModTime() != bakStat2.ModTime() {
		t.Fatal("backup file was rewritten during second pass")
	}
}
