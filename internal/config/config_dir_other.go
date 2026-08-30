//go:build !windows

package config

import (
	"os"
	"path/filepath"
)

// osConfigDir returns the agentUsage config directory on Unix: ~/.config/agentusage.
// XDG_CONFIG_HOME is intentionally not honored (see docs/reference/paths.md).
func osConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "agentusage")
}
