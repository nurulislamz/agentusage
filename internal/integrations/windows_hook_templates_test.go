package integrations

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Cross-platform guard: Windows Claude/Codex hooks must invoke the live
// `daemon hook` CLI. The telemetry subcommand was removed from the public
// surface; a hidden alias remains only for already-installed hooks.
func TestWindowsHookTemplatesUseDaemonHook(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	tpl := filepath.Join(filepath.Dir(thisFile), "assets", "claude-hook.cmd.tpl")
	data, err := os.ReadFile(tpl)
	if err != nil {
		t.Fatalf("read %s: %v", tpl, err)
	}
	content := string(data)
	if !strings.Contains(content, "daemon hook claude_code") {
		t.Fatalf("claude-hook.cmd.tpl must invoke daemon hook, got:\n%s", content)
	}
	if strings.Contains(content, "telemetry hook") {
		t.Fatalf("claude-hook.cmd.tpl still invokes removed telemetry hook:\n%s", content)
	}
}
