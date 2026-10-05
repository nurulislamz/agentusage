//go:build windows

package telemetry

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func storeOpenLockPath(dbPath string) string {
	return dbPath + ".openlock"
}

func lockStoreShared(dbPath string) (*os.File, error) {
	lockPath := storeOpenLockPath(dbPath)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, fmt.Errorf("telemetry: create store lock dir: %w", err)
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("telemetry: open store lock: %w", err)
	}
	var ol windows.Overlapped
	err = windows.LockFileEx(
		windows.Handle(f.Fd()),
		windows.LOCKFILE_FAIL_IMMEDIATELY, // shared when exclusive bit unset
		0,
		1,
		0,
		&ol,
	)
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("telemetry: shared store lock: %w", err)
	}
	return f, nil
}

func tryLockStoreExclusive(f *os.File) bool {
	if f == nil {
		return false
	}
	var ol windows.Overlapped
	// Unlock shared first, then try exclusive. Windows LockFileEx does not
	// upgrade in place the way flock(2) does.
	_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &ol)
	err := windows.LockFileEx(
		windows.Handle(f.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0,
		1,
		0,
		&ol,
	)
	if err != nil {
		// Re-acquire shared so the Store still holds a presence lock.
		_ = windows.LockFileEx(
			windows.Handle(f.Fd()),
			windows.LOCKFILE_FAIL_IMMEDIATELY,
			0,
			1,
			0,
			&ol,
		)
		return false
	}
	return true
}

func downgradeStoreLockShared(f *os.File) {
	if f == nil {
		return
	}
	var ol windows.Overlapped
	_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &ol)
	_ = windows.LockFileEx(
		windows.Handle(f.Fd()),
		windows.LOCKFILE_FAIL_IMMEDIATELY,
		0,
		1,
		0,
		&ol,
	)
}

func unlockStore(f *os.File) {
	if f == nil {
		return
	}
	var ol windows.Overlapped
	_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &ol)
	_ = f.Close()
}
