package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// LegacyStatuslineOutcome represents the action taken on a statusline file.
type LegacyStatuslineOutcome string

const (
	OutcomeCleaned       LegacyStatuslineOutcome = "cleaned"
	OutcomeSkippedCustom LegacyStatuslineOutcome = "skipped_custom"
	OutcomeNoStatusline  LegacyStatuslineOutcome = "no_statusline"
	OutcomeNotFound      LegacyStatuslineOutcome = "not_found"
	OutcomeError         LegacyStatuslineOutcome = "error"
)

// LegacyStatuslineResult records the migration outcome for a single config file.
type LegacyStatuslineResult struct {
	Path    string                  `json:"path"`
	Kind    string                  `json:"kind"` // "antigravity" or "cursor"
	Outcome LegacyStatuslineOutcome `json:"outcome"`
	Backup  string                  `json:"backup,omitempty"`
	Err     error                   `json:"-"`
}

// CleanLegacyStatusLines finds and removes agentUsage/openusage-owned Antigravity
// and Cursor status-line configurations across host and container box profiles.
// It creates a .bak backup before mutating any file, preserves custom commands
// and unrelated JSON, and leaves malformed files untouched with a per-file error.
// Running CleanLegacyStatusLines repeatedly is a no-op.
func CleanLegacyStatusLines(dirs Dirs) []LegacyStatuslineResult {
	var results []LegacyStatuslineResult

	// 1. Antigravity host config
	antigravityHost := findAntigravityConfigFile(dirs)
	results = append(results, CleanLegacyStatuslineFile(antigravityHost, "antigravity"))

	// 2. Antigravity box container profiles
	agyContainersDir := filepath.Join(dirs.Home, ".agy-containers")
	if entries, err := os.ReadDir(agyContainersDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			boxFile := filepath.Join(agyContainersDir, entry.Name(), ".gemini", "antigravity-cli", "settings.json")
			results = append(results, CleanLegacyStatuslineFile(boxFile, "antigravity"))
		}
	}

	// 3. Cursor host config
	cursorHost := findCursorConfigFile(dirs)
	results = append(results, CleanLegacyStatuslineFile(cursorHost, "cursor"))

	// 4. Cursor box container profiles (.agent-containers and .cursor-containers)
	for _, cDirName := range []string{".agent-containers", ".cursor-containers"} {
		cDir := filepath.Join(dirs.Home, cDirName)
		if entries, err := os.ReadDir(cDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				boxFile := filepath.Join(cDir, entry.Name(), ".cursor", "cli-config.json")
				results = append(results, CleanLegacyStatuslineFile(boxFile, "cursor"))
			}
		}
	}

	return results
}

func findAntigravityConfigFile(dirs Dirs) string {
	if f := strings.TrimSpace(os.Getenv("ANTIGRAVITY_SETTINGS_FILE")); f != "" {
		return f
	}
	configDir := strings.TrimSpace(os.Getenv("ANTIGRAVITY_CONFIG_DIR"))
	if configDir == "" {
		configDir = filepath.Join(dirs.Home, ".gemini", "antigravity-cli")
	}
	return filepath.Join(configDir, "settings.json")
}

func findCursorConfigFile(dirs Dirs) string {
	if f := strings.TrimSpace(os.Getenv("CURSOR_SETTINGS_FILE")); f != "" {
		return f
	}
	configDir := strings.TrimSpace(os.Getenv("CURSOR_CONFIG_DIR"))
	if configDir == "" {
		configDir = filepath.Join(dirs.Home, ".cursor")
	}
	return filepath.Join(configDir, "cli-config.json")
}

// CleanLegacyStatuslineFile inspects path for an owned statusline command of the given kind.
// If owned, it creates path.bak, deletes statusLine, and writes back the updated JSON.
// If custom, it preserves it (OutcomeSkippedCustom).
// If malformed, it leaves the file untouched and returns an error (OutcomeError).
// If missing, returns OutcomeNotFound.
// If no statusLine present, returns OutcomeNoStatusline.
func CleanLegacyStatuslineFile(path string, kind string) LegacyStatuslineResult {
	res := LegacyStatuslineResult{
		Path: path,
		Kind: kind,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			res.Outcome = OutcomeNotFound
			return res
		}
		res.Outcome = OutcomeError
		res.Err = err
		return res
	}

	if len(bytes.TrimSpace(data)) == 0 {
		res.Outcome = OutcomeNoStatusline
		return res
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		res.Outcome = OutcomeError
		res.Err = fmt.Errorf("malformed JSON in %s: %w", path, err)
		return res
	}

	statusLineRaw, exists := root["statusLine"]
	if !exists {
		res.Outcome = OutcomeNoStatusline
		return res
	}

	statusLine, ok := statusLineRaw.(map[string]any)
	if !ok {
		// Non-object statusLine entry; preserve custom configuration.
		res.Outcome = OutcomeSkippedCustom
		return res
	}

	cmdStr, _ := statusLine["command"].(string)
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" || !IsOwnedStatuslineCommand(cmdStr, kind) {
		res.Outcome = OutcomeSkippedCustom
		return res
	}

	// Owned command: save backup before mutation.
	backupPath := path + ".bak"
	if err := os.WriteFile(backupPath, data, 0o600); err != nil {
		res.Outcome = OutcomeError
		res.Err = fmt.Errorf("backup %s: %w", path, err)
		return res
	}
	res.Backup = backupPath

	// Delete statusLine from root.
	delete(root, "statusLine")

	encoded, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		res.Outcome = OutcomeError
		res.Err = fmt.Errorf("serialize %s: %w", path, err)
		return res
	}
	encoded = append(encoded, '\n')

	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		res.Outcome = OutcomeError
		res.Err = fmt.Errorf("write %s: %w", path, err)
		return res
	}

	res.Outcome = OutcomeCleaned
	return res
}

// IsOwnedStatuslineCommand reports whether command is an agentUsage or openusage
// owned statusline command for the given kind ("antigravity", "cursor", or "" for any).
func IsOwnedStatuslineCommand(command string, kind string) bool {
	tokens := parseCommandTokens(command)
	if len(tokens) < 3 {
		return false
	}

	// Check binary (token 0): base name must be agentusage or openusage
	exe := tokens[0]
	base := strings.ToLower(filepath.Base(exe))
	base = strings.TrimSuffix(base, ".exe")
	if base != "agentusage" && base != "openusage" {
		return false
	}

	// Check subcommand (token 1): must be antigravity or cursor
	subcmd := strings.ToLower(tokens[1])
	if kind != "" && subcmd != strings.ToLower(kind) {
		return false
	}
	if subcmd != "antigravity" && subcmd != "cursor" {
		return false
	}

	// Check action (token 2): must be statusline
	action := strings.ToLower(tokens[2])
	return action == "statusline"
}

func parseCommandTokens(cmd string) []string {
	var tokens []string
	var current strings.Builder
	inQuote := rune(0)
	for _, r := range cmd {
		switch {
		case inQuote != 0:
			if r == inQuote {
				inQuote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '"' || r == '\'':
			inQuote = r
		case unicode.IsSpace(r):
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}
