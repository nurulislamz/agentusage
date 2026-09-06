//go:build windows

package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

func daemonLockPath(socketPath string) string {
	return strings.TrimSpace(socketPath) + ".lock"
}

// acquireDaemonLock takes an exclusive non-blocking lock so only one daemon
// startup can open the telemetry store at a time. The lock must be held for
// the lifetime of the daemon (or until releaseDaemonLock).
func acquireDaemonLock(socketPath string) (*os.File, error) {
	lockPath := daemonLockPath(socketPath)
	if lockPath == ".lock" {
		return nil, fmt.Errorf("telemetry daemon socket path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, fmt.Errorf("create telemetry daemon lock dir: %w", err)
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open telemetry daemon lock: %w", err)
	}
	var ol windows.Overlapped
	err = windows.LockFileEx(
		windows.Handle(f.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0,
		1,
		0,
		&ol,
	)
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("telemetry daemon already starting or running")
	}
	return f, nil
}

func releaseDaemonLock(f *os.File) {
	if f == nil {
		return
	}
	var ol windows.Overlapped
	_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &ol)
	_ = f.Close()
}
