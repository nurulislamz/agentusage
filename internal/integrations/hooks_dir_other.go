//go:build !windows

package integrations

import "path/filepath"

// platformHooksDir places agentUsage's hook scripts under <configRoot>/agentusage/hooks
// on Unix (configRoot is XDG_CONFIG_HOME or ~/.config).
func platformHooksDir(configRoot string) string {
	return filepath.Join(configRoot, "agentusage", "hooks")
}
