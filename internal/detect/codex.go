package detect

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nurulislamz/agentusage/internal/core"
)

// codexOpenAIAccountID is the account ID we use when adopting an OPENAI_API_KEY
// stored in ~/.codex/auth.json. It matches the canonical id used by
// detectEnvKeys so addAccount() de-dupes consistently with the env-var path.
const codexOpenAIAccountID = "openai"

func detectCodex(result *Result) {
	home := homeDir()
	bin := findBinary("codex")

	// 1. Auto-detect all active codex-box container profiles in ~/.codex-containers
	containersDir := filepath.Join(home, ".codex-containers")
	hasBoxes := false
	if dirExists(containersDir) {
		entries, err := os.ReadDir(containersDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				boxName := entry.Name()
				boxProfileDir := filepath.Join(containersDir, boxName)
				boxConfigDir := filepath.Join(boxProfileDir, ".codex")
				if !dirExists(boxConfigDir) {
					if dirExists(filepath.Join(boxProfileDir, "sessions")) || fileExists(filepath.Join(boxProfileDir, "auth.json")) {
						boxConfigDir = boxProfileDir
					}
				}

				boxSessionsDir := filepath.Join(boxConfigDir, "sessions")
				boxAuthFile := filepath.Join(boxConfigDir, "auth.json")
				hasBoxSessions := dirExists(boxSessionsDir)
				hasBoxAuth := fileExists(boxAuthFile)

				acct := core.AccountConfig{
					ID:           fmt.Sprintf("codex-%s", boxName),
					Provider:     "codex",
					Auth:         "local",
					Binary:       bin,
					RuntimeHints: make(map[string]string),
				}
				acct.SetHint("config_dir", boxConfigDir)
				acct.SetHint("box_name", boxName)
				if hasBoxSessions {
					acct.SetHint("sessions_dir", boxSessionsDir)
				}
				if hasBoxAuth {
					acct.SetHint("auth_file", boxAuthFile)
					email, accountID, planType, _ := extractCodexAuth(boxAuthFile)
					if email != "" {
						acct.RuntimeHints["email"] = email
					}
					if accountID != "" {
						acct.RuntimeHints["account_id"] = accountID
					}
					if planType != "" {
						acct.RuntimeHints["plan_type"] = planType
					}
				}
				addAccount(result, acct)
				hasBoxes = true
			}
		}
	}

	// 2. Default single-profile Codex config dir
	configDir := strings.TrimSpace(os.Getenv("CODEX_CONFIG_DIR"))
	if configDir == "" {
		configDir = filepath.Join(home, ".codex")
	}

	if !dirExists(configDir) && bin == "" && !hasBoxes {
		return
	}

	if bin != "" {
		tool := DetectedTool{
			Name:       "OpenAI Codex CLI",
			BinaryPath: bin,
			ConfigDir:  configDir,
			Type:       "cli",
		}
		result.Tools = append(result.Tools, tool)
		log.Printf("[detect] Found Codex CLI at %s", bin)
	}

	sessionsDir := filepath.Join(configDir, "sessions")
	authFile := filepath.Join(configDir, "auth.json")

	hasSessions := dirExists(sessionsDir)
	hasAuth := fileExists(authFile)

	if !hasSessions && !hasAuth {
		return
	}

	log.Printf("[detect] Codex CLI data found (sessions=%v, auth=%v)", hasSessions, hasAuth)

	acct := core.AccountConfig{
		ID:           "codex-cli",
		Provider:     "codex",
		Auth:         "local",
		Binary:       bin,
		RuntimeHints: make(map[string]string),
	}

	acct.SetHint("config_dir", configDir)
	acct.RuntimeHints["config_dir"] = configDir

	if hasSessions {
		acct.SetHint("sessions_dir", sessionsDir)
		acct.RuntimeHints["sessions_dir"] = sessionsDir
	}

	if hasAuth {
		acct.SetHint("auth_file", authFile)
		acct.RuntimeHints["auth_file"] = authFile
		email, accountID, planType, openaiAPIKey := extractCodexAuth(authFile)
		if email != "" {
			acct.RuntimeHints["email"] = email
			log.Printf("[detect] Codex account: %s", email)
		}
		if accountID != "" {
			acct.RuntimeHints["account_id"] = accountID
		}
		if planType != "" {
			acct.RuntimeHints["plan_type"] = planType
			log.Printf("[detect] Codex plan: %s", planType)
		}
		// When the user logged in via API key, codex stores the raw
		// OPENAI_API_KEY at the top level of auth.json (Rust struct field
		// `#[serde(rename = "OPENAI_API_KEY")] api_key`). Adopt it as a
		// standard openai account so the openai provider can use it.
		// Skip if the env var is already set — env wins over file.
		if openaiAPIKey != "" && os.Getenv("OPENAI_API_KEY") == "" {
			openai := core.AccountConfig{
				ID:       codexOpenAIAccountID,
				Provider: "openai",
				Auth:     "api_key",
				Token:    openaiAPIKey,
			}
			openai.SetHint("credential_source", "codex_auth_json")
			before := len(result.Accounts)
			addAccount(result, openai)
			if len(result.Accounts) > before {
				log.Printf("[detect] Adopted OPENAI_API_KEY from %s (key=%s)",
					authFile, maskKey(openaiAPIKey))
			}
		}
	}

	addAccount(result, acct)
}

type codexAuthFile struct {
	Tokens       codexTokens `json:"tokens"`
	AccountID    string      `json:"account_id"`
	OpenAIAPIKey string      `json:"OPENAI_API_KEY"`
}

type codexTokens struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func extractCodexAuth(authFile string) (email, accountID, planType, openaiAPIKey string) {
	data, err := os.ReadFile(authFile)
	if err != nil {
		log.Printf("[detect] Cannot read Codex auth.json: %v", err)
		return "", "", "", ""
	}

	var auth codexAuthFile
	if err := json.Unmarshal(data, &auth); err != nil {
		log.Printf("[detect] Cannot parse Codex auth.json: %v", err)
		return "", "", "", ""
	}

	accountID = auth.AccountID
	openaiAPIKey = strings.TrimSpace(auth.OpenAIAPIKey)

	if auth.Tokens.IDToken != "" {
		claims := decodeJWTPayload(auth.Tokens.IDToken)
		if claims != nil {
			if e, ok := claims["email"].(string); ok {
				email = e
			}
			if authData, ok := claims["https://api.openai.com/auth"].(map[string]interface{}); ok {
				if pt, ok := authData["chatgpt_plan_type"].(string); ok {
					planType = pt
				}
			}
		}
	}

	return email, accountID, planType, openaiAPIKey
}

func decodeJWTPayload(token string) map[string]interface{} {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) < 2 {
		return nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil
	}
	return claims
}
