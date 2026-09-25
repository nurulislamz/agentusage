package boxes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type CodexBoxStatus string

const (
	CodexStatusInitialized   CodexBoxStatus = "Initialized"
	CodexStatusConfigured    CodexBoxStatus = "Configured"
	CodexStatusAuthenticated CodexBoxStatus = "Authenticated"
	CodexStatusReady         CodexBoxStatus = "Ready"
)

type CodexBox struct {
	Name         string         `json:"name"`
	AccountID    string         `json:"account_id"`
	Path         string         `json:"path"`
	Status       CodexBoxStatus `json:"status"`
	LastModified time.Time      `json:"last_modified,omitempty"`
}

func DefaultCodexContainersDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".codex-containers")
}

func resolveCodexContainersDir(baseDirs []string) string {
	if len(baseDirs) > 0 && strings.TrimSpace(baseDirs[0]) != "" {
		return baseDirs[0]
	}
	return DefaultCodexContainersDir()
}

func CreateCodexBox(ctx context.Context, name string, baseDirs ...string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("box name cannot be empty")
	}
	if strings.HasPrefix(name, "-") || strings.ContainsAny(name, " /\\:") {
		return "", fmt.Errorf("invalid box name: %q", name)
	}

	root := resolveCodexContainersDir(baseDirs)
	if root == "" {
		return "", fmt.Errorf("cannot determine containers directory")
	}

	profileDir := filepath.Join(root, name)
	if _, err := os.Stat(profileDir); err == nil {
		return profileDir, fmt.Errorf("box %q already exists", name)
	}

	codexDir := filepath.Join(profileDir, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		return "", fmt.Errorf("create profile directory: %w", err)
	}

	// Sync host skills, rules, AGENTS.md if present
	home, _ := os.UserHomeDir()
	if home != "" {
		for _, folder := range []string{"skills", "rules"} {
			src := filepath.Join(home, ".codex", folder)
			if info, err := os.Stat(src); err == nil && info.IsDir() {
				dst := filepath.Join(codexDir, folder)
				_ = os.MkdirAll(dst, 0o755)
				_ = copyDirContents(src, dst)
			}
		}
		agentsMD := filepath.Join(home, ".codex", "AGENTS.md")
		if info, err := os.Stat(agentsMD); err == nil && !info.IsDir() {
			if content, err := os.ReadFile(agentsMD); err == nil {
				_ = os.WriteFile(filepath.Join(codexDir, "AGENTS.md"), content, 0o644)
			}
		}
	}

	return profileDir, nil
}

func DeleteCodexBox(ctx context.Context, name string, baseDirs ...string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("box name cannot be empty")
	}
	root := resolveCodexContainersDir(baseDirs)
	if root == "" {
		return fmt.Errorf("cannot determine containers directory")
	}
	profileDir := filepath.Join(root, name)
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		return fmt.Errorf("box %q does not exist", name)
	}
	return os.RemoveAll(profileDir)
}

func ListCodexBoxes(ctx context.Context, baseDirs ...string) ([]CodexBox, error) {
	root := resolveCodexContainersDir(baseDirs)
	if root == "" {
		return nil, fmt.Errorf("cannot determine containers directory")
	}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return []CodexBox{}, nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read containers directory: %w", err)
	}

	var boxList []CodexBox
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		profileDir := filepath.Join(root, name)
		codexDir := filepath.Join(profileDir, ".codex")
		if !dirExists(codexDir) {
			codexDir = profileDir
		}

		status := CodexStatusReady
		authFile := filepath.Join(codexDir, "auth.json")
		configFile := filepath.Join(codexDir, "config.toml")
		if fileExists(authFile) {
			status = CodexStatusAuthenticated
		} else if fileExists(configFile) {
			status = CodexStatusConfigured
		} else if dirExists(codexDir) {
			status = CodexStatusInitialized
		}

		var lastMod time.Time
		if fi, err := os.Stat(profileDir); err == nil {
			lastMod = fi.ModTime()
		}

		boxList = append(boxList, CodexBox{
			Name:         name,
			AccountID:    fmt.Sprintf("codex-%s", name),
			Path:         profileDir,
			Status:       status,
			LastModified: lastMod,
		})
	}

	sort.Slice(boxList, func(i, j int) bool {
		return boxList[i].Name < boxList[j].Name
	})

	return boxList, nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
