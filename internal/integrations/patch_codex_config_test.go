package integrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPatchCodexConfig_MultilineNotifyInstallUninstall(t *testing.T) {
	home := t.TempDir()
	codexDir := filepath.Join(home, ".codex")
	hooksDir := filepath.Join(home, ".config", "agentusage", "hooks")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}

	original := "model = \"o3\"\nnotify = [\n  \"python3\",\n  \"/home/user/bin/notify.py\",\n]\nmcp_servers.example.command = \"echo\"\n"
	configPath := filepath.Join(codexDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}

	def, ok := DefinitionByID(CodexID)
	if !ok {
		t.Fatal("missing codex def")
	}
	dirs := Dirs{Home: home, HooksDir: hooksDir, AgentusageBin: "/usr/bin/agentusage"}

	if _, err := Install(def, dirs); err != nil {
		t.Fatalf("install: %v", err)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if strings.Contains(got, "python3") || strings.Contains(got, "notify.py") {
		t.Fatalf("orphan notify array lines remain:\n%s", got)
	}
	if strings.Count(got, "[") != strings.Count(got, "]") {
		t.Fatalf("unbalanced brackets:\n%s", got)
	}
	if !strings.Contains(got, "model = \"o3\"") || !strings.Contains(got, "mcp_servers.example.command") {
		t.Fatalf("lost unrelated config:\n%s", got)
	}
	if !codexConfigured(got) {
		t.Fatalf("agentusage notify not registered:\n%s", got)
	}

	bakPath := configPath + ".bak"
	bak, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatalf("read bak: %v", err)
	}
	if string(bak) != original {
		t.Fatalf("first bak should be pristine original, got:\n%s", bak)
	}

	// Second install must keep the original backup and leave config valid.
	if _, err := Install(def, dirs); err != nil {
		t.Fatalf("second install: %v", err)
	}
	bak2, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(bak2) != original {
		t.Fatalf("second install overwrote pristine bak:\n%s", bak2)
	}

	if err := Uninstall(def, dirs); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	afterStr := string(after)
	if strings.Contains(afterStr, "notify") || strings.Contains(afterStr, "python3") || strings.Contains(afterStr, "notify.py") {
		t.Fatalf("uninstall left notify remnants:\n%s", afterStr)
	}
	if !strings.Contains(afterStr, "model = \"o3\"") || !strings.Contains(afterStr, "mcp_servers.example.command") {
		t.Fatalf("uninstall lost unrelated config:\n%s", afterStr)
	}
}

func TestFindNotifyAssignmentSpan(t *testing.T) {
	cases := []struct {
		name      string
		lines     []string
		wantStart int
		wantEnd   int
		wantFound bool
	}{
		{
			name:      "single line",
			lines:     []string{`model = "o3"`, `notify = ["hook.sh"]`, `x = 1`},
			wantStart: 1,
			wantEnd:   1,
			wantFound: true,
		},
		{
			name: "multi line array",
			lines: []string{
				`model = "o3"`,
				`notify = [`,
				`  "python3",`,
				`  "/home/user/bin/notify.py",`,
				`]`,
				`x = 1`,
			},
			wantStart: 1,
			wantEnd:   4,
			wantFound: true,
		},
		{
			name:      "absent",
			lines:     []string{`model = "o3"`},
			wantFound: false,
		},
		{
			name:      "notify_prefix_not_key",
			lines:     []string{`notify_extra = 1`, `model = "o3"`},
			wantFound: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			start, end, found := findNotifyAssignmentSpan(tc.lines)
			if found != tc.wantFound || start != tc.wantStart || end != tc.wantEnd {
				t.Fatalf("got start=%d end=%d found=%v, want start=%d end=%d found=%v",
					start, end, found, tc.wantStart, tc.wantEnd, tc.wantFound)
			}
		})
	}
}
