//go:build unix

package config

import (
	"os"
	"path/filepath"
	"syscall"
)

// LockCredentials takes an exclusive flock beside the credentials file so
// writers (API-key saves, session saves, OAuth renewals) cannot interleave
// read-modify-write cycles. Retry EINTR: Go's async preemption (SIGURG) and
// other signals otherwise abort a blocking flock, and the caller must not
// proceed unlocked — a Claude refresh token is single-use.
func LockCredentials(path string) (func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	fd := int(f.Fd())
	for {
		err := syscall.Flock(fd, syscall.LOCK_EX)
		if err == nil {
			break
		}
		if err == syscall.EINTR {
			continue
		}
		_ = f.Close()
		return nil, err
	}
	return func() {
		_ = syscall.Flock(fd, syscall.LOCK_UN)
		_ = f.Close()
	}, nil
}
