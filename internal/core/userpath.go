package core

import (
	"os"
	"path/filepath"
	"strings"
)

// UserLocalBinDir is $HOME/.local/bin, where `make install` and `make box`
// place agentusage and the box CLIs (agy-box, agent-box, opencode-box).
func UserLocalBinDir() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ""
	}
	return filepath.Join(home, ".local", "bin")
}

// EnsureUserLocalBinOnPATH prepends $HOME/.local/bin to PATH when it is
// missing. Systemd user units (and similar service managers) typically omit
// that directory, so box CLIs are invisible to agentusage and to children
// such as `agy-box … -p ping` (which then execs `agy` from PATH).
func EnsureUserLocalBinOnPATH() {
	if path, ok := prependUserLocalBin(os.Getenv("PATH")); ok {
		_ = os.Setenv("PATH", path)
	}
}

// EnvironWithUserLocalBin returns env with $HOME/.local/bin prepended to PATH.
// If env is nil, it copies os.Environ().
func EnvironWithUserLocalBin(env []string) []string {
	if env == nil {
		env = os.Environ()
	}
	path := envValue(env, "PATH")
	next, ok := prependUserLocalBin(path)
	if !ok {
		return env
	}
	return setEnvValue(env, "PATH", next)
}

func prependUserLocalBin(path string) (string, bool) {
	dir := UserLocalBinDir()
	if dir == "" {
		return path, false
	}
	sep := string(os.PathListSeparator)
	for _, part := range strings.Split(path, sep) {
		if part == dir {
			return path, false
		}
	}
	if strings.TrimSpace(path) == "" {
		return dir, true
	}
	return dir + sep + path, true
}

func envValue(env []string, key string) string {
	prefix := key + "="
	for i := len(env) - 1; i >= 0; i-- {
		if strings.HasPrefix(env[i], prefix) {
			return env[i][len(prefix):]
		}
	}
	return ""
}

func setEnvValue(env []string, key, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	replaced := false
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			if !replaced {
				out = append(out, prefix+value)
				replaced = true
			}
			continue
		}
		out = append(out, entry)
	}
	if !replaced {
		out = append(out, prefix+value)
	}
	return out
}
