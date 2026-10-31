//go:build windows

package claude_code

func lockAgentOAuthFile(path string) func() {
	return func() {}
}
