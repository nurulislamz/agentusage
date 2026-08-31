//go:build linux

package daemon

import (
	"strings"
	"testing"
)

func TestSystemdUnit_UsesDaemonRunSubcommand(t *testing.T) {
	unit := systemdUnit("/usr/local/bin/agentusage", "/tmp/agentusage.sock", "/tmp/agentusage.env")

	if !strings.Contains(unit, "ExecStart=/usr/local/bin/agentusage telemetry daemon run --socket-path /tmp/agentusage.sock") {
		t.Fatalf("systemd unit does not include daemon run subcommand:\n%s", unit)
	}
	if !strings.Contains(unit, "EnvironmentFile=-/tmp/agentusage.env") {
		t.Fatalf("systemd unit does not include env file:\n%s", unit)
	}
	if !strings.Contains(unit, "Environment=PATH=%h/.local/bin:%h/bin:") {
		t.Fatalf("systemd unit does not prepend user-local bin to PATH:\n%s", unit)
	}
}

func TestServiceManager_Install_RejectsTransientExe(t *testing.T) {
	mgr := ServiceManager{
		Kind:    "linux",
		exePath: "/tmp/go-build12345/exe/agentusage",
	}

	err := mgr.Install()
	if err == nil {
		t.Fatal("expected error when installing service from transient executable")
	}
	if !strings.Contains(err.Error(), "refusing to install telemetry daemon service from transient executable") {
		t.Errorf("error = %q, want transient executable refusal message", err)
	}
}
