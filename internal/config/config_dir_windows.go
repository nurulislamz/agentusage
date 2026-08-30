//go:build windows

package config

import (
	"os"
	"path/filepath"
)

// osConfigDir returns the agentUsage config directory on Windows: %APPDATA%\agentusage.
func osConfigDir() string {
	return filepath.Join(os.Getenv("APPDATA"), "agentusage")
}
