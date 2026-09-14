package webserve

import (
	"sort"
	"strings"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/tui"
)

// ProviderPanelRow is one account row in the providers dropdown.
type ProviderPanelRow struct {
	AccountID   string
	ProviderID  string
	Status      string
	StatusBadge string
	Enabled     bool
}

// ProviderPanelGroup groups dropdown rows under their provider.
type ProviderPanelGroup struct {
	ProviderID   string
	ProviderName string
	Rows         []ProviderPanelRow
}

// providerPanelData is the template payload for the providers dropdown.
type providerPanelData struct {
	Groups []ProviderPanelGroup
}

// providerPanelGroups lists every configured and auto-detected account —
// including accounts hidden from the dashboard — so the dropdown can show,
// hide, or remove boxes.
func (s *Server) providerPanelGroups(env Envelope) []ProviderPanelGroup {
	cfg := s.panelConfig()

	enabled := make(map[string]bool, len(cfg.Dashboard.Providers))
	for _, p := range cfg.Dashboard.Providers {
		if id := strings.TrimSpace(p.AccountID); id != "" {
			enabled[id] = p.Enabled
		}
	}

	status := make(map[string]ProviderPanelRow, len(env.Snapshots))
	snapByID := make(map[string]core.UsageSnapshot, len(env.Snapshots))
	for _, snap := range env.Snapshots {
		id := strings.TrimSpace(snap.AccountID)
		if id == "" {
			continue
		}
		snapByID[id] = snap
		status[id] = ProviderPanelRow{
			Status:      string(core.EffectiveStatus(snap)),
			StatusBadge: tui.StripANSI(tui.SnapshotStatusBadge(snap)),
		}
	}

	names := providerNameLookup()
	providerByID := make(map[string]string)
	order := make([]string, 0, 16)
	addID := func(id, pid string) {
		if id == "" {
			return
		}
		var snap core.UsageSnapshot
		if s, ok := snapByID[id]; ok {
			snap = s
			if pid == "" {
				pid = snap.ProviderID
			}
		}
		if isInternalTelemetryAccount(id, pid, snap) {
			return
		}
		if _, seen := providerByID[id]; seen {
			return
		}
		providerByID[id] = pid
		order = append(order, id)
	}
	for _, acct := range core.MergeAccounts(cfg.Accounts, cfg.AutoDetectedAccounts) {
		addID(strings.TrimSpace(acct.ID), strings.TrimSpace(acct.Provider))
	}
	for _, p := range cfg.Dashboard.Providers {
		addID(strings.TrimSpace(p.AccountID), "")
	}
	for _, snap := range env.Snapshots {
		addID(strings.TrimSpace(snap.AccountID), strings.TrimSpace(snap.ProviderID))
	}

	groups := make(map[string]*ProviderPanelGroup)
	for _, id := range order {
		pid := providerByID[id]
		if pid == "" {
			pid = "other"
		}
		g, ok := groups[pid]
		if !ok {
			name := names[pid]
			if name == "" {
				name = pid
			}
			g = &ProviderPanelGroup{ProviderID: pid, ProviderName: name}
			groups[pid] = g
		}
		row := status[id]
		row.AccountID = id
		row.ProviderID = pid
		if visible, listed := enabled[id]; listed {
			row.Enabled = visible
		} else {
			row.Enabled = true
		}
		g.Rows = append(g.Rows, row)
	}

	out := make([]ProviderPanelGroup, 0, len(groups))
	for _, g := range groups {
		if len(g.Rows) == 0 {
			continue
		}
		sort.Slice(g.Rows, func(i, j int) bool {
			if g.Rows[i].Enabled != g.Rows[j].Enabled {
				return g.Rows[i].Enabled
			}
			return g.Rows[i].AccountID < g.Rows[j].AccountID
		})
		out = append(out, *g)
	}
	groupShownRatio := func(g ProviderPanelGroup) float64 {
		if len(g.Rows) == 0 {
			return 0
		}
		shown := 0
		for _, r := range g.Rows {
			if r.Enabled {
				shown++
			}
		}
		return float64(shown) / float64(len(g.Rows))
	}
	sort.Slice(out, func(i, j int) bool {
		rI := groupShownRatio(out[i])
		rJ := groupShownRatio(out[j])
		if rI != rJ {
			return rI > rJ
		}
		if out[i].ProviderName != out[j].ProviderName {
			return out[i].ProviderName < out[j].ProviderName
		}
		return out[i].ProviderID < out[j].ProviderID
	})
	return out
}

func providerNameLookup() map[string]string {
	names := make(map[string]string)
	for _, c := range providerCatalog() {
		names[c.ID] = c.Name
	}
	return names
}

// panelConfig resolves the config backing the providers dropdown. The
// collector's in-memory copy wins: it is the same startup config the dashboard
// filters with, and hiding/showing must stay consistent with what renders.
func (s *Server) panelConfig() config.Config {
	if s.collector != nil {
		if s.collector.opts.Config != nil {
			return *s.collector.opts.Config
		}
		return configOrDefault(s.collector.opts)
	}
	if cfg, err := config.Load(); err == nil {
		return cfg
	}
	return config.DefaultConfig()
}

// setAccountVisibility hides or shows one account on the dashboard and
// persists the toggle in settings.json (dashboard.providers[].enabled).
func (s *Server) setAccountVisibility(accountID string, visible bool) error {
	cfg := s.panelConfig()
	providers := append([]config.DashboardProviderConfig(nil), cfg.Dashboard.Providers...)
	found := false
	for i := range providers {
		if providers[i].AccountID == accountID {
			providers[i].Enabled = visible
			found = true
			break
		}
	}
	if !found {
		providers = append(providers, config.DashboardProviderConfig{AccountID: accountID, Enabled: visible})
	}
	if !s.persistDisabled() {
		if err := config.SaveDashboardProviders(providers); err != nil {
			return err
		}
	}
	cfg.Dashboard.Providers = providers
	s.applyConfig(cfg)
	return nil
}

// deleteAccount removes an account from settings and credentials. A disabled
// dashboard entry is kept so re-detected boxes stay hidden until restored.
func (s *Server) deleteAccount(accountID string) error {
	cfg := s.panelConfig()
	cfg.Accounts = dropAccount(cfg.Accounts, accountID)
	cfg.AutoDetectedAccounts = dropAccount(cfg.AutoDetectedAccounts, accountID)

	providers := make([]config.DashboardProviderConfig, 0, len(cfg.Dashboard.Providers)+1)
	for _, p := range cfg.Dashboard.Providers {
		if p.AccountID == accountID {
			continue
		}
		providers = append(providers, p)
	}
	providers = append(providers, config.DashboardProviderConfig{AccountID: accountID, Enabled: false})
	cfg.Dashboard.Providers = providers

	if !s.persistDisabled() {
		if err := config.Save(cfg); err != nil {
			return err
		}
		_ = config.DeleteCredential(accountID)
	}
	s.applyConfig(cfg)
	return nil
}

func dropAccount(accounts []core.AccountConfig, accountID string) []core.AccountConfig {
	out := make([]core.AccountConfig, 0, len(accounts))
	for _, a := range accounts {
		if a.ID == accountID {
			continue
		}
		out = append(out, a)
	}
	return out
}

// persistDisabled reports whether writes must stay in memory (demo servers and
// test harnesses that inject Options.Config).
func (s *Server) persistDisabled() bool {
	return s.collector == nil || s.collector.demo
}

func (s *Server) applyConfig(cfg config.Config) {
	if s.collector != nil {
		s.collector.setConfig(cfg)
	}
}

func isInternalTelemetryAccount(accountID, providerID string, snap core.UsageSnapshot) bool {
	id := strings.ToLower(strings.TrimSpace(accountID))
	pid := strings.ToLower(strings.TrimSpace(providerID))
	if id == "" {
		return true
	}
	all := id + " " + pid
	telemetryKeys := []string{
		"telemetry root", "telemetry scope", "telemetry_root", "telemetry_scope",
		"link hint", "status file", "installation id", "session id",
		"ineligible reasons", "oauth scope", "sessions dirs", "mapped_target_missing",
		"limit_snapshot", "telemetry-only", "internal-telemetry",
	}
	for _, k := range telemetryKeys {
		if strings.Contains(all, k) {
			return true
		}
	}
	if pid == "telemetry" || id == "telemetry" || id == "telemetry-daemon" || id == "daemon" {
		return true
	}
	for k, v := range snap.Attributes {
		kLow := strings.ToLower(strings.TrimSpace(k))
		vLow := strings.ToLower(strings.TrimSpace(v))
		if strings.Contains(kLow, "telemetry_scope") && strings.Contains(vLow, "internal") {
			return true
		}
	}
	for k, v := range snap.Raw {
		kLow := strings.ToLower(strings.TrimSpace(k))
		vLow := strings.ToLower(strings.TrimSpace(v))
		if strings.Contains(kLow, "telemetry_scope") && strings.Contains(vLow, "internal") {
			return true
		}
	}
	return false
}
