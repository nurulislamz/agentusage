package detect

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nurulislamz/agentusage/internal/core"
)

func detectCursor(result *Result) {
	bin := findBinary("cursor")

	home := homeDir()
	if home == "" {
		return
	}

	// 1. Auto-detect all active cursor-box container profiles in ~/.agent-containers and ~/.cursor-containers
	hasBoxes := false
	for _, containersDirName := range []string{".agent-containers", ".cursor-containers"} {
		containersDir := filepath.Join(home, containersDirName)
		if !dirExists(containersDir) {
			continue
		}
		entries, err := os.ReadDir(containersDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			boxName := entry.Name()
			boxConfigDir := filepath.Join(containersDir, boxName, ".cursor")

			acct := core.AccountConfig{
				ID:           fmt.Sprintf("cursor-%s", boxName),
				Provider:     "cursor",
				Auth:         "local",
				Binary:       bin,
				RuntimeHints: make(map[string]string),
			}
			acct.SetHint("config_dir", boxConfigDir)
			acct.SetHint("auth_file", filepath.Join(containersDir, boxName, ".config", "cursor", "auth.json"))
			addAccount(result, acct)
			hasBoxes = true
		}
	}

	// 2. Default single-profile Cursor config dir
	configDir := strings.TrimSpace(os.Getenv("CURSOR_CONFIG_DIR"))
	if configDir == "" {
		configDir = filepath.Join(home, ".cursor")
	}

	if !dirExists(configDir) && bin == "" {
		return
	}

	if bin != "" {
		log.Printf("[detect] Found Cursor at %s", bin)
		result.Tools = append(result.Tools, DetectedTool{
			Name:       "Cursor CLI",
			BinaryPath: bin,
			ConfigDir:  configDir,
			Type:       "cli",
		})
	}

	if !hasBoxes {
		acct := core.AccountConfig{
			ID:           "cursor",
			Provider:     "cursor",
			Auth:         "local",
			Binary:       bin,
			RuntimeHints: make(map[string]string),
		}
		acct.SetHint("config_dir", configDir)
		acct.SetHint("auth_file", filepath.Join(home, ".config", "cursor", "auth.json"))
		addAccount(result, acct)
	}
}
