package daemon

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

// TestMain keeps tests away from the real telemetry database and service unit: a
// spawned daemon opening the database alongside the live one corrupts it. Tests
// that exercise spawning install their own spawnDaemonFunc.
func TestMain(m *testing.M) {
	stateDir, err := os.MkdirTemp("", "agentusage-daemon-test-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Setenv("XDG_STATE_HOME", stateDir)
	os.Setenv("HOME", stateDir)
	spawnDaemonFunc = func(string, bool) error {
		return errors.New("daemon spawn disabled in tests")
	}
	code := m.Run()
	os.RemoveAll(stateDir)
	os.Exit(code)
}
