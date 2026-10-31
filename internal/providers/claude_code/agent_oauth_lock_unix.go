//go:build unix

package claude_code

import (
	"os"
	"syscall"
)

// lockAgentOAuthFile takes an exclusive flock beside the credentials file so
// only one process renews the sign-in at a time. Best-effort: if the lock
// can't be taken, the caller proceeds unlocked.
func lockAgentOAuthFile(path string) func() {
	f, err := os.OpenFile(path+".claude-oauth.lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return func() {}
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return func() {}
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}
}
