//go:build unix

package claude_code

import "github.com/nurulislamz/agentusage/internal/config"

// lockAgentOAuthFile serializes Claude sign-in renewal against every other
// credentials.json writer (API keys, sessions, logout, a second process).
// The lock is the same file as config.LockCredentials; fail-closed.
func lockAgentOAuthFile(path string) (func(), error) {
	return config.LockCredentials(path)
}
