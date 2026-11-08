//go:build windows

package claude_code

import "github.com/nurulislamz/agentusage/internal/config"

func lockAgentOAuthFile(path string) (func(), error) {
	return config.LockCredentials(path)
}
