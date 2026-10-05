//go:build unix

package telemetry

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func storeOpenLockPath(dbPath string) string {
	return dbPath + ".openlock"
}

// lockStoreShared acquires a shared flock for the lifetime of an open Store.
// Destructive open recovery (unlink -shm / rename-on-corrupt) requires an
// exclusive upgrade; that fails while any other Store holds this shared lock.
func lockStoreShared(dbPath string) (*os.File, error) {
	lockPath := storeOpenLockPath(dbPath)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, fmt.Errorf("telemetry: create store lock dir: %w", err)
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("telemetry: open store lock: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("telemetry: shared store lock: %w", err)
	}
	return f, nil
}

func tryLockStoreExclusive(f *os.File) bool {
	if f == nil {
		return false
	}
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) == nil
}

func downgradeStoreLockShared(f *os.File) {
	if f == nil {
		return
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_SH)
}

func unlockStore(f *os.File) {
	if f == nil {
		return
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	_ = f.Close()
}
