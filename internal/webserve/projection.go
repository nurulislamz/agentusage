package webserve

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/tui"
)

// forceTrueColor ensures lipgloss emits ANSI even when stdout is not a TTY
// (e.g. `openusage serve` under systemd / as a background HTTP process).
var forceTrueColorOnce sync.Once

func ensureTrueColor() {
	forceTrueColorOnce.Do(func() {
		lipgloss.SetColorProfile(termenv.TrueColor)
	})
}

const (
	webDetailWidth = 96
	webStripWidth  = 5
)
)

func buildViews(opts Options, meta collectorMeta, snaps []core.UsageSnapshot) ([]AccountView, ThemeTokens) {
	ensureTrueColor()
	cfg := configOrDefault(opts)
	if theme := strings.TrimSpace(cfg.Theme); theme != "" {
		_ = tui.ThemeTokensForName(theme)
	} else if theme := strings.TrimSpace(meta.theme); theme != "" {
		_ = tui.ThemeTokensForName(theme)
	}

	projector := tui.NewWebProjectorFromConfig(cfg)
	if opts.Now != nil {
		projector.Now = opts.Now()
	}
	projector.DetailWidth = webDetailWidth

	ordered := projector.OrderSnapshots(tui.SnapshotsToMap(snaps))
	if len(ordered) == 0 {
		ordered = snaps
	}

	names := catalogNames(meta)
	views := projector.ProjectSnapshots(ordered, names)
	out := make([]AccountView, len(views))
	for i, v := range views {
		out[i] = enrichAccountView(opts, cfg, orderedMatch(ordered, v.AccountID), accountViewFromTUI(v))
		frame := renderTUIFrame(cfg, ordered, i, defaultFrameWidth, defaultFrameHeight)
		if frame != "" {
			out[i].FrameHTML = ANSIToHTML(frame)
		}
	}
	tokens := tui.WebThemeTokensFromTheme(tui.ActiveTheme())
	return out, themeTokensFromTUI(tokens)
}

func orderedMatch(ordered []core.UsageSnapshot, accountID string) core.UsageSnapshot {
	for _, snap := range ordered {
		if snap.AccountID == accountID {
			return snap
		}
	}
	return core.UsageSnapshot{}
}

func enrichAccountView(opts Options, cfg config.Config, snap core.UsageSnapshot, view AccountView) AccountView {
	if snap.AccountID == "" {
		return view
	}

	now := time.Now()
	if opts.Now != nil {
		now = opts.Now()
	}

	warn := cfg.UI.WarnThreshold
	if warn <= 0 {
		warn = 0.25
	}
	crit := cfg.UI.CritThreshold
	if crit <= 0 {
		crit = 0.1
	}
	usageMode := cfg.Dashboard.UsageMode
	if usageMode == "" {
		usageMode = config.UsageModeRemaining
	}
	tw := core.ParseTimeWindow(cfg.Data.TimeWindow)
	hideCosts := resolveHideCosts(cfg, snap)

	detail := tui.RenderDetailContent(
		snap,
		now,
		webDetailWidth,
		warn,
		crit,
		0,
		tw,
		hideCosts,
		usageMode,
	)
	view.DetailHTML = ANSIToHTML(detail)
	view.BadgeHTML = ANSIToHTML(tui.SnapshotStatusBadge(snap))

	status := core.EffectiveStatus(snap)
	iconColor := string(tui.StatusColor(status))
	view.Status = string(status)
	view.StatusIcon = tui.StatusIcon(status)
	view.IconHTML = ANSIToHTML(
		lipgloss.NewStyle().Foreground(lipgloss.Color(iconColor)).Render(view.StatusIcon),
	)

	if view.GaugePercent >= 0 {
		fill := listFillColor(view.GaugePercent, usageMode == config.UsageModeUsed, warn, crit, status, tui.ActiveTheme())
		strip := tui.RenderCompactBlockStrip(view.GaugePercent, webStripWidth, lipgloss.Color(fill))
		view.StripHTML = ANSIToHTML(strip)
		view.SummaryHTML = ANSIToHTML(
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(fill)).Render(view.Summary),
		)
	} else if view.Summary != "" {
		view.SummaryHTML = ANSIToHTML(
			lipgloss.NewStyle().Bold(true).Foreground(tui.ActiveTheme().Text).Render(view.Summary),
		)
	}

	if len(view.Resets) > 0 {
		r := view.Resets[0]
		view.ResetHint = fmt.Sprintf("Resets in %s", r.Duration)
		if r.Label != "" {
			view.ResetHint = fmt.Sprintf("%s resets in %s", r.Label, r.Duration)
		}
	}

	return view
}

func resolveHideCosts(cfg config.Config, snap core.UsageSnapshot) bool {
	if cfg.Dashboard.HideCosts != nil {
		return *cfg.Dashboard.HideCosts
	}
	for _, p := range cfg.Dashboard.Providers {
		if p.AccountID == snap.AccountID && p.HideCosts != nil {
			return *p.HideCosts
		}
	}
	return false
}

func listFillColor(gaugePercent float64, usedMode bool, warn, crit float64, status core.Status, theme tui.Theme) string {
	if gaugePercent < 0 {
		switch status {
		case core.StatusLimited:
			return string(theme.Peach)
		case core.StatusError:
			return string(theme.Red)
		case core.StatusAuth:
			return string(theme.Yellow)
		default:
			return string(theme.Text)
		}
	}
	if usedMode {
		critCutoff := (1 - crit) * 100
		warnCutoff := (1 - warn) * 100
		if warnCutoff > 75 {
			warnCutoff = 75
		}
		if critCutoff < 90 {
			critCutoff = 90
		}
		switch {
		case gaugePercent >= critCutoff:
			return string(theme.Red)
		case gaugePercent >= warnCutoff:
			return string(theme.Peach)
		case gaugePercent >= 50:
			return string(theme.Yellow)
		default:
			return string(theme.Green)
		}
	}
	critCutoff := crit * 100
	warnCutoff := warn * 100
	if warnCutoff < 25 {
		warnCutoff = 25
	}
	if critCutoff > 10 {
		critCutoff = 10
	}
	switch {
	case gaugePercent <= critCutoff:
		return string(theme.Red)
	case gaugePercent <= warnCutoff:
		return string(theme.Peach)
	case gaugePercent <= 50:
		return string(theme.Yellow)
	default:
		return string(theme.Green)
	}
}

func configOrDefault(opts Options) config.Config {
	if opts.Config != nil {
		return *opts.Config
	}
	cfg := config.DefaultConfig()
	if opts.TimeWindow != "" {
		cfg.Data.TimeWindow = opts.TimeWindow
	}
	if opts.Theme != "" {
		cfg.Theme = opts.Theme
	}
	if opts.UsageMode != "" {
		cfg.Dashboard.UsageMode = opts.UsageMode
	}
	if opts.WarnThreshold > 0 {
		cfg.UI.WarnThreshold = opts.WarnThreshold
	}
	if opts.CritThreshold > 0 {
		cfg.UI.CritThreshold = opts.CritThreshold
	}
	if opts.RefreshSeconds > 0 {
		cfg.UI.RefreshIntervalSeconds = opts.RefreshSeconds
	}
	return cfg
}

func catalogNames(meta collectorMeta) map[string]string {
	names := make(map[string]string)
	for _, c := range meta.catalog {
		names[c.ID] = c.Name
	}
	return names
}

func accountViewFromTUI(v tui.WebAccountView) AccountView {
	sections := make([]DetailSection, len(v.DetailSections))
	for i, s := range v.DetailSections {
		sections[i] = DetailSection{Title: s.Title, Icon: s.Icon, Lines: s.Lines}
	}
	resets := make([]ResetPill, len(v.Resets))
	for i, r := range v.Resets {
		resets[i] = ResetPill{Label: r.Label, Duration: r.Duration, Urgent: r.Urgent}
	}
	return AccountView{
		Key:            v.Key,
		ProviderID:     v.ProviderID,
		ProviderName:   v.ProviderName,
		AccountID:      v.AccountID,
		Status:         v.Status,
		StatusBadge:    v.StatusBadge,
		StatusIcon:     v.StatusIcon,
		AccentColor:    v.AccentColor,
		Summary:        v.Summary,
		Detail:         v.Detail,
		TagEmoji:       v.TagEmoji,
		TagLabel:       v.TagLabel,
		GaugePercent:   v.GaugePercent,
		Message:        v.Message,
		Timestamp:      v.Timestamp,
		TileLines:      v.TileLines,
		DetailSections: sections,
		Resets:         resets,
		DailyCost:      v.DailyCost,
	}
}

func themeTokensFromTUI(t tui.WebThemeTokens) ThemeTokens {
	return ThemeTokens{
		Name: t.Name, Icon: t.Icon, Base: t.Base, Mantle: t.Mantle,
		Surface0: t.Surface0, Surface1: t.Surface1, Surface2: t.Surface2,
		Text: t.Text, Subtext: t.Subtext, Dim: t.Dim,
		Accent: t.Accent, Blue: t.Blue, Sapphire: t.Sapphire,
		Green: t.Green, Yellow: t.Yellow, Red: t.Red,
		Peach: t.Peach, Teal: t.Teal, Lavender: t.Lavender, Mauve: t.Mauve,
	}
}
