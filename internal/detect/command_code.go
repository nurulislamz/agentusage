package detect

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nurulislamz/agentusage/internal/core"
)

func detectCommandCode(result *Result) {
	home := homeDir()
	cmdcConfigDir := ""
	if home != "" {
		cDir := filepath.Join(home, ".commandcode")
		if info, err := os.Stat(cDir); err == nil && info.IsDir() {
			cmdcConfigDir = cDir
		}
	}

	binPath := findCommandCodeBinary()
	if binPath == "" && cmdcConfigDir == "" {
		return
	}

	result.Tools = append(result.Tools, DetectedTool{
		Name:       "Command Code CLI",
		BinaryPath: binPath,
		ConfigDir:  cmdcConfigDir,
		Type:       "cli",
	})

	apiKey := findCommandCodeKey(home)

	acct := core.AccountConfig{
		ID:        "command_code",
		Provider:  "command_code",
		Auth:      "api_key",
		APIKeyEnv: "COMMAND_CODE_API_KEY",
		Token:     apiKey,
	}
	if apiKey != "" {
		acct.SetHint("credential_source", "command_code_detected")
	}
	for _, existing := range result.Accounts {
		if existing.ID == acct.ID {
			return
		}
	}
	result.Accounts = append(result.Accounts, acct)
}

func findCommandCodeBinary() string {
	for _, name := range []string{"command-code", "cmdc", "commandcode"} {
		if binPath, err := exec.LookPath(name); err == nil && binPath != "" {
			return binPath
		}
	}
	return ""
}

func findCommandCodeKey(home string) string {
	if envKey := os.Getenv("COMMAND_CODE_API_KEY"); strings.TrimSpace(envKey) != "" {
		return strings.TrimSpace(envKey)
	}
	if home == "" {
		return ""
	}
	if key := readCommandCodeAuthJSON(filepath.Join(home, ".commandcode", "auth.json")); key != "" {
		return key
	}
	return readCommandCodeSecretsEnv(filepath.Join(home, ".secrets", "commandcode.env"))
}

func readCommandCodeAuthJSON(path string) string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var payload map[string]any
	if json.Unmarshal(raw, &payload) != nil {
		return ""
	}
	if k, ok := payload["apiKey"].(string); ok && strings.TrimSpace(k) != "" {
		return strings.TrimSpace(k)
	}
	if k, ok := payload["key"].(string); ok && strings.TrimSpace(k) != "" {
		return strings.TrimSpace(k)
	}
	return ""
}

func readCommandCodeSecretsEnv(path string) string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "COMMAND_CODE_API_KEY=") {
			k := strings.TrimPrefix(line, "COMMAND_CODE_API_KEY=")
			k = strings.Trim(k, `"' `)
			if k != "" {
				return k
			}
		}
	}
	return ""
}
