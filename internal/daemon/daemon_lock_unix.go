//go:build unix

package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("telemetry daemon already starting or running")
	}
	return f, nil
}

func releaseDaemonLock(f *os.File) {
	if f == nil {
		return
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	_ = f.Close()
}
