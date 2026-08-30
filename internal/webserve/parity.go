package webserve

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/tui"
)

var tuiQuotaRe = regexp.MustCompile(`(?i)([\d.]+)%\s+(remaining|used)`)

// ParityIssue is one TUI vs web information mismatch.
type ParityIssue struct {
	AccountID string
	Field     string
	Detail    string
}

func (i ParityIssue) String() string {
	if i.AccountID == "" {
		return fmt.Sprintf("%s: %s", i.Field, i.Detail)
	}
	return fmt.Sprintf("%s %s: %s", i.AccountID, i.Field, i.Detail)
}

// VerifyServeParity collects the same envelope the web port serves and
// compares it to TUI-rendered detail for those snapshots.
func VerifyServeParity(opts Options) (Envelope, []ParityIssue, error) {
	c := newCollector(opts)
	env, err := c.envelope()
	if err != nil {
		return Envelope{}, nil, err
	}
	return env, VerifyTUIWebParity(opts, env), nil
}

// VerifyTUIWebParity checks that web views carry the same account list,
// badges, summaries, section titles, quota percents, and timer labels as the
// Bubble Tea detail panel for the same snapshots. It does not require glyph
// parity (bars, sparklines, box drawing).
func VerifyTUIWebParity(opts Options, env Envelope) []ParityIssue {
	cfg := configOrDefault(opts)
	if tw := strings.TrimSpace(env.TimeWindow); tw != "" {
		cfg.Data.TimeWindow = tw
	}
	if mode := strings.TrimSpace(env.UsageMode); mode != "" {
		cfg.Dashboard.UsageMode = mode
	}
	now := time.Now()
	if opts.Now != nil {
		now = opts.Now()
	}

	projector := tui.NewWebProjectorFromConfig(cfg)
	projector.Now = now
	projector.DetailWidth = webDetailWidth

	ordered := projector.OrderSnapshots(tui.SnapshotsToMap(env.Snapshots))
	issues := make([]ParityIssue, 0)
	if len(ordered) != len(env.Views) {
		issues = append(issues, ParityIssue{
			Field:  "accounts",
			Detail: fmt.Sprintf("tui order has %d accounts, web has %d", len(ordered), len(env.Views)),
		})
	}
	n := len(ordered)
	if len(env.Views) < n {
		n = len(env.Views)
	}
	for i := 0; i < n; i++ {
		if ordered[i].AccountID != env.Views[i].AccountID {
			issues = append(issues, ParityIssue{
				Field:  "order",
				Detail: fmt.Sprintf("index %d tui=%q web=%q", i, ordered[i].AccountID, env.Views[i].AccountID),
			})
		}
	}

	snaps := tui.SnapshotsToMap(env.Snapshots)
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

	for _, view := range env.Views {
		snap, ok := snaps[view.AccountID]
		if !ok {
			issues = append(issues, ParityIssue{AccountID: view.AccountID, Field: "snapshot", Detail: "missing from envelope snapshots"})
			continue
		}
		issues = append(issues, compareAccount(view, snap, now, warn, crit, tw, resolveHideCosts(cfg, snap), usageMode)...)
	}
	return issues
}

func compareAccount(
	view AccountView,
	snap core.UsageSnapshot,
	now time.Time,
	warn, crit float64,
	tw core.TimeWindow,
	hideCosts bool,
	usageMode string,
) []ParityIssue {
	issues := make([]ParityIssue, 0)
	id := view.AccountID
	tuiDetail := collapsePlain(tui.RenderDetailContent(
		snap, now, webDetailWidth, warn, crit, 0, tw, hideCosts, usageMode,
	))
	tuiBadge := collapsePlain(tui.SnapshotStatusBadge(snap))

	if view.AccountID != "" && !strings.Contains(tuiDetail, view.AccountID) {
		issues = append(issues, ParityIssue{AccountID: id, Field: "account_id", Detail: "not in TUI detail header"})
	}
	if tuiBadge != "" && collapsePlain(view.StatusBadge) != tuiBadge {
		issues = append(issues, ParityIssue{
			AccountID: id, Field: "status_badge",
			Detail: fmt.Sprintf("tui=%q web=%q", tuiBadge, view.StatusBadge),
		})
	}
	if view.Summary != "" && !strings.Contains(tuiDetail, collapsePlain(view.Summary)) {
		issues = append(issues, ParityIssue{
			AccountID: id, Field: "summary",
			Detail: fmt.Sprintf("web %q missing from TUI detail", view.Summary),
		})
	}
	if view.CycleSchedule != "" && !strings.Contains(tuiDetail, collapsePlain(view.CycleSchedule)) {
		issues = append(issues, ParityIssue{
			AccountID: id, Field: "cycle_schedule",
			Detail: fmt.Sprintf("web %q missing from TUI detail", view.CycleSchedule),
		})
	}

	tuiQuotas := quotaBag(tuiDetail)
	webQuotas := make(map[string]int)
	for _, card := range view.DetailCards {
		for _, row := range card.Rows {
			switch row.Kind {
			case "gauge":
				if row.Label != "" && !strings.Contains(tuiDetail, collapsePlain(row.Label)) {
					issues = append(issues, ParityIssue{
						AccountID: id, Field: "gauge_label",
						Detail: fmt.Sprintf("%q missing from TUI", row.Label),
					})
				}
				if row.Percent != nil {
					key := formatQuotaKey(*row.Percent, usageMode)
					webQuotas[key]++
					pct := formatPct(*row.Percent)
					if !strings.Contains(tuiDetail, pct) {
						issues = append(issues, ParityIssue{
							AccountID: id, Field: "gauge_percent",
							Detail: fmt.Sprintf("%s %s missing from TUI", row.Label, pct),
						})
					}
				}
			case "timer":
				if row.Label != "" && !strings.Contains(tuiDetail, collapsePlain(row.Label)) {
					issues = append(issues, ParityIssue{
						AccountID: id, Field: "timer_label",
						Detail: fmt.Sprintf("%q missing from TUI", row.Label),
					})
				}
				if row.Value != "" && !strings.Contains(tuiDetail, collapsePlain(row.Value)) {
					issues = append(issues, ParityIssue{
						AccountID: id, Field: "timer_when",
						Detail: fmt.Sprintf("%q missing from TUI", row.Value),
					})
				}
			case "heading", "text", "kv":
				chunk := collapsePlain(strings.TrimSpace(row.Value))
				if chunk == "" {
					chunk = collapsePlain(strings.TrimSpace(row.Label))
				}
				if chunk == "" {
					continue
				}
				if !strings.Contains(tuiDetail, chunk) {
					issues = append(issues, ParityIssue{
						AccountID: id, Field: "row",
						Detail: fmt.Sprintf("web %s %q missing from TUI", row.Kind, truncate(chunk, 80)),
					})
				}
			}
		}
	}

	for key, n := range tuiQuotas {
		if webQuotas[key] < n {
			issues = append(issues, ParityIssue{
				AccountID: id, Field: "quota",
				Detail: fmt.Sprintf("TUI has %d× %s, web gauges have %d", n, key, webQuotas[key]),
			})
		}
	}
	return issues
}

func quotaBag(plain string) map[string]int {
	out := make(map[string]int)
	for _, m := range tuiQuotaRe.FindAllStringSubmatch(plain, -1) {
		pct, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			continue
		}
		out[formatQuotaKey(pct, m[2])]++
	}
	return out
}

func formatQuotaKey(pct float64, mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != "used" {
		mode = "remaining"
	}
	return formatPct(pct) + " " + mode
}

func formatPct(pct float64) string {
	return strconv.FormatFloat(pct, 'f', 2, 64) + "%"
}

func collapsePlain(s string) string {
	return strings.Join(strings.Fields(tui.StripANSI(s)), " ")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
