//go:build windows

package integrations

import (
	"path/filepath"

	"github.com/nurulislamz/agentusage/internal/config"
)

// platformHooksDir places agentUsage's hook scripts beside settings.json under
// %APPDATA%\agentusage\hooks on Windows. The configRoot argument (the XDG-style
// base used for third-party tool dirs) is intentionally ignored here.
func platformHooksDir(_ string) string {
	return filepath.Join(config.ConfigDir(), "hooks")
}
