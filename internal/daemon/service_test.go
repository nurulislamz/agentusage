package daemon

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nurulislamz/agentusage/internal/telemetry"
)

func TestLastErrorLine_ReturnsMostRecentError(t *testing.T) {
	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "daemon.stderr.log")
	content := strings.Join([]string{
		"hint: missing integration",
		"Error: telemetry daemon already running on socket /tmp/agentusage.sock",
		"Usage:",
		"Error: open daemon telemetry store: telemetry: opening DB: permission denied",
		"tail line",
	}, "\n")
	if err := os.WriteFile(logPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	got := LastErrorLine(logPath)
	want := "open daemon telemetry store: telemetry: opening DB: permission denied"
	if got != want {
		t.Fatalf("LastErrorLine() = %q, want %q", got, want)
	}

	// Non-matching log
	noErrPath := filepath.Join(tmp, "clean.log")
	_ = os.WriteFile(noErrPath, []byte("all good\ninfo: started"), 0o644)
	if got := LastErrorLine(noErrPath); got != "" {
		t.Errorf("LastErrorLine() on clean log = %q, want empty", got)
	}

	// Missing file
	if got := LastErrorLine(filepath.Join(tmp, "missing.log")); got != "" {
		t.Errorf("LastErrorLine() on missing file = %q, want empty", got)
	}
}

func TestServiceManager_PathsAndHints(t *testing.T) {
	tempState := t.TempDir()

	kinds := []struct {
		kind        string
		wantSupport bool
		wantHintSub string
	}{
		{"darwin", true, "launchctl"},
		{"linux", true, "systemctl"},
		{"windows", true, "schtasks"},
		{"other", false, ""},
	}

	for _, tc := range kinds {
		t.Run(tc.kind, func(t *testing.T) {
			m := ServiceManager{
				Kind:       tc.kind,
				stateDir:   tempState,
				socketPath: "/tmp/test.sock",
			}
			if got := m.IsSupported(); got != tc.wantSupport {
				t.Errorf("IsSupported() = %v, want %v", got, tc.wantSupport)
			}
			hint := m.StatusHint()
			if tc.wantHintSub != "" && !strings.Contains(hint, tc.wantHintSub) {
				t.Errorf("StatusHint() = %q, want containing %q", hint, tc.wantHintSub)
			}
			if tc.wantHintSub == "" && hint != "" {
				t.Errorf("StatusHint() for unsupported = %q, want empty", hint)
			}
			if !strings.HasSuffix(m.StdoutLogPath(), "daemon.stdout.log") {
				t.Errorf("StdoutLogPath() = %q", m.StdoutLogPath())
			}
			if !strings.HasSuffix(m.StderrLogPath(), "daemon.stderr.log") {
				t.Errorf("StderrLogPath() = %q", m.StderrLogPath())
			}
			if !strings.HasSuffix(m.EnvFilePath(), "daemon.env") {
				t.Errorf("EnvFilePath() = %q", m.EnvFilePath())
			}
		})
	}

	// Empty stateDir returns empty paths
	mEmpty := ServiceManager{}
	if mEmpty.StdoutLogPath() != "" || mEmpty.StderrLogPath() != "" || mEmpty.EnvFilePath() != "" {
		t.Errorf("expected empty paths when stateDir is empty, got stdout=%q", mEmpty.StdoutLogPath())
	}
}

func TestServiceHelpers_YesNo_ValueOrNA(t *testing.T) {
	if got := yesNo(true); got != "yes" {
		t.Errorf("yesNo(true) = %q, want 'yes'", got)
	}
	if got := yesNo(false); got != "no" {
		t.Errorf("yesNo(false) = %q, want 'no'", got)
	}
	if got := valueOrNA(""); got != "n/a" {
		t.Errorf("valueOrNA(\"\") = %q, want 'n/a'", got)
	}
	if got := valueOrNA("  "); got != "n/a" {
		t.Errorf("valueOrNA(\"  \") = %q, want 'n/a'", got)
	}
	if got := valueOrNA("val"); got != "val" {
		t.Errorf("valueOrNA(\"val\") = %q, want 'val'", got)
	}
}

func TestIsTransientExecutablePath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"", true},
		{"   ", true},
		{"/tmp/go-build1234/exe/main", true},
		{"/usr/local/bin/agentusage", false},
		{"/home/user/bin/agentusage", false},
	}

	for _, tc := range cases {
		if got := isTransientExecutablePath(tc.path); got != tc.want {
			t.Errorf("isTransientExecutablePath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestParseLSOFFirstRecord(t *testing.T) {
	out := "p4321\nckooky\nn/tmp/agentusage.sock\n"
	got := parseLSOFFirstRecord(out)
	want := "pid=4321 command=kooky socket=/tmp/agentusage.sock"
	if got != want {
		t.Errorf("parseLSOFFirstRecord = %q, want %q", got, want)
	}

	// Partial output
	partial := "p9999\n"
	if got := parseLSOFFirstRecord(partial); got != "pid=9999" {
		t.Errorf("parseLSOFFirstRecord(partial) = %q, want 'pid=9999'", got)
	}

	// Empty output
	if got := parseLSOFFirstRecord(""); got != "" {
		t.Errorf("parseLSOFFirstRecord(\"\") = %q, want empty", got)
	}
}

func TestWriteServiceEnvFile_And_CurrentServiceEnvSnapshot(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, "daemon.env")

	envMap := map[string]string{
		"OPENAI_API_KEY":       "sk-test-12345",
		"AGENTUSAGE_DEBUG":     "1",
		"AGENTUSAGE_HUB_TOKEN": "hub-token-xyz",
	}

	if err := writeServiceEnvFile(envPath, envMap); err != nil {
		t.Fatalf("writeServiceEnvFile: %v", err)
	}

	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	sContent := string(content)
	if !strings.Contains(sContent, `OPENAI_API_KEY="sk-test-12345"`) {
		t.Errorf("missing OPENAI_API_KEY in env file: %s", sContent)
	}
	if !strings.Contains(sContent, `AGENTUSAGE_DEBUG="1"`) {
		t.Errorf("missing AGENTUSAGE_DEBUG in env file: %s", sContent)
	}

	// Calling with empty path is safe no-op
	if err := writeServiceEnvFile("", envMap); err != nil {
		t.Errorf("writeServiceEnvFile with empty path returned err: %v", err)
	}

	// currentServiceEnvSnapshot returns map of set vars
	t.Setenv("OPENAI_API_KEY", "snapshot-test")
	defer os.Unsetenv("OPENAI_API_KEY")
	snap := currentServiceEnvSnapshot()
	if snap["OPENAI_API_KEY"] != "snapshot-test" {
		t.Errorf("currentServiceEnvSnapshot() missing set key: %+v", snap)
	}
}

func TestServiceStatus_Diagnostics(t *testing.T) {
	ctx := context.Background()

	// 1. Not running / offline socket with details=false
	err := ServiceStatus(ctx, "/tmp/nonexistent_test.sock", false)
	if err != nil {
		t.Errorf("ServiceStatus(details=false) returned unexpected err: %v", err)
	}

	// 2. Not running / offline socket with details=true
	err = ServiceStatus(ctx, "/tmp/nonexistent_test.sock", true)
	if err != nil {
		t.Errorf("ServiceStatus(details=true) returned unexpected err: %v", err)
	}
}

func TestSocketOwnerSummary(t *testing.T) {
	// Empty socket path
	if got := SocketOwnerSummary(""); got != "" {
		t.Errorf("SocketOwnerSummary(\"\") = %q, want empty", got)
	}

	// Non-existent socket path
	if got := SocketOwnerSummary("/tmp/nonexistent_test_owner.sock"); got != "" {
		t.Errorf("SocketOwnerSummary(nonexistent) = %q, want empty", got)
	}

	// Existing file
	tmpFile := filepath.Join(t.TempDir(), "dummy.sock")
	_ = os.WriteFile(tmpFile, []byte(""), 0644)
	_ = SocketOwnerSummary(tmpFile)
}

func TestServiceManager_InstallHint_DomainCandidates_LoadServiceEnv(t *testing.T) {
	mgr := ServiceManager{
		Kind: "darwin",
	}
	if hint := mgr.InstallHint(); !strings.Contains(hint, "agentusage") {
		t.Errorf("InstallHint() = %q", hint)
	}
	if doms := mgr.domainCandidates(); len(doms) == 0 {
		t.Error("domainCandidates() should return candidates on darwin")
	}

	// LoadServiceEnv should parse daemon.env file if present
	stateDir, err := telemetry.DefaultStateDir()
	if err == nil && stateDir != "" {
		_ = os.MkdirAll(stateDir, 0755)
		envPath := filepath.Join(stateDir, "daemon.env")
		_ = os.WriteFile(envPath, []byte("# comment line\n\nTEST_AGENTUSAGE_ENV_TEST_VAR=\"test_val_123\"\nINVALID_LINE\n"), 0644)
		defer os.Remove(envPath)
		LoadServiceEnv()
		if os.Getenv("TEST_AGENTUSAGE_ENV_TEST_VAR") != "test_val_123" {
			t.Errorf("TEST_AGENTUSAGE_ENV_TEST_VAR = %q, want test_val_123", os.Getenv("TEST_AGENTUSAGE_ENV_TEST_VAR"))
		}
	}
}

func TestInstallService_And_RunCommand(t *testing.T) {
	// 1. RunCommand success
	out, err := RunCommand("echo", "agentusage-test")
	if err != nil || out != "agentusage-test" {
		t.Errorf("RunCommand(echo) = %q, err=%v", out, err)
	}

	// 2. RunCommand failure
	_, err = RunCommand("sh", "-c", "echo 'cmd error' >&2; exit 1")
	if err == nil {
		t.Error("expected error from failed command")
	}

	// 3. InstallService from test binary (which is transient) should return transient error
	err = InstallService("/tmp/test.sock")
	if err == nil {
		t.Error("expected error installing service from transient test binary")
	}

	// 4. UninstallService
	_ = UninstallService("/tmp/test.sock")
}
