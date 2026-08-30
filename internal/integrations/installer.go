package integrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InstallResult describes the outcome of an Install or Upgrade operation.
type InstallResult struct {
	ID           ID
	Action       string // "installed", "upgraded", "already_current", "uninstalled"
	TemplateFile string
	ConfigFile   string
	PreviousVer  string
	InstalledVer string
}

// Install renders the integration template, writes it to disk, and patches
// the target tool's config file to register the hook/plugin.
func Install(def Definition, dirs Dirs) (InstallResult, error) {
	targetFile := def.TargetFileFunc(dirs)
	configFile := def.ConfigFileFunc(dirs)

	// writesArtifact defaults to true; only integrations that register the
	// agentusage binary directly (no hook script on this platform) opt out.
	writesArtifact := def.WritesArtifact == nil || def.WritesArtifact(dirs)

	// Determine previous version (if any) for the result action.
	previousVer := ""
	if writesArtifact {
		if data, err := os.ReadFile(targetFile); err == nil {
			previousVer = parseIntegrationVersion(data)
		}
	}

	// Create parent directories.
	if writesArtifact {
		if err := os.MkdirAll(filepath.Dir(targetFile), 0o755); err != nil {
			return InstallResult{}, fmt.Errorf("integrations: create target dir: %w", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(configFile), 0o755); err != nil {
		return InstallResult{}, fmt.Errorf("integrations: create config dir: %w", err)
	}

	// Render template with version and binary placeholders.
	content := strings.ReplaceAll(def.Template, "__AGENTUSAGE_INTEGRATION_VERSION__", IntegrationVersion)
	content = strings.ReplaceAll(content, "__AGENTUSAGE_BIN_DEFAULT__", def.EscapeBin(dirs.AgentusageBin))

	// Backup existing files before overwriting.
	if writesArtifact {
		if err := backupIfExists(targetFile); err != nil {
			return InstallResult{}, fmt.Errorf("integrations: backup target: %w", err)
		}
	}
	if err := backupIfExists(configFile); err != nil {
		return InstallResult{}, fmt.Errorf("integrations: backup config: %w", err)
	}

	// Write rendered template.
	if writesArtifact {
		if err := os.WriteFile(targetFile, []byte(content), def.TemplateFileMode); err != nil {
			return InstallResult{}, fmt.Errorf("integrations: write template: %w", err)
		}
	}

	// Read config, patch it, write it back.
	configData, err := os.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return InstallResult{}, fmt.Errorf("integrations: read config: %w", err)
	}
	patched, err := def.ConfigPatcher(configData, targetFile, true)
	if err != nil {
		return InstallResult{}, fmt.Errorf("integrations: patch config: %w", err)
	}
	if err := os.WriteFile(configFile, patched, 0o600); err != nil {
		return InstallResult{}, fmt.Errorf("integrations: write config: %w", err)
	}

	syncContainerBoxes(def, dirs, targetFile, true)

	action := "installed"
	if previousVer != "" {
		action = "upgraded"
	}

	return InstallResult{
		ID:           def.ID,
		Action:       action,
		TemplateFile: targetFile,
		ConfigFile:   configFile,
		PreviousVer:  previousVer,
		InstalledVer: IntegrationVersion,
	}, nil
}

// Uninstall removes the integration's template file and patches the target
// tool's config file to unregister the hook/plugin.
func Uninstall(def Definition, dirs Dirs) error {
	targetFile := def.TargetFileFunc(dirs)
	configFile := def.ConfigFileFunc(dirs)

	// Patch config to remove hook/plugin entries.
	configData, err := os.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("integrations: read config for uninstall: %w", err)
	}
	if len(configData) > 0 {
		patched, err := def.ConfigPatcher(configData, targetFile, false)
		if err != nil {
			return fmt.Errorf("integrations: unpatch config: %w", err)
		}
		if err := os.WriteFile(configFile, patched, 0o600); err != nil {
			return fmt.Errorf("integrations: write config: %w", err)
		}
	}

	syncContainerBoxes(def, dirs, targetFile, false)

	// Remove the template file (unless this integration writes no artifact on
	// this platform, e.g. it registers the agentusage binary directly).
	if def.WritesArtifact == nil || def.WritesArtifact(dirs) {
		if err := os.Remove(targetFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("integrations: remove template: %w", err)
		}
	}

	return nil
}

func syncContainerBoxes(def Definition, dirs Dirs, targetFile string, install bool) {
	if def.ID == AntigravityID {
		containersDir := filepath.Join(dirs.Home, ".agy-containers")
		entries, err := os.ReadDir(containersDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				boxDir := filepath.Join(containersDir, entry.Name(), ".gemini", "antigravity-cli")
				boxConfigFile := filepath.Join(boxDir, "settings.json")
				configData, _ := os.ReadFile(boxConfigFile)
				if !install && len(configData) == 0 {
					continue
				}
				_ = os.MkdirAll(boxDir, 0o755)
				if patched, err := def.ConfigPatcher(configData, targetFile, install); err == nil {
					_ = os.WriteFile(boxConfigFile, patched, 0o600)
				}
			}
		}
	}
	if def.ID == CursorID {
		for _, cDirName := range []string{".agent-containers", ".cursor-containers"} {
			containersDir := filepath.Join(dirs.Home, cDirName)
			entries, err := os.ReadDir(containersDir)
			if err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						continue
					}
					boxDir := filepath.Join(containersDir, entry.Name(), ".cursor")
					boxConfigFile := filepath.Join(boxDir, "cli-config.json")
					configData, _ := os.ReadFile(boxConfigFile)
					if !install && len(configData) == 0 {
						continue
					}
					_ = os.MkdirAll(boxDir, 0o755)
					if patched, err := def.ConfigPatcher(configData, targetFile, install); err == nil {
						_ = os.WriteFile(boxConfigFile, patched, 0o600)
					}
				}
			}
		}
	}
}

// Upgrade re-installs the integration, always reporting the action as "upgraded".
func Upgrade(def Definition, dirs Dirs) (InstallResult, error) {
	result, err := Install(def, dirs)
	if err != nil {
		return result, err
	}
	result.Action = "upgraded"
	return result, nil
}
