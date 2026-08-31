package tmux

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/export"
	"github.com/nurulislamz/agentusage/internal/report"
)

// errTestWriter is an io.Writer that always fails, used for testing negative write branches.
type errTestWriter struct{}

func (errTestWriter) Write(p []byte) (int, error) {
	return 0, errors.New("simulated write error")
}

// -----------------------------------------------------------------------------
// Axis 1: Happy Paths
// -----------------------------------------------------------------------------

func TestHappyPath_RenderAllModifiersAndWidgets(t *testing.T) {
	theme := ThemeColors{
		Base:   "#1E1E2E",
		Text:   "#CDD6F4",
		Accent: "#CBA6F7",
		Green:  "#A6E3A1",
		Yellow: "#F9E2AF",
		Red:    "#F38BA8",
	}
	used := func(v float64) *float64 { return &v }

	ctx := Context{
		Provider: "claude_code",
		Account:  "work",
		Snapshot: core.UsageSnapshot{
			ProviderID: "claude_code",
			AccountID:  "work",
			Metrics: map[string]core.Metric{
				"today_api_cost":  {Used: used(12.3456)},
				"5h_block_cost":   {Used: used(5.50)},
				"usage_five_hour": {Used: used(45.0)},
				"tokens_used":     {Used: used(1_500_000)},
			},
			Attributes: map[string]string{
				"model": "claude-3-7-sonnet",
			},
		},
		Synthetic: map[string]string{
			"_block_remaining":  "1h45m",
			"_block_burn_rate":  "$2.50",
			"_block_projection": "$8.00",
			"_context_tokens":   "15000",
			"_context_pct":      "30",
		},
		Theme:     theme,
		ThemeRefs: ThemeRefs(theme, ColorModeTruecolor),
		Variables: map[string]string{
			"custom_var": "{tool:upper} [{today_cost:short}]",
		},
		ColorRules: map[string]ColorRule{
			"custom_metric": {
				LowAt: 0, MediumAt: 40, HighAt: 75,
				LowColor: "#00FF00", MediumColor: "#FFFF00", HighColor: "#FF0000",
			},
		},
		Now:       time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC),
		ColorMode: ColorModeTruecolor,
		Glyphs:    GlyphTierUnicode,
	}

	tests := []struct {
		name     string
		template string
		wantSub  string
	}{
		{"tool upper", "{tool:upper}", "CLAUDE_CODE"},
		{"tool lower", "{tool:lower}", "claude_code"},
		{"provider bare", "{provider}", "claude_code"},
		{"account bare", "{account}", "work"},
		{"model bare", "{model}", "claude-3-7-sonnet"},
		{"today_cost short", "{today_cost:short}", "$12.35"},
		{"today_cost long", "{today_cost:long}", "$12.35 today"},
		{"today_cost money prec 1", "{today_cost:money:1}", "$12.3"},
		{"today_cost money prec 3", "{today_cost:money:3}", "$12.346"},
		{"block_cost long", "{block_cost:long}", "$5.50 block"},
		{"burn_rate long", "{burn_rate:long}", "$2.50/hr"},
		{"block_pct pct", "{block_pct:pct}", "45%"},
		{"block_pct pct prec 1", "{block_pct:pct:1}", "45.0%"},
		{"block_pct bar", "{block_pct:bar:10}", "▓▓▓▓▓░░░░░"},
		{"block_pct color", "{block_pct:color}", "#[fg=#A6E3A1]45#[default]"},
		{"tool icon unicode", "{tool:icon}", "\U0001F916"},
		{"tool brand", "{tool:brand}", "#[fg=#D97757]claude_code#[default]"},
		{"tokens modifier M", "{tokens_used:tokens}", "1.5M"},
		{"context tokens k", "{context_tokens:tokens}", "15k"},
		{"duration synthetic pass", "{block_remaining:duration}", "1h45m"},
		{"trunc 6", "{model:trunc:6}", "claud…"},
		{"pad right 15", "{tool:pad:15:r}", "claude_code    "},
		{"pad left 15", "{tool:pad:15:l}", "    claude_code"},
		{"default fallback used", "{nonexistent:default:fallback}", "fallback"},
		{"default ignored if set", "{tool:default:fallback}", "claude_code"},
		{"user variable expansion", "{custom_var}", "CLAUDE_CODE [$12.35]"},
		{"conditional truthy", "{?today_cost:Spent {today_cost:short}:Free}", "Spent $12.35"},
		{"conditional falsy", "{?nonexistent:Yes:No}", "No"},
		{"theme ref fg and bg", "#[fg=$accent,bg=$base]hello#[default]", "#[fg=#CBA6F7,bg=#1E1E2E]hello#[default]"},
		{"builtin segment cost", "{cost}", "$12.35"},
		{"builtin segment block", "{block}", "$5.50 block (1h45m)"},
		{"builtin segment burn", "{burn}", "$2.50/hr"},
		{"builtin segment tool", "{tool}", "claude_code"},
		{"builtin segment model", "{model}", "claude-3-7-sonnet"},
		{"builtin segment tokens", "{tokens}", "15k"},
		{"builtin segment context", "{context}", "15k (30%)"},
		{"builtin segment daily", "{daily}", "$12.35 today"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Render(tc.template, ctx)
			if err != nil {
				t.Fatalf("Render(%q) unexpected error: %v", tc.template, err)
			}
			if !strings.Contains(got, tc.wantSub) {
				t.Errorf("Render(%q) = %q; want substring %q", tc.template, got, tc.wantSub)
			}
		})
	}
}

func TestHappyPath_PreviewTranslation(t *testing.T) {
	input := "#[fg=#10A37F,bg=#1E1E2E,bold,underline]Active#[default] #[fg=colour208,dim,reverse]Alert#[none] #[fg=red,bg=blue]Plain"
	got := Preview(input)

	if !strings.Contains(got, "\x1b[38;2;16;163;127;48;2;30;30;46;1;4mActive") {
		t.Errorf("Preview missing RGB/bold/underline SGR: %q", got)
	}
	if !strings.Contains(got, "\x1b[0m") {
		t.Errorf("Preview missing reset SGR: %q", got)
	}
	if !strings.Contains(got, "\x1b[38;5;208;2;7mAlert") {
		t.Errorf("Preview missing 256/dim/reverse SGR: %q", got)
	}
	if !strings.Contains(got, "\x1b[31;44mPlain") {
		t.Errorf("Preview missing named ANSI fg/bg SGR: %q", got)
	}
}

func TestHappyPath_StatusRightRenderingAndInstallation(t *testing.T) {
	tempDir := t.TempDir()
	confPath := filepath.Join(tempDir, "tmux.conf")

	var buf bytes.Buffer
	opts := InstallOptions{
		Position:    "right",
		Preset:      "compact",
		Providers:   []string{"claude_code", "cursor"},
		Interval:    3,
		RightLength: 220,
		LeftLength:  90,
		BindPopup:   "u",
		BindRefresh: "r",
		Write:       true,
		Binary:      "agentusage",
		ConfPath:    confPath,
	}

	path, err := Install(&buf, opts)
	if err != nil {
		t.Fatalf("Install unexpected error: %v", err)
	}
	if path != confPath {
		t.Fatalf("Install returned path %q, want %q", path, confPath)
	}

	present, err := SentinelPresent(confPath)
	if err != nil || !present {
		t.Fatalf("SentinelPresent(%q) = (%v, %v); want (true, nil)", confPath, present, err)
	}

	data, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "set -g status-interval 3") {
		t.Errorf("missing status-interval 3: %s", content)
	}
	if !strings.Contains(content, "set -g status-right-length 220") {
		t.Errorf("missing status-right-length 220: %s", content)
	}
	if !strings.Contains(content, "bind-key u display-popup") {
		t.Errorf("missing popup binding: %s", content)
	}
	if !strings.Contains(content, "bind-key r run-shell") {
		t.Errorf("missing refresh binding: %s", content)
	}
	if !strings.Contains(content, "claude_code") || !strings.Contains(content, "cursor") {
		t.Errorf("missing multi-provider inner segments: %s", content)
	}
}

func TestHappyPath_JSONBuildAndWrite(t *testing.T) {
	now := time.Date(2026, 8, 31, 14, 0, 0, 0, time.UTC)
	ctx := Context{
		Provider: "claude_code",
		Account:  "personal",
		Snapshot: core.UsageSnapshot{
			ProviderID: "claude_code",
			AccountID:  "personal",
		},
		Block: report.Row{
			Cost:               4.50,
			BurnRateUSDPerHour: 1.20,
			TimeRemaining:      2 * time.Hour,
		},
		HaveBlock: true,
		Synthetic: map[string]string{"_block_remaining": "2h00m"},
		Now:       now,
	}
	detected := &DetectResult{
		Primary: "claude_code",
		Ordered: []string{"claude_code", "cursor"},
		Source:  "recency",
	}

	payload := BuildJSON(ctx, "$4.50 (2h00m)", detected)
	if payload.Provider != "claude_code" || payload.Account != "personal" {
		t.Errorf("BuildJSON bad metadata: %+v", payload)
	}
	if payload.Block == nil || payload.Block.Cost != 4.50 {
		t.Errorf("BuildJSON missing or incorrect block: %+v", payload.Block)
	}
	if payload.Rendered != "$4.50 (2h00m)" {
		t.Errorf("BuildJSON rendered mismatch: %q", payload.Rendered)
	}

	var buf bytes.Buffer
	if err := WriteJSON(&buf, payload); err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}

	var parsed JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("WriteJSON emitted invalid JSON: %v\nJSON:\n%s", err, buf.String())
	}
	if parsed.Provider != "claude_code" || parsed.Rendered != "$4.50 (2h00m)" {
		t.Errorf("Unmarshaled JSON mismatch: %+v", parsed)
	}
}

func TestHappyPath_ProviderIconsAndGlyphs(t *testing.T) {
	tiers := []GlyphTier{GlyphTierASCII, GlyphTierUnicode, GlyphTierNerdfont, GlyphTierCustomFont}
	for _, tier := range tiers {
		icon := ProviderIcon("claude_code", tier)
		if icon == "" {
			t.Errorf("ProviderIcon(claude_code, %s) returned empty string", tier)
		}
	}

	color := ProviderBrandColor("claude_code")
	if color != "#D97757" {
		t.Errorf("ProviderBrandColor(claude_code) = %q, want #D97757", color)
	}

	loRune, hiRune := IconCodepointRange()
	if loRune == 0 || hiRune == 0 || loRune > hiRune {
		t.Errorf("IconCodepointRange() = (%U, %U); invalid range", loRune, hiRune)
	}

	if IconFontFamily() != "OpenUsage Icons" {
		t.Errorf("IconFontFamily() = %q, want 'OpenUsage Icons'", IconFontFamily())
	}
	if IconFontVersion() == "" {
		t.Errorf("IconFontVersion() is empty")
	}

	// Bar glyphs
	fullA, emptyA := barGlyphs(GlyphTierASCII)
	if fullA != "#" || emptyA != "." {
		t.Errorf("barGlyphs(ASCII) = (%q, %q), want (#, .)", fullA, emptyA)
	}
	fullN, emptyN := barGlyphs(GlyphTierNerdfont)
	if fullN != "█" || emptyN != "░" {
		t.Errorf("barGlyphs(Nerdfont) = (%q, %q)", fullN, emptyN)
	}
	fullU, emptyU := barGlyphs(GlyphTierUnicode)
	if fullU != "▓" || emptyU != "░" {
		t.Errorf("barGlyphs(Unicode) = (%q, %q)", fullU, emptyU)
	}
}

// -----------------------------------------------------------------------------
// Axis 2: Edge / Boundary Cases
// -----------------------------------------------------------------------------

func TestEdge_EmptySnapshotAndNilMetrics(t *testing.T) {
	ctx := Context{
		Provider:  "",
		Account:   "",
		Snapshot:  core.UsageSnapshot{},
		ColorMode: ColorModeNone,
		Glyphs:    GlyphTierASCII,
	}

	res, err := Render("{tool} {cost} {burn} {tokens} {context} {model}", ctx)
	if err != nil {
		t.Fatalf("Render empty context failed: %v", err)
	}
	if strings.TrimSpace(res) != "" {
		t.Errorf("Expected empty/whitespace output for empty context, got %q", res)
	}

	// Zero tokens
	usedZero := 0.0
	ctx.Snapshot.Metrics = map[string]core.Metric{"context_tokens": {Used: &usedZero}}
	ctx.Provider = "claude_code"
	resTokens, err := Render("{context_tokens:tokens}", ctx)
	if err != nil {
		t.Fatalf("Render zero tokens failed: %v", err)
	}
	if resTokens != "0" {
		t.Errorf("Render zero tokens = %q, want '0'", resTokens)
	}
}

func TestEdge_SpecialCharactersAndLongNames(t *testing.T) {
	longName := strings.Repeat("VeryLongAccountName", 10)
	specialTool := "tool#with[brackets]and$dollars:colons\\slashes"
	ctx := Context{
		Provider:  specialTool,
		Account:   longName,
		ColorMode: ColorModeNone,
		Glyphs:    GlyphTierUnicode,
	}

	res, err := Render("{account:trunc:15} | {tool}", ctx)
	if err != nil {
		t.Fatalf("Render special characters failed: %v", err)
	}
	if !strings.HasSuffix(strings.Split(res, " | ")[0], "…") {
		t.Errorf("trunc:15 did not add ellipsis to long account name: %q", res)
	}
	if !strings.Contains(res, "##") {
		t.Errorf("sanitizeUserValue should have escaped '#' to '##': %q", res)
	}
}

func TestEdge_DurationFormattingBoundaries(t *testing.T) {
	durations := map[string]string{
		"":        "",
		"0s":      "0s",
		"30s":     "30s",
		"45m":     "45m",
		"1h0m":    "1h0m",
		"2h35m":   "2h35m",
		"120":     "2m",
		"7200":    "2h00m",
		"invalid": "invalid",
	}

	for in, want := range durations {
		got := modDuration(in)
		if got != want {
			t.Errorf("modDuration(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEdge_MoneyAndPctPrecisionBoundaries(t *testing.T) {
	ctx := Context{ColorMode: ColorModeNone}

	// Money precision boundaries: negative, 0, 8, >8 (falls back to default 2)
	moneyTests := []struct {
		val  string
		arg  string
		want string
	}{
		{"12.3456", "0", "$12"},
		{"12.3456", "1", "$12.3"},
		{"12.3456", "4", "$12.3456"},
		{"12.3456", "8", "$12.34560000"},
		{"12.3456", "-1", "$12.35"},
		{"12.3456", "99", "$12.35"},
		{"not-a-number", "2", "not-a-number"},
	}
	for _, tc := range moneyTests {
		ctx.Variables = map[string]string{"v": tc.val}
		got, err := Render(fmt.Sprintf("{v:money:%s}", tc.arg), ctx)
		if err != nil {
			t.Fatalf("money:%s error: %v", tc.arg, err)
		}
		if got != tc.want {
			t.Errorf("money:%s on %q = %q, want %q", tc.arg, tc.val, got, tc.want)
		}
	}

	// Pct precision boundaries: negative, 0, 4, >4 (falls back to default 0)
	pctTests := []struct {
		val  string
		arg  string
		want string
	}{
		{"75.6789", "0", "76%"},
		{"75.6789", "1", "75.7%"},
		{"75.6789", "3", "75.679%"},
		{"75.6789", "4", "75.6789%"},
		{"75.6789", "-1", "76%"},
		{"75.6789", "10", "76%"},
		{"invalid_text", "2", "invalid_text"},
	}
	for _, tc := range pctTests {
		ctx.Variables = map[string]string{"v": tc.val}
		got, err := Render(fmt.Sprintf("{v:pct:%s}", tc.arg), ctx)
		if err != nil {
			t.Fatalf("pct:%s error: %v", tc.arg, err)
		}
		if got != tc.want {
			t.Errorf("pct:%s on %q = %q, want %q", tc.arg, tc.val, got, tc.want)
		}
	}
}

func TestEdge_BarGraphBoundaries(t *testing.T) {
	// Negative clamped to 0, >100 clamped to 100, custom widths up to 64
	tests := []struct {
		val   string
		width string
		tier  GlyphTier
		want  string
	}{
		{"-10", "4", GlyphTierASCII, "...."},
		{"0", "4", GlyphTierASCII, "...."},
		{"50", "4", GlyphTierASCII, "##.."},
		{"100", "4", GlyphTierASCII, "####"},
		{"150", "4", GlyphTierASCII, "####"},
		{"50", "0", GlyphTierASCII, "####...."},   // default width 8
		{"50", "100", GlyphTierASCII, "####...."}, // width > 64 falls back to default 8
		{"bad", "4", GlyphTierASCII, "bad"},
	}
	for _, tc := range tests {
		got := modBar(tc.val, tc.width, tc.tier)
		if got != tc.want {
			t.Errorf("modBar(%q, %q, %s) = %q, want %q", tc.val, tc.width, tc.tier, got, tc.want)
		}
	}
}

func TestEdge_TruncAndPadEdgeCases(t *testing.T) {
	// Trunc edge cases: length 0, 1, exact, runes with multibyte characters
	multibyte := "日本語文字列"
	if modTrunc(multibyte, "0") != multibyte {
		t.Errorf("trunc 0 should return string unchanged")
	}
	if modTrunc(multibyte, "1") != "日" {
		t.Errorf("trunc 1 = %q, want '日'", modTrunc(multibyte, "1"))
	}
	if modTrunc(multibyte, "3") != "日本…" {
		t.Errorf("trunc 3 = %q, want '日本…'", modTrunc(multibyte, "3"))
	}
	if modTrunc(multibyte, "10") != multibyte {
		t.Errorf("trunc 10 should return full multibyte string")
	}

	// Pad edge cases: already wide, empty args, invalid args
	if modPad("hello", nil) != "hello" {
		t.Errorf("pad with nil args should return string unchanged")
	}
	if modPad("hello", []string{"0"}) != "hello" {
		t.Errorf("pad with 0 width should return string unchanged")
	}
	if modPad("hello", []string{"3"}) != "hello" {
		t.Errorf("pad with smaller width should return string unchanged")
	}
	if modPad("hello", []string{"8", "l"}) != "   hello" {
		t.Errorf("pad left 8 = %q, want '   hello'", modPad("hello", []string{"8", "l"}))
	}
	if modPad("hello", []string{"8", "r"}) != "hello   " {
		t.Errorf("pad right 8 = %q, want 'hello   '", modPad("hello", []string{"8", "r"}))
	}
}

// -----------------------------------------------------------------------------
// Axis 3: Negative Branches
// -----------------------------------------------------------------------------

func TestNegative_UnparseableAndMalformedTemplates(t *testing.T) {
	ctx := Context{ColorMode: ColorModeTruecolor}

	malformed := []struct {
		name     string
		template string
	}{
		{"unterminated tmux attr", "#[fg=red"},
		{"unterminated native tmux expr #(", "#(echo test"},
		{"unterminated native tmux expr #{", "#{window_name"},
		{"unterminated variable {", "{today_cost"},
		{"unknown modifier", "{today_cost:unrecognized_mod}"},
		{"malformed conditional", "{?invalid_cond}"},
	}

	for _, tc := range malformed {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Render(tc.template, ctx)
			if err == nil {
				t.Fatalf("Render(%q) expected error, got nil", tc.template)
			}
		})
	}
}

func TestNegative_DoctorAllChecksFailureAndWarnBranches(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Missing tmux binary
	var buf bytes.Buffer
	checkTmuxBinary(&buf, DoctorOptions{TmuxBinary: "/non/existent/tmux/binary"})
	if !strings.Contains(buf.String(), "[FAIL] tmux:") {
		t.Errorf("checkTmuxBinary missing binary report = %q", buf.String())
	}

	// 2. Mock tmux versions: old version < 3.0, 3.1, unparseable
	scriptDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(scriptDir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	createMockTmux := func(versionOutput string) string {
		path := filepath.Join(scriptDir, fmt.Sprintf("tmux_%d.sh", time.Now().UnixNano()))
		script := fmt.Sprintf("#!/bin/sh\necho %q\n", versionOutput)
		if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
		return path
	}

	buf.Reset()
	checkTmuxBinary(&buf, DoctorOptions{TmuxBinary: createMockTmux("tmux 2.8")})
	if !strings.Contains(buf.String(), "[WARN] tmux: tmux 2.8 (3.0+ recommended)") {
		t.Errorf("checkTmuxBinary version 2.8 report = %q", buf.String())
	}

	buf.Reset()
	checkTmuxBinary(&buf, DoctorOptions{TmuxBinary: createMockTmux("tmux 3.1c")})
	if !strings.Contains(buf.String(), "[ OK ] tmux: tmux 3.1c (3.2+ needed for --bind-popup)") {
		t.Errorf("checkTmuxBinary version 3.1 report = %q", buf.String())
	}

	buf.Reset()
	checkTmuxBinary(&buf, DoctorOptions{TmuxBinary: createMockTmux("tmux git-master-unparseable")})
	if !strings.Contains(buf.String(), "[WARN] tmux:") || !strings.Contains(buf.String(), "could not parse version") {
		t.Errorf("checkTmuxBinary unparseable report = %q", buf.String())
	}

	buf.Reset()
	checkTmuxBinary(&buf, DoctorOptions{TmuxBinary: createMockTmux("tmux 3.4")})
	if !strings.Contains(buf.String(), "[ OK ] tmux: tmux 3.4") {
		t.Errorf("checkTmuxBinary version 3.4 report = %q", buf.String())
	}

	// 3. Tmux environment check
	t.Setenv("TMUX", "")
	buf.Reset()
	checkTmuxEnv(&buf)
	if !strings.Contains(buf.String(), "[INFO] $TMUX: unset") {
		t.Errorf("checkTmuxEnv unset report = %q", buf.String())
	}
	t.Setenv("TMUX", "/tmp/tmux-1000/default,1234,0")
	buf.Reset()
	checkTmuxEnv(&buf)
	if !strings.Contains(buf.String(), "[ OK ] $TMUX: set") {
		t.Errorf("checkTmuxEnv set report = %q", buf.String())
	}

	// 4. Terminal Truecolor checks
	t.Setenv("COLORTERM", "truecolor")
	buf.Reset()
	checkTerminalTruecolor(&buf)
	if !strings.Contains(buf.String(), "[ OK ] terminal: truecolor") {
		t.Errorf("checkTerminalTruecolor truecolor report = %q", buf.String())
	}

	t.Setenv("COLORTERM", "")
	t.Setenv("TERM", "xterm-256color")
	buf.Reset()
	checkTerminalTruecolor(&buf)
	if !strings.Contains(buf.String(), "[INFO] terminal: 256-color") {
		t.Errorf("checkTerminalTruecolor 256 report = %q", buf.String())
	}

	t.Setenv("TERM", "vt100")
	buf.Reset()
	checkTerminalTruecolor(&buf)
	if !strings.Contains(buf.String(), "[WARN] terminal: COLORTERM unset") {
		t.Errorf("checkTerminalTruecolor fallback report = %q", buf.String())
	}

	// 5. Daemon check with missing socket, error socket, and healthy socket
	buf.Reset()
	checkDaemon(&buf, DoctorOptions{SocketPath: filepath.Join(tempDir, "missing.sock")})
	if !strings.Contains(buf.String(), "[INFO] daemon: not running") {
		t.Errorf("checkDaemon missing socket report = %q", buf.String())
	}

	// Mock daemon listener
	sockPath := filepath.Join(tempDir, "daemon.sock")
	ln, err := net.Listen("unix", sockPath)
	if err == nil {
		defer ln.Close()
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					return
				}
				resp := `{"status":"healthy","uptime":"10m"}`
				fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Length: %d\r\n\r\n%s", len(resp), resp)
				conn.Close()
			}
		}()
		buf.Reset()
		checkDaemon(&buf, DoctorOptions{SocketPath: sockPath})
		if buf.Len() == 0 {
			t.Errorf("checkDaemon listener report was empty")
		}
	}

	// 6. Active provider detection check
	buf.Reset()
	checkActiveProvider(&buf, DoctorOptions{Now: time.Now()})
	if buf.Len() == 0 {
		t.Errorf("checkActiveProvider report was empty")
	}

	// 7. Snippet check (present vs missing)
	emptyConf := filepath.Join(tempDir, "empty.conf")
	if err := os.WriteFile(emptyConf, []byte("# empty conf\n"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	buf.Reset()
	checkSnippet(&buf, DoctorOptions{ConfPath: emptyConf})
	if !strings.Contains(buf.String(), "has no agentusage block") {
		t.Errorf("checkSnippet missing block report = %q", buf.String())
	}

	withSnippetConf := filepath.Join(tempDir, "with_snippet.conf")
	snippet := BuildSnippet(InstallOptions{})
	if err := os.WriteFile(withSnippetConf, []byte(snippet), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	buf.Reset()
	checkSnippet(&buf, DoctorOptions{ConfPath: withSnippetConf})
	if !strings.Contains(buf.String(), "[ OK ] tmux.conf: agentusage block present") {
		t.Errorf("checkSnippet present report = %q", buf.String())
	}

	// 8. Full Run execution
	buf.Reset()
	if err := Run(&buf, DoctorOptions{ConfPath: withSnippetConf}); err != nil {
		t.Fatalf("Run unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "agentusage tmux doctor") || !strings.Contains(buf.String(), "done.") {
		t.Errorf("Run output incomplete:\n%s", buf.String())
	}
}

func TestNegative_PreviewMalformedAndUnknownDirectives(t *testing.T) {
	// 1. Malformed directive without closing bracket
	malformed := "prefix #[fg=red unclosed text"
	if got := Preview(malformed); got != malformed {
		t.Errorf("Preview malformed directive = %q, want %q", got, malformed)
	}

	// 2. Directive with unknown tokens / empty tokens
	unknown := "#[unknown_tok,,fg=invalid_color]Text#[default]"
	got := Preview(unknown)
	if !strings.Contains(got, "Text\x1b[0m") {
		t.Errorf("Preview unknown tokens = %q, want 'Text\\x1b[0m'", got)
	}

	// 3. colorTokenToSGR negative branches
	if _, ok := colorTokenToSGR("not-key-value"); ok {
		t.Errorf("colorTokenToSGR with no '=' should fail")
	}
	if _, ok := colorTokenToSGR("side=val"); ok {
		t.Errorf("colorTokenToSGR with unknown side should fail")
	}
	if _, ok := colorTokenToSGR("fg=invalid_hex_123"); ok {
		t.Errorf("colorTokenToSGR with invalid hex should fail")
	}
}

func TestNegative_JSONWriteError(t *testing.T) {
	payload := JSONOutput{Rendered: "test"}
	err := WriteJSON(errTestWriter{}, payload)
	if err == nil {
		t.Fatalf("WriteJSON with failing writer expected error, got nil")
	}
	if !strings.Contains(err.Error(), "tmux: encoding json output") {
		t.Errorf("WriteJSON error string mismatch: %v", err)
	}
}

func TestNegative_ManagedBlockUnbalancedMarkers(t *testing.T) {
	existing := []byte("first line\n# >>> start >>>\nsome content without end\n")
	cleaned := removeBlock(existing, "# >>> start >>>", "# <<< end <<<")
	if !bytes.Equal(cleaned, existing) {
		t.Errorf("removeBlock with unbalanced markers should return original slice, got %q", string(cleaned))
	}

	if blockPresent(existing, "# non-existent #") {
		t.Errorf("blockPresent returned true for non-existent marker")
	}
}

func TestNegative_TermSetupFallbackAndErrors(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("KITTY_CONFIG_DIRECTORY", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	// 1. kitty and ghostty directories not present
	if _, ok := setupKitty(); ok {
		t.Errorf("setupKitty should return false when kitty config dir does not exist")
	}
	if _, ok := setupGhostty(); ok {
		t.Errorf("setupGhostty should return false when ghostty config dir does not exist")
	}
	if _, ok := weztermGuidance(); ok {
		t.Errorf("weztermGuidance should return false when wezterm config does not exist")
	}

	// 2. Setup kitty config dir and test successful write and .bak creation
	kittyDir := filepath.Join(tempHome, ".config", "kitty")
	if err := os.MkdirAll(kittyDir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	kittyFile := filepath.Join(kittyDir, "kitty.conf")
	if err := os.WriteFile(kittyFile, []byte("# existing config\n"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	resK, ok := setupKitty()
	if !ok || resK.Action != "configured" {
		t.Fatalf("setupKitty result = (%+v, %v), want action=configured", resK, ok)
	}
	if _, err := os.Stat(kittyFile + ".bak"); err != nil {
		t.Errorf("kitty backup file .bak was not created: %v", err)
	}

	// 3. Setup ghostty config dir
	ghosttyDir := filepath.Join(tempHome, ".config", "ghostty")
	if err := os.MkdirAll(ghosttyDir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	resG, ok := setupGhostty()
	if !ok || resG.Action != "configured" {
		t.Fatalf("setupGhostty result = (%+v, %v), want action=configured", resG, ok)
	}

	// 4. Setup wezterm config file
	weztermFile := filepath.Join(tempHome, ".wezterm.lua")
	if err := os.WriteFile(weztermFile, []byte("-- wezterm config\n"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	resW, ok := weztermGuidance()
	if !ok || resW.Action != "manual" {
		t.Fatalf("weztermGuidance result = (%+v, %v), want action=manual", resW, ok)
	}

	// 5. Run full SetupTerminalFallback
	results := SetupTerminalFallback()
	if len(results) < 3 {
		t.Errorf("SetupTerminalFallback returned %d results, expected at least 3: %+v", len(results), results)
	}

	// 6. Test writeManagedBlock error on unwritable directory
	unwritablePath := filepath.Join(tempHome, "unwritable_dir", "dummy.conf")
	if err := os.WriteFile(filepath.Join(tempHome, "unwritable_dir"), []byte("file_blocking_dir"), 0o644); err == nil {
		if err := writeManagedBlock(unwritablePath, "block"); err == nil {
			t.Errorf("writeManagedBlock should have failed on invalid directory creation")
		}
	}
}

func TestNegative_WatchPIDFileAndRunnerErrors(t *testing.T) {
	tempDir := t.TempDir()

	// 1. PID file operations
	pidPath := filepath.Join(tempDir, "watch.pid")
	prev, err := WritePIDFile(pidPath)
	if err != nil || prev != 0 {
		t.Fatalf("WritePIDFile initial = (%d, %v), want (0, nil)", prev, err)
	}

	prev2, err := WritePIDFile(pidPath)
	if err != nil || prev2 != os.Getpid() {
		t.Fatalf("WritePIDFile overwrite = (%d, %v), want (%d, nil)", prev2, err, os.Getpid())
	}

	if _, err := WritePIDFile(""); err == nil {
		t.Fatalf("WritePIDFile with empty path should error")
	}

	// 2. Default PID file
	t.Setenv("HOME", tempDir)
	if DefaultPIDFile() == "" {
		t.Errorf("DefaultPIDFile() returned empty string")
	}

	// 3. realTmuxRunner error test
	if err := realTmuxRunner("non-existent-tmux-command-12345"); err == nil {
		t.Errorf("realTmuxRunner with non-existent command expected error, got nil")
	}

	// 4. evaluate with failing context
	state := alertState{}
	opts := WatchOptions{
		Runner: func(args ...string) error { return errors.New("tmux runner error") },
		Out:    &bytes.Buffer{},
		Now:    time.Now,
		Source: export.Source("invalid_source_path"),
		Alerts: config.TmuxAlerts{BurnRatePerHour: 5.0, BlockMinutesRemaining: 10},
	}
	evaluate(context.Background(), opts, AlertModeBoth, &state)

	// 5. Watch with canceled context
	ctx, cancel := context.WithCancel(context.Background())
	t.Setenv("TMUX", "1")
	cancel() // cancel immediately
	if err := Watch(ctx, WatchOptions{Runner: func(args ...string) error { return nil }, Out: io.Discard}); err != nil {
		t.Fatalf("Watch with canceled context should return nil, got: %v", err)
	}
}

func TestNegative_FontInstallAndUninstall(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_DATA_HOME", filepath.Join(tempHome, ".local", "share"))

	// 1. Initial status: not installed
	if FontInstalled() {
		t.Fatalf("FontInstalled should be false in fresh dir")
	}
	if FontNeedsUpdate() {
		t.Fatalf("FontNeedsUpdate should be false when not installed")
	}

	// 2. Install font
	path, err := InstallFont()
	if err != nil {
		t.Fatalf("InstallFont failed: %v", err)
	}
	if !FontInstalled() {
		t.Fatalf("FontInstalled should be true after InstallFont")
	}

	st := FontStatus()
	if !st.Installed || !st.UpToDate {
		t.Fatalf("FontStatus after install = %+v, want Installed=true, UpToDate=true", st)
	}

	// 3. Corrupt font on disk to test FontNeedsUpdate
	if err := os.WriteFile(path, []byte("corrupted font bytes"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if !FontNeedsUpdate() {
		t.Fatalf("FontNeedsUpdate should be true when installed font bytes differ from embedded")
	}

	// 4. Uninstall font
	uninstalledPath, err := UninstallFont()
	if err != nil {
		t.Fatalf("UninstallFont failed: %v", err)
	}
	if uninstalledPath != path {
		t.Errorf("UninstallFont path = %q, want %q", uninstalledPath, path)
	}
	if FontInstalled() {
		t.Fatalf("FontInstalled should be false after UninstallFont")
	}

	// 5. Uninstall non-existent font (should be no-op)
	if _, err := UninstallFont(); err != nil {
		t.Fatalf("UninstallFont on missing font should be nil, got: %v", err)
	}
}

// -----------------------------------------------------------------------------
// Axis 4: Concurrency / Race Conditions
// -----------------------------------------------------------------------------

func TestConcurrency_ConcurrentRenderAndFormatUnderRace(t *testing.T) {
	theme := ThemeColors{
		Base:   "#1E1E2E",
		Text:   "#CDD6F4",
		Accent: "#CBA6F7",
		Green:  "#A6E3A1",
		Yellow: "#F9E2AF",
		Red:    "#F38BA8",
	}
	used := func(v float64) *float64 { return &v }

	ctx := Context{
		Provider: "claude_code",
		Account:  "team",
		Snapshot: core.UsageSnapshot{
			ProviderID: "claude_code",
			AccountID:  "team",
			Metrics: map[string]core.Metric{
				"today_api_cost":  {Used: used(45.67)},
				"5h_block_cost":   {Used: used(12.34)},
				"usage_five_hour": {Used: used(88.0)},
			},
		},
		Theme:     theme,
		ThemeRefs: ThemeRefs(theme, ColorModeTruecolor),
		Variables: map[string]string{
			"summary": "{tool:upper} - {today_cost:short} ({block_pct:pct})",
		},
		ColorMode: ColorModeTruecolor,
		Glyphs:    GlyphTierUnicode,
		Now:       time.Now(),
	}

	const goroutines = 30
	const iterations = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Concurrent Render
				rendered, err := Render("{summary} | {block_pct:bar:8} | #[fg=$accent]Live#[default]", ctx)
				if err != nil {
					t.Errorf("goroutine %d Render failed: %v", id, err)
					return
				}
				if !strings.Contains(rendered, "CLAUDE_CODE") {
					t.Errorf("goroutine %d missing expected tool: %q", id, rendered)
					return
				}

				// Concurrent Preview
				ansi := Preview(rendered)
				if ansi == "" {
					t.Errorf("goroutine %d Preview returned empty string", id)
					return
				}

				// Concurrent JSON
				payload := BuildJSON(ctx, rendered, nil)
				var buf bytes.Buffer
				if err := WriteJSON(&buf, payload); err != nil {
					t.Errorf("goroutine %d WriteJSON failed: %v", id, err)
					return
				}

				// Concurrent Font & Glyphs inspection
				_, _ = IconCodepointRange()
				_ = ProviderIcon("claude_code", GlyphTierCustomFont)
				_ = ProviderBrandColor("cursor")
				_ = Presets()
				_ = AliasNames()
				_ = SegmentNames()
			}
		}(i)
	}

	wg.Wait()
}

// -----------------------------------------------------------------------------
// Axis 5: Domain Invariants
// -----------------------------------------------------------------------------

func TestDomainInvariants_WidthConstraintsAndEscaping(t *testing.T) {
	// 1. Escaping: \{, \}, \#, \$, \\, \n
	ctx := Context{
		ColorMode: ColorModeTruecolor,
		ThemeRefs: map[string]string{"red": "#FF0000"},
	}
	escapedTmpl := `Literal: \{escaped_braces\} \#escaped_hash \$escaped_dollar \\escaped_backslash\nNew line`
	got, err := Render(escapedTmpl, ctx)
	if err != nil {
		t.Fatalf("Render escaped template error: %v", err)
	}
	want := "Literal: {escaped_braces} #escaped_hash $escaped_dollar \\escaped_backslash\nNew line"
	if got != want {
		t.Errorf("Render escape sequences = %q, want %q", got, want)
	}

	// 2. Variable recursion cap
	cyclicCtx := Context{
		Variables: map[string]string{
			"a": "{b}",
			"b": "{c}",
			"c": "{d}",
			"d": "{e}",
			"e": "{a}", // cycle
		},
	}
	res, err := Render("{a}", cyclicCtx)
	if err != nil {
		t.Fatalf("Cyclic variable resolution should not crash or error infinitely: %v", err)
	}
	if res != "" && strings.Contains(res, "{") {
		// recursion terminates cleanly without infinite loop
	}

	// 3. Domain invariant: Truthiness evaluation
	truthyCases := map[string]bool{
		"":       false,
		" ":      false,
		"0":      false,
		"0.0":    false,
		"0.00":   false,
		"0.000":  false,
		"1":      true,
		"0.01":   true,
		"-5":     true,
		"active": true,
	}
	for in, wantTruthy := range truthyCases {
		gotTruthy := isTruthy(in)
		if gotTruthy != wantTruthy {
			t.Errorf("isTruthy(%q) = %v, want %v", in, gotTruthy, wantTruthy)
		}
	}

	// 4. Domain invariant: Aliases and segments registry
	aliases := AliasNames()
	if len(aliases) < 10 {
		t.Errorf("AliasNames returned %d entries, expected >= 10", len(aliases))
	}
	for _, a := range []string{"today_cost", "block_pct", "burn_rate", "context_pct"} {
		if !loContains(aliases, a) {
			t.Errorf("AliasNames missing key alias %q", a)
		}
	}

	segments := SegmentNames()
	if len(segments) < 8 {
		t.Errorf("SegmentNames returned %d entries, expected >= 8", len(segments))
	}
	for _, s := range []string{"cost", "block", "burn", "tool", "tokens", "context", "daily", "active_tools"} {
		if !loContains(segments, s) {
			t.Errorf("SegmentNames missing key segment %q", s)
		}
	}
}

func TestDomainInvariants_HelperFunctions(t *testing.T) {
	// ParseTmuxVersion
	if major, minor, ok := parseTmuxVersion("tmux 3.4"); !ok || major != 3 || minor != 4 {
		t.Errorf("parseTmuxVersion(tmux 3.4) = (%d, %d, %v), want (3, 4, true)", major, minor, ok)
	}
	if _, _, ok := parseTmuxVersion("unparseable"); ok {
		t.Errorf("parseTmuxVersion(unparseable) expected ok=false")
	}

	// Context Window guesses
	if w := contextWindowFor("claude-3-5-sonnet-20241022", 50000); w != 200_000 {
		t.Errorf("contextWindowFor default = %d, want 200_000", w)
	}
	if w := contextWindowFor("claude-3-opus-1m", 50000); w != 1_000_000 {
		t.Errorf("contextWindowFor 1m = %d, want 1_000_000", w)
	}
	if w := contextWindowFor("claude-3-5-sonnet", 250_000); w != 1_000_000 {
		t.Errorf("contextWindowFor large observed = %d, want 1_000_000", w)
	}

	// Formatter helpers: fmtMoneyDefault and fmtDurationDefault
	if fmtMoneyDefault(0) != "" || fmtMoneyDefault(-5) != "" {
		t.Errorf("fmtMoneyDefault non-positive should be empty")
	}
	if fmtMoneyDefault(12.345) != "$12.35" {
		t.Errorf("fmtMoneyDefault(12.345) = %q, want '$12.35'", fmtMoneyDefault(12.345))
	}

	if fmtDurationDefault(0) != "" || fmtDurationDefault(-time.Minute) != "" {
		t.Errorf("fmtDurationDefault non-positive should be empty")
	}
	if fmtDurationDefault(45*time.Minute) != "45m" {
		t.Errorf("fmtDurationDefault(45m) = %q, want '45m'", fmtDurationDefault(45*time.Minute))
	}
	if fmtDurationDefault(2*time.Hour+15*time.Minute) != "2h15m" {
		t.Errorf("fmtDurationDefault(2h15m) = %q, want '2h15m'", fmtDurationDefault(2*time.Hour+15*time.Minute))
	}

	// Builtin segments with various context configurations
	usedVal := 10.0
	r := &renderer{
		ctx: Context{
			Provider: "claude_code",
			Snapshot: core.UsageSnapshot{
				ProviderID: "claude_code",
				Metrics:    map[string]core.Metric{"today_api_cost": {Used: &usedVal}},
				Attributes: map[string]string{"model": "test-model"},
			},
			AllSnapshots: []core.UsageSnapshot{
				{ProviderID: "claude_code"},
				{ProviderID: "cursor"},
			},
		},
	}
	if segCost(r) != "$10.00" {
		t.Errorf("segCost = %q, want '$10.00'", segCost(r))
	}
	if segTool(r) != "claude_code" {
		t.Errorf("segTool = %q, want 'claude_code'", segTool(r))
	}
	if segModel(r) != "test-model" {
		t.Errorf("segModel = %q, want 'test-model'", segModel(r))
	}
	if segActiveTools(r) != "claude_code | cursor" {
		t.Errorf("segActiveTools = %q, want 'claude_code | cursor'", segActiveTools(r))
	}

	// BuildContext exercises with daemon fallback source
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	bctx, err := BuildContext(ctxTimeout, BuildOptions{
		Source:               export.SourceDaemon,
		Provider:             "claude_code",
		Candidates:           []string{"claude_code", "cursor"},
		Now:                  time.Now(),
		OfflineClaudePricing: true,
	})
	if err != nil {
		t.Fatalf("BuildContext unexpected error: %v", err)
	}
	if bctx.ColorMode != ColorModeTruecolor {
		t.Errorf("BuildContext default color mode = %s, want truecolor", bctx.ColorMode)
	}
}

func loContains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
