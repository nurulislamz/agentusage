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
}
