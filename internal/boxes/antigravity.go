package boxes

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

type AntigravityBoxStatus string

const (
	StatusInitialized   AntigravityBoxStatus = "Initialized"
	StatusAuthenticated AntigravityBoxStatus = "Authenticated"
	StatusReady         AntigravityBoxStatus = "Ready"
)

type AntigravityBox struct {
	Name         string               `json:"name"`
	AccountID    string               `json:"account_id"`
	Path         string               `json:"path"`
	Status       AntigravityBoxStatus `json:"status"`
	TokenFile    string               `json:"token_file,omitempty"`
	TokenStatus  string               `json:"token_status,omitempty"`
	LastModified time.Time            `json:"last_modified,omitempty"`
}

var oauthURLPattern = regexp.MustCompile(`https?://[^\s\"\'\<\>]+`)

type BoxRunner func(ctx context.Context, box string, args ...string) (io.ReadCloser, func(), error)

type LoginOptions struct {
	BrowserOpener func(url string) error
	OnAuthURL     func(url string)
	OnTokenSaved  func()
	PollInterval  time.Duration
	Timeout       time.Duration
	Runner        BoxRunner
	BaseDir       string
}

func DefaultContainersDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".agy-containers")
}

func resolveContainersDir(baseDirs []string) string {
	if len(baseDirs) > 0 && strings.TrimSpace(baseDirs[0]) != "" {
		return baseDirs[0]
	}
	return DefaultContainersDir()
}

func CreateBox(ctx context.Context, name string, baseDirs ...string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("box name cannot be empty")
	}
	if strings.HasPrefix(name, "-") || strings.ContainsAny(name, " /\\:") {
		return "", fmt.Errorf("invalid box name: %q", name)
	}

	root := resolveContainersDir(baseDirs)
	if root == "" {
		return "", fmt.Errorf("cannot determine containers directory")
	}

	profileDir := filepath.Join(root, name)
	if _, err := os.Stat(profileDir); err == nil {
		return profileDir, fmt.Errorf("box %q already exists", name)
	}

	cliDir := filepath.Join(profileDir, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(cliDir, 0o755); err != nil {
		return "", fmt.Errorf("create profile directory: %w", err)
	}

	// Seed settings.json with OpenUsage statusLine hook
	settingsPath := filepath.Join(cliDir, "settings.json")
	settingsPayload := map[string]interface{}{
		"statusLine": map[string]string{
			"command": "openusage antigravity statusline",
		},
	}
	data, _ := json.MarshalIndent(settingsPayload, "", "  ")
	_ = os.WriteFile(settingsPath, append(data, '\n'), 0o644)

	// Sync host workflows, skills, rules if present
	home, _ := os.UserHomeDir()
	if home != "" {
		for _, folder := range []string{"workflows", "global_workflows", "skills", "agents", "rules"} {
			src := filepath.Join(home, ".gemini", "config", folder)
			if info, err := os.Stat(src); err == nil && info.IsDir() {
				dst := filepath.Join(profileDir, ".gemini", "config", folder)
				_ = os.MkdirAll(dst, 0o755)
				_ = copyDirContents(src, dst)
			}
		}
	}

	return profileDir, nil
}

func DeleteBox(ctx context.Context, name string, baseDirs ...string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("box name cannot be empty")
	}
	root := resolveContainersDir(baseDirs)
	if root == "" {
		return fmt.Errorf("cannot determine containers directory")
	}
	profileDir := filepath.Join(root, name)
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		return fmt.Errorf("box %q does not exist", name)
	}
	return os.RemoveAll(profileDir)
}

func ListBoxes(ctx context.Context, baseDirs ...string) ([]AntigravityBox, error) {
	root := resolveContainersDir(baseDirs)
	if root == "" {
		return nil, fmt.Errorf("cannot determine containers directory")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []AntigravityBox{}, nil
		}
		return nil, fmt.Errorf("read containers directory: %w", err)
	}

	var boxList []AntigravityBox
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		box := InspectBox(root, name)
		boxList = append(boxList, box)
	}

	sort.Slice(boxList, func(i, j int) bool {
		return strings.ToLower(boxList[i].Name) < strings.ToLower(boxList[j].Name)
	})

	return boxList, nil
}

func InspectBox(root, name string) AntigravityBox {
	profileDir := filepath.Join(root, name)
	info, err := os.Stat(profileDir)
	var modTime time.Time
	if err == nil {
		modTime = info.ModTime()
	}

	box := AntigravityBox{
		Name:         name,
		AccountID:    "antigravity-" + name,
		Path:         profileDir,
		Status:       StatusInitialized,
		LastModified: modTime,
	}

	tokenPath := filepath.Join(profileDir, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	legacyOauth := filepath.Join(profileDir, ".gemini", "oauth_creds.json")
	legacyGemini := filepath.Join(profileDir, ".gemini", "gemini-credentials.json")

	if fileExists(tokenPath) {
		box.TokenFile = tokenPath
		box.Status = StatusAuthenticated
		if isValidTokenFile(tokenPath) {
			box.Status = StatusReady
			box.TokenStatus = "ready"
		} else {
			box.TokenStatus = "token present"
		}
	} else if fileExists(legacyOauth) {
		box.TokenFile = legacyOauth
		box.Status = StatusAuthenticated
		box.TokenStatus = "legacy oauth"
	} else if fileExists(legacyGemini) {
		box.TokenFile = legacyGemini
		box.Status = StatusAuthenticated
		box.TokenStatus = "gemini creds"
	}

	return box
}

func isValidTokenFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return false
	}
	var payload struct {
		Token struct {
			AccessToken string `json:"access_token"`
		} `json:"token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return false
	}
	return strings.TrimSpace(payload.Token.AccessToken) != "" || strings.TrimSpace(payload.AccessToken) != ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func copyDirContents(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		s := filepath.Join(src, entry.Name())
		d := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			_ = os.MkdirAll(d, 0o755)
			_ = copyDirContents(s, d)
		} else {
			data, err := os.ReadFile(s)
			if err == nil {
				_ = os.WriteFile(d, data, 0o644)
			}
		}
	}
	return nil
}

func DefaultBoxRunner(ctx context.Context, box string, args ...string) (io.ReadCloser, func(), error) {
	bin := "agy-box"
	if path, err := exec.LookPath("agy-box"); err == nil && path != "" {
		bin = path
	} else if dir := core.UserLocalBinDir(); dir != "" {
		cand := filepath.Join(dir, "agy-box")
		if info, err := os.Stat(cand); err == nil && !info.IsDir() {
			bin = cand
		}
	}

	cmdArgs := append([]string{box}, args...)
	cmd := exec.CommandContext(ctx, bin, cmdArgs...)
	cmd.Env = core.EnvironWithUserLocalBin(os.Environ())

	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		_ = pr.Close()
		_ = pw.Close()
		return nil, nil, fmt.Errorf("start agy-box %s: %w", box, err)
	}

	cancel := func() {
		_ = cmd.Process.Kill()
		_ = pw.Close()
		_ = pr.Close()
	}

	go func() {
		_ = cmd.Wait()
		_ = pw.Close()
	}()

	return pr, cancel, nil
}

func LoginBoxSession(ctx context.Context, name string, opts LoginOptions) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("box name cannot be empty")
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	pollInterval := opts.PollInterval
	if pollInterval <= 0 {
		pollInterval = 500 * time.Millisecond
	}

	root := opts.BaseDir
	if root == "" {
		root = DefaultContainersDir()
	}
	tokenPath := filepath.Join(root, name, ".gemini", "antigravity-cli", "antigravity-oauth-token")

	runner := opts.Runner
	if runner == nil {
		runner = DefaultBoxRunner
	}

	reader, runnerCancel, err := runner(ctx, name)
	if err != nil {
		return err
	}
	defer runnerCancel()

	urlDetected := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			matches := oauthURLPattern.FindAllString(line, -1)
			for _, m := range matches {
				if isLikelyAuthURL(m) {
					select {
					case urlDetected <- m:
					default:
					}
				}
			}
		}
	}()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	openedBrowser := false

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("login timeout: oauth token not saved within %v", timeout)

		case authURL := <-urlDetected:
			if !openedBrowser {
				openedBrowser = true
				if opts.OnAuthURL != nil {
					opts.OnAuthURL(authURL)
				}
				if opts.BrowserOpener != nil {
					_ = opts.BrowserOpener(authURL)
				}
			}

		case <-ticker.C:
			if fileExists(tokenPath) && isValidTokenFile(tokenPath) {
				if opts.OnTokenSaved != nil {
					opts.OnTokenSaved()
				}
				return nil
			}
		}
	}
}

func isLikelyAuthURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Host)
	path := strings.ToLower(u.Path)
	if strings.Contains(host, "google.com") && (strings.Contains(path, "oauth") || strings.Contains(path, "auth")) {
		return true
	}
	return strings.Contains(raw, "oauth") || strings.Contains(raw, "client_id") || strings.Contains(raw, "code=")
}
