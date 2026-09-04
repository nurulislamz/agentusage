package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/daemon"
	"github.com/nurulislamz/agentusage/internal/providers"
)

type PoolMetric struct {
	Limit     *float64   `json:"limit,omitempty"`
	Used      *float64   `json:"used,omitempty"`
	Remaining *float64   `json:"remaining,omitempty"`
	Unit      string     `json:"unit,omitempty"`
	Window    string     `json:"window,omitempty"`
	ResetsAt  *time.Time `json:"resets_at,omitempty"`
	ResetsIn  string     `json:"resets_in,omitempty"`
}

type GetResponse struct {
	ID        string                `json:"id"`
	Provider  string                `json:"provider"`
	Status    string                `json:"status"`
	Window    string                `json:"window"`
	Limit     *float64              `json:"limit,omitempty"`
	Used      *float64              `json:"used,omitempty"`
	Remaining *float64              `json:"remaining,omitempty"`
	Unit      string                `json:"unit,omitempty"`
	ResetsAt  *time.Time            `json:"resets_at,omitempty"`
	ResetsIn  string                `json:"resets_in,omitempty"`
	Pools     map[string]PoolMetric `json:"pools,omitempty"`
	Message   string                `json:"message,omitempty"`
}

type getOptions struct {
	window  string
	format  string
	plain   bool
	timeout time.Duration
}

func newGetCommand() *cobra.Command {
	var opts getOptions

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get the usage left and limits for a box or account (5h limit by default)",
		Long: `Retrieve the current quota and usage left for a specified box or account ID.

Defaults to the 5-hour usage limit window.
Output is formatted as JSON by default for seamless AI deserialization and scripting.
Pass --plain to get just the remaining percentage (e.g. '85%') for shell prompts or tmux.`,
		Example: strings.Join([]string{
			"  agentusage get antigravity-nurulz",
			"  agu get nurulz",
			"  agu get antigravity-nurulz --plain",
			"  agu get antigravity-nurulz --window weekly",
			"  agu get antigravity-nurulz --format table",
		}, "\n"),
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runGet(os.Stdout, args[0], opts)
		},
	}

	fl := cmd.Flags()
	fl.StringVarP(&opts.window, "window", "w", "5h", "usage window: 5h (default), weekly, session, or all")
	fl.StringVar(&opts.format, "format", "json", "output format: json (default), plain, or table")
	fl.BoolVarP(&opts.plain, "plain", "p", false, "output remaining percentage only (e.g. 85%)")
	fl.DurationVar(&opts.timeout, "timeout", 10*time.Second, "fetch timeout")

	return cmd
}

func runGet(out io.Writer, id string, opts getOptions) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	accounts := daemon.ResolveAccounts(&cfg)
	acct, ok := findAccount(accounts, id)
	if !ok {
		return fmt.Errorf("unknown account or box %q; run 'agentusage list' (or 'agu list') to see available IDs", id)
	}

	timeout := opts.timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	snap, err := fetchAccountSnapshot(ctx, acct, cfg)
	if err != nil {
		return fmt.Errorf("fetch usage for %s (%s): %w", acct.ID, acct.Provider, err)
	}

	resp := buildGetResponse(acct, snap, opts.window)

	format := strings.ToLower(strings.TrimSpace(opts.format))
	if opts.plain {
		format = "plain"
	}
	if format == "" {
		format = "json"
	}

	switch format {
	case "json":
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(resp)

	case "plain":
		if resp.Remaining != nil {
			unit := resp.Unit
			if unit == "" {
				unit = "%"
			}
			fmt.Fprintf(out, "%.0f%s\n", *resp.Remaining, unit)
		} else if resp.Message != "" {
			fmt.Fprintln(out, resp.Message)
		} else {
			fmt.Fprintln(out, "unknown")
		}
		return nil

	case "table":
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "Account ID:\t%s\n", resp.ID)
		fmt.Fprintf(w, "Provider:\t%s\n", resp.Provider)
		fmt.Fprintf(w, "Status:\t%s\n", resp.Status)
		fmt.Fprintf(w, "Window:\t%s\n", resp.Window)
		if resp.Remaining != nil {
			fmt.Fprintf(w, "Remaining:\t%.1f%s\n", *resp.Remaining, resp.Unit)
		}
		if resp.Used != nil {
			fmt.Fprintf(w, "Used:\t%.1f%s\n", *resp.Used, resp.Unit)
		}
		if resp.ResetsIn != "" {
			fmt.Fprintf(w, "Resets In:\t%s\n", resp.ResetsIn)
		}
		if len(resp.Pools) > 0 {
			fmt.Fprintln(w, "\nPools:")
			for poolName, p := range resp.Pools {
				remStr := "-"
				if p.Remaining != nil {
					remStr = fmt.Sprintf("%.1f%s", *p.Remaining, p.Unit)
				}
				fmt.Fprintf(w, "  %s:\t%s remaining\t(resets in %s)\n", poolName, remStr, p.ResetsIn)
			}
		}
		return w.Flush()

	default:
		return fmt.Errorf("unknown format %q (supported: json, plain, table)", format)
	}
}

func findAccount(accounts []core.AccountConfig, target string) (core.AccountConfig, bool) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return core.AccountConfig{}, false
	}

	// 1. Exact ID match
	for _, a := range accounts {
		if strings.EqualFold(a.ID, trimmed) {
			return a, true
		}
	}

	// 2. Antigravity box ID match: e.g. "nurulz" -> "antigravity-nurulz"
	for _, a := range accounts {
		if strings.EqualFold(a.ID, "antigravity-"+trimmed) {
			return a, true
		}
	}

	// 3. Antigravity short-name alias: e.g. "agy-chaos" -> "antigravity-chaos"
	if strings.HasPrefix(strings.ToLower(trimmed), "agy-") {
		boxName := trimmed[4:]
		for _, a := range accounts {
			if strings.EqualFold(a.ID, "antigravity-"+boxName) {
				return a, true
			}
		}
	}

	// 4. Slash or colon delimiter alias: e.g. "cursor/physics" -> "cursor-physics"
	if strings.ContainsAny(trimmed, "/:") {
		normalized := strings.ReplaceAll(trimmed, "/", "-")
		normalized = strings.ReplaceAll(normalized, ":", "-")
		if acct, ok := findAccount(accounts, normalized); ok {
			return acct, true
		}
	}

	// 5. Antigravity box hint match
	for _, a := range accounts {
		if a.Provider == "antigravity" {
			if box := a.Hint("box_name", ""); box != "" && strings.EqualFold(box, trimmed) {
				return a, true
			}
		}
	}

	// 6. Other box name match: e.g. "nurulz" matching "cursor-nurulz" or "opencode-nurulz"
	for _, a := range accounts {
		if box := a.Hint("box_name", ""); box != "" && strings.EqualFold(box, trimmed) {
			return a, true
		}
		if strings.EqualFold(strings.TrimPrefix(a.ID, a.Provider+"-"), trimmed) {
			return a, true
		}
	}

	// 5. Case-insensitive prefix match
	lower := strings.ToLower(trimmed)
	var matches []core.AccountConfig
	for _, a := range accounts {
		if strings.HasPrefix(strings.ToLower(a.ID), lower) {
			matches = append(matches, a)
		}
	}
	if len(matches) == 1 {
		return matches[0], true
	}

	return core.AccountConfig{}, false
}

func fetchAccountSnapshot(ctx context.Context, acct core.AccountConfig, cfg config.Config) (core.UsageSnapshot, error) {
	for _, p := range providers.AllProviders() {
		if strings.EqualFold(p.ID(), acct.Provider) {
			snap, err := p.Fetch(ctx, acct)
			if err != nil {
				return snap, err
			}
			snap = core.NormalizeUsageSnapshotWithConfig(snap, cfg.ModelNormalization)
			return snap, nil
		}
	}
	return core.UsageSnapshot{}, fmt.Errorf("no provider adapter registered for %q", acct.Provider)
}

func buildGetResponse(acct core.AccountConfig, snap core.UsageSnapshot, requestedWindow string) GetResponse {
	resp := GetResponse{
		ID:       acct.ID,
		Provider: acct.Provider,
		Status:   string(snap.Status),
		Window:   requestedWindow,
		Message:  snap.Message,
		Pools:    make(map[string]PoolMetric),
	}
	if resp.Status == "" {
		resp.Status = "ok"
	}

	now := time.Now().UTC()
	targetWindow := strings.ToLower(strings.TrimSpace(requestedWindow))
	if targetWindow == "" {
		targetWindow = "5h"
	}

	// Collect pool metrics matching the requested window
	keys := make([]string, 0, len(snap.Metrics))
	for k := range snap.Metrics {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		m := snap.Metrics[k]
		kLower := strings.ToLower(k)

		isTargetWindow := false
		if targetWindow == "all" {
			isTargetWindow = true
		} else if targetWindow == "5h" {
			isTargetWindow = m.Window == "5h" || strings.Contains(kLower, "5h")
		} else if targetWindow == "weekly" || targetWindow == "7d" {
			isTargetWindow = m.Window == "7d" || strings.Contains(kLower, "weekly") || strings.Contains(kLower, "7d")
		} else {
			isTargetWindow = m.Window == targetWindow || strings.Contains(kLower, targetWindow)
		}

		if !isTargetWindow {
			continue
		}

		poolName := strings.TrimPrefix(k, "quota_")
		pm := PoolMetric{
			Limit:     m.Limit,
			Used:      m.Used,
			Remaining: m.Remaining,
			Unit:      m.Unit,
			Window:    m.Window,
		}

		// Find matching reset
		resetTime := findResetTime(snap.Resets, k)
		if !resetTime.IsZero() {
			pm.ResetsAt = &resetTime
			pm.ResetsIn = formatDuration(resetTime.Sub(now))
		}

		resp.Pools[poolName] = pm
	}

	// Determine overall primary metric for requested window
	if targetWindow == "5h" {
		selectFiveHourMetric(&resp, snap, now)
	} else if len(resp.Pools) > 0 {
		selectFirstPoolMetric(&resp, now)
	} else if primary, ok := snap.Metrics["quota"]; ok {
		resp.Limit = primary.Limit
		resp.Used = primary.Used
		resp.Remaining = primary.Remaining
		resp.Unit = primary.Unit
		resp.Window = "quota"
		if resetTime, ok := snap.Resets["quota"]; ok && !resetTime.IsZero() {
			resp.ResetsAt = &resetTime
			resp.ResetsIn = formatDuration(resetTime.Sub(now))
		}
	}

	return resp
}

func selectFiveHourMetric(resp *GetResponse, snap core.UsageSnapshot, now time.Time) {
	// Look for active model preference or lowest remaining 5h pool
	var (
		minRemaining = math.MaxFloat64
		chosenMetric *core.Metric
		chosenReset  time.Time
	)

	activeModel := strings.ToLower(snap.Attributes["model"])

	for k, m := range snap.Metrics {
		kLower := strings.ToLower(k)
		if m.Window != "5h" && !strings.Contains(kLower, "5h") {
			continue
		}
		if m.Remaining == nil {
			continue
		}

		rem := *m.Remaining
		isModelMatch := activeModel != "" && (strings.Contains(activeModel, "gemini") && strings.Contains(kLower, "gemini") ||
			(strings.Contains(activeModel, "claude") || strings.Contains(activeModel, "sonnet")) && (strings.Contains(kLower, "claude") || strings.Contains(kLower, "3p")))

		if isModelMatch {
			chosenMetric = &m
			chosenReset = findResetTime(snap.Resets, k)
			break
		}

		if rem < minRemaining {
			minRemaining = rem
			chosenMetric = &m
			chosenReset = findResetTime(snap.Resets, k)
		}
	}

	if chosenMetric != nil {
		resp.Limit = chosenMetric.Limit
		resp.Used = chosenMetric.Used
		resp.Remaining = chosenMetric.Remaining
		resp.Unit = chosenMetric.Unit
		resp.Window = "5h"
		if !chosenReset.IsZero() {
			resp.ResetsAt = &chosenReset
			resp.ResetsIn = formatDuration(chosenReset.Sub(now))
		}
	} else if generalQuota, ok := snap.Metrics["quota"]; ok {
		resp.Limit = generalQuota.Limit
		resp.Used = generalQuota.Used
		resp.Remaining = generalQuota.Remaining
		resp.Unit = generalQuota.Unit
		resp.Window = "quota"
		if resetTime, ok := snap.Resets["quota"]; ok && !resetTime.IsZero() {
			resp.ResetsAt = &resetTime
			resp.ResetsIn = formatDuration(resetTime.Sub(now))
		}
	}
}

func selectFirstPoolMetric(resp *GetResponse, now time.Time) {
	poolNames := make([]string, 0, len(resp.Pools))
	for name := range resp.Pools {
		poolNames = append(poolNames, name)
	}
	sort.Strings(poolNames)

	if len(poolNames) > 0 {
		first := resp.Pools[poolNames[0]]
		resp.Limit = first.Limit
		resp.Used = first.Used
		resp.Remaining = first.Remaining
		resp.Unit = first.Unit
		resp.ResetsAt = first.ResetsAt
		resp.ResetsIn = first.ResetsIn
	}
}

func findResetTime(resets map[string]time.Time, metricKey string) time.Time {
	if t, ok := resets[metricKey]; ok && !t.IsZero() {
		return t
	}
	if t, ok := resets[metricKey+"_reset"]; ok && !t.IsZero() {
		return t
	}
	baseKey := strings.TrimPrefix(metricKey, "quota_")
	if t, ok := resets[baseKey]; ok && !t.IsZero() {
		return t
	}
	if t, ok := resets[baseKey+"_reset"]; ok && !t.IsZero() {
		return t
	}
	return time.Time{}
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "now"
	}
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	if hours > 24 {
		days := hours / 24
		remHours := hours % 24
		return fmt.Sprintf("%dd %dh", days, remHours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	if mins > 0 {
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}
