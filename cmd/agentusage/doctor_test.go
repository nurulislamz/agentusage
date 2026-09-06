package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDoctorCommandExecution(t *testing.T) {
	var buf bytes.Buffer
	runDoctorDiagnostics(&buf, false)
	output := buf.String()

	if !strings.Contains(output, "agentUsage Doctor") {
		t.Errorf("expected header in doctor output, got:\n%s", output)
	}
	if !strings.Contains(output, "[ OK ] System:") {
		t.Errorf("expected system check in doctor output, got:\n%s", output)
	}
	if !strings.Contains(output, "Result:") {
		t.Errorf("expected result summary in doctor output, got:\n%s", output)
	}
}

func TestDoctorCheckerFormatting(t *testing.T) {
	var buf bytes.Buffer
	d := &doctorChecker{out: &buf}

	d.ok("check 1: %s", "passed")
	d.info("check 2: %s", "informational")
	d.warn("check 3: %s", "warning")
	d.fail("check 4: %s", "failed")

	if d.okCount != 1 {
		t.Errorf("expected okCount=1, got %d", d.okCount)
	}
	if d.warnCount != 1 {
		t.Errorf("expected warnCount=1, got %d", d.warnCount)
	}
	if d.failCount != 1 {
		t.Errorf("expected failCount=1, got %d", d.failCount)
	}

	out := buf.String()
	if !strings.Contains(out, "[ OK ] check 1: passed") {
		t.Errorf("missing [ OK ] format: %s", out)
	}
	if !strings.Contains(out, "[INFO] check 2: informational") {
		t.Errorf("missing [INFO] format: %s", out)
	}
	if !strings.Contains(out, "[WARN] check 3: warning") {
		t.Errorf("missing [WARN] format: %s", out)
	}
	if !strings.Contains(out, "[FAIL] check 4: failed") {
		t.Errorf("missing [FAIL] format: %s", out)
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{5 * 1024 * 1024, "5.0 MB"},
	}

	for _, tt := range tests {
		got := formatFileSize(tt.bytes)
		if got != tt.expected {
			t.Errorf("formatFileSize(%d) = %q, expected %q", tt.bytes, got, tt.expected)
		}
	}
}

func TestDoctorDiagnostics_DoesNotCheckStatuslineOrTmux(t *testing.T) {
	var buf bytes.Buffer
	runDoctorDiagnostics(&buf, false)
	output := buf.String()

	if strings.Contains(output, "Statusline") {
		t.Errorf("doctor output should not contain Statusline check, got:\n%s", output)
	}
	if strings.Contains(output, "tmux Integration") {
		t.Errorf("doctor output should not contain tmux Integration check, got:\n%s", output)
	}
}

func TestDoctorCommand_DetectFlag(t *testing.T) {
	cmd := newDoctorCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--detect"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error running doctor --detect: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Tools detected:") {
		t.Errorf("expected 'Tools detected:' in doctor --detect output, got:\n%s", out)
	}
	if !strings.Contains(out, "Accounts detected:") {
		t.Errorf("expected 'Accounts detected:' in doctor --detect output, got:\n%s", out)
	}
}

func TestDoctorCommand_FixLegacyStatuslines(t *testing.T) {
	tmp := t.TempDir()

	// Setup fake Antigravity config with owned statusline
	agyDir := tmp + "/agy"
	if err := os.MkdirAll(agyDir, 0o700); err != nil {
		t.Fatal(err)
	}
	agyFile := agyDir + "/settings.json"
	agyJSON := `{
  "statusLine": {
    "command": "openusage antigravity statusline",
    "priority": 1
  },
  "otherSetting": true
}`
	if err := os.WriteFile(agyFile, []byte(agyJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ANTIGRAVITY_SETTINGS_FILE", agyFile)

	// Setup fake Cursor config with custom statusline
	cursorDir := tmp + "/cursor"
	if err := os.MkdirAll(cursorDir, 0o700); err != nil {
		t.Fatal(err)
	}
	cursorFile := cursorDir + "/cli-config.json"
	cursorJSON := `{
  "statusLine": {
    "command": "/usr/local/bin/my-custom-statusline --verbose"
  },
  "cursorSetting": "abc"
}`
	if err := os.WriteFile(cursorFile, []byte(cursorJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURSOR_SETTINGS_FILE", cursorFile)

	// Run doctor --fix-legacy-statuslines
	cmd := newDoctorCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--fix-legacy-statuslines"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error running doctor --fix-legacy-statuslines: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "[CLEANED]") || !strings.Contains(out, agyFile) {
		t.Errorf("expected agyFile to be cleaned, output:\n%s", out)
	}
	if !strings.Contains(out, "[SKIPPED]") || !strings.Contains(out, cursorFile) {
		t.Errorf("expected cursorFile to be skipped (custom), output:\n%s", out)
	}

	// Verify backup was created for agyFile
	backupData, err := os.ReadFile(agyFile + ".bak")
	if err != nil {
		t.Fatalf("backup file not created: %v", err)
	}
	if !strings.Contains(string(backupData), "openusage antigravity statusline") {
		t.Errorf("backup file missing original content: %s", string(backupData))
	}

	// Verify agyFile no longer has statusLine but preserved otherSetting
	cleanedData, err := os.ReadFile(agyFile)
	if err != nil {
		t.Fatalf("reading cleaned file: %v", err)
	}
	if strings.Contains(string(cleanedData), "statusLine") {
		t.Errorf("cleaned file still contains statusLine: %s", string(cleanedData))
	}
	if !strings.Contains(string(cleanedData), "otherSetting") {
		t.Errorf("cleaned file lost unrelated settings: %s", string(cleanedData))
	}

	// Verify cursorFile still has its custom statusLine
	cursorData, err := os.ReadFile(cursorFile)
	if err != nil {
		t.Fatalf("reading cursor file: %v", err)
	}
	if !strings.Contains(string(cursorData), "my-custom-statusline") {
		t.Errorf("custom statusline was not preserved: %s", string(cursorData))
	}

	// Run migration a second time - should be a no-op for agyFile
	var buf2 bytes.Buffer
	cmd2 := newDoctorCommand()
	cmd2.SetOut(&buf2)
	cmd2.SetArgs([]string{"--fix-legacy-statuslines"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	out2 := buf2.String()
	if strings.Contains(out2, "[CLEANED]") {
		t.Errorf("second run should not clean any files, output:\n%s", out2)
	}
}
