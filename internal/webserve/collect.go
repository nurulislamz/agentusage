package webserve

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/daemon"
	"github.com/nurulislamz/agentusage/internal/export"
	"github.com/nurulislamz/agentusage/internal/tui"
	"github.com/nurulislamz/agentusage/internal/version"
)

type collector struct {
	mu       sync.Mutex
	fetchMu  sync.Mutex
	cached   Envelope
	cachedAt time.Time
	ttl      time.Duration
	source   export.Source
	demo     bool
	meta     collectorMeta
	opts     Options
	now      func() time.Time
	collect  CollectFunc
	rt       *daemon.ViewRuntime
	enrich   func(ctx context.Context, snaps map[string]core.UsageSnapshot, accountID string)

	fetchTimeout  time.Duration
	snapshotFetch func(ctx context.Context, refresh bool, accountID string) ([]core.UsageSnapshot, string, error)
}

type collectorMeta struct {
	version        string
	timeWindow     string
	theme          string
	refreshSeconds int
	catalog        []CatalogEntry
}

func newCollector(opts Options) *collector {
	refresh := opts.RefreshSeconds
	if refresh <= 0 {
		refresh = 30
	}
	src := export.Source(strings.ToLower(strings.TrimSpace(opts.Source)))
	if src == "" {
		src = export.SourceAuto
	}
	now := opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	cfg := configOrDefault(opts)
	if err := tui.LoadThemes(config.ConfigDir()); err != nil && core.DebugEnabled() {
		log.Printf("serve: theme load: %v", err)
	}
	if theme := strings.TrimSpace(opts.Theme); theme != "" {
		tui.SetThemeByName(theme)
	} else if theme := strings.TrimSpace(cfg.Theme); theme != "" {
		tui.SetThemeByName(theme)
	}
	tw := core.ParseTimeWindow(cfg.Data.TimeWindow)
	rt := daemon.NewViewRuntime(nil, daemon.ResolveSocketPath(), core.DebugEnabled())
	rt.SetTimeWindow(tw)

	c := &collector{
		ttl:    time.Duration(refresh) * time.Second,
		source: src,
		demo:   opts.Demo,
		now:    now,
		opts:   opts,
		rt:     rt,
		enrich: nil,
		meta: collectorMeta{
			version:        strings.TrimSpace(opts.Version),
			timeWindow:     strings.TrimSpace(opts.TimeWindow),
			theme:          strings.TrimSpace(opts.Theme),
			refreshSeconds: refresh,
			catalog:        providerCatalog(),
		},
		collect:      opts.Collect,
		fetchTimeout: 12 * time.Second,
	}
	if c.meta.version == "" {
		c.meta.version = strings.TrimSpace(version.Version)
	}
	if c.meta.timeWindow == "" {
		c.meta.timeWindow = "30d"
	}
	if c.meta.theme == "" {
		c.meta.theme = "Gruvbox"
	}
	return c
}

func (c *collector) envelope() (Envelope, error) {
	return c.envelopeRefresh(false, "")
}

func (c *collector) envelopeRefresh(refresh bool, accountID string) (Envelope, error) {
	if c.collect != nil {
		env, err := c.collect()
		if err != nil {
			return Envelope{}, err
		}
		return c.decorate(env), nil
	}

	if env, ok := c.cachedEnvelope(refresh); ok {
		return env, nil
	}

	c.fetchMu.Lock()
	defer c.fetchMu.Unlock()

	if env, ok := c.cachedEnvelope(refresh); ok {
		return env, nil
	}

	env, err := c.fetch(refresh, accountID)
	if err != nil {
		return Envelope{}, err
	}
	env = c.decorate(env)
	c.mu.Lock()
	c.cached = env
	c.cachedAt = c.now()
	c.mu.Unlock()
	return env, nil
}

func (c *collector) cachedEnvelope(refresh bool) (Envelope, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if refresh || c.cachedAt.IsZero() || c.ttl <= 0 {
		return Envelope{}, false
	}
	if c.now().Sub(c.cachedAt) >= c.ttl {
		return Envelope{}, false
	}
	return c.cached, true
}

func (c *collector) fetch(refresh bool, accountID string) (Envelope, error) {
	timeout := c.fetchTimeout
	if timeout <= 0 {
		timeout = 12 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type outcome struct {
		snaps  []core.UsageSnapshot
		source string
		err    error
	}
	ch := make(chan outcome, 1)
	go func() {
		snaps, source, err := c.fetchSnapshots(ctx, refresh, accountID)
		ch <- outcome{snaps: snaps, source: source, err: err}
	}()
	select {
	case out := <-ch:
		if out.err != nil {
			return Envelope{}, out.err
		}
		return Envelope{
			Source:    out.source,
			Snapshots: out.snaps,
		}, nil
	case <-ctx.Done():
		timer := time.NewTimer(150 * time.Millisecond)
		defer timer.Stop()
		select {
		case out := <-ch:
			if out.err != nil {
				return Envelope{}, out.err
			}
			return Envelope{
				Source:    out.source,
				Snapshots: out.snaps,
			}, nil
		case <-timer.C:
			return Envelope{}, fmt.Errorf("serve: collecting snapshots: %w", ctx.Err())
		}
	}
}

func (c *collector) setUsageMode(mode string) {
	mode = normalizeUsageMode(mode)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.opts.UsageMode = mode
	if c.opts.Config != nil {
		cfg := *c.opts.Config
		cfg.Dashboard.UsageMode = mode
		c.opts.Config = &cfg
	}
	if c.collect == nil && !c.cachedAt.IsZero() {
		c.cached = c.decorate(Envelope{Source: c.cached.Source, Snapshots: c.cached.Snapshots})
	}
}

func (c *collector) setTheme(theme string) {
	theme = strings.TrimSpace(theme)
	if theme == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.meta.theme = theme
	c.opts.Theme = theme
	if c.opts.Config != nil {
		cfg := *c.opts.Config
		cfg.Theme = theme
		c.opts.Config = &cfg
	}
	if c.collect == nil && !c.cachedAt.IsZero() {
		c.cached = c.decorate(Envelope{Source: c.cached.Source, Snapshots: c.cached.Snapshots})
	}
}

func normalizeUsageMode(mode string) string {
	if strings.EqualFold(strings.TrimSpace(mode), config.UsageModeUsed) {
		return config.UsageModeUsed
	}
	return config.UsageModeRemaining
}

func (c *collector) decorate(env Envelope) Envelope {
	out := env
	out.SchemaVersion = schemaVersion
	out.GeneratedAt = c.now()
	out.AgentUsageVersion = c.meta.version
	out.TimeWindow = c.meta.timeWindow
	out.Theme = c.meta.theme
	out.RefreshIntervalSeconds = c.meta.refreshSeconds
	out.Catalog = c.meta.catalog
	if c.opts.UsageMode != "" {
		out.UsageMode = c.opts.UsageMode
	} else if c.opts.Config != nil && c.opts.Config.Dashboard.UsageMode != "" {
		out.UsageMode = c.opts.Config.Dashboard.UsageMode
	}
	if strings.TrimSpace(out.Source) == "" {
		if c.demo {
			out.Source = "demo"
		} else {
			out.Source = string(c.source)
		}
	}
	out.Snapshots = stripRaw(out.Snapshots)
	views, tokens := buildViews(c.opts, c.meta, out.Snapshots)
	out.Views = views
	out.ThemeTokens = tokens
	out.TimeWindowLabel = core.ParseTimeWindow(out.TimeWindow).Label()
	if out.UsageMode == "" {
		out.UsageMode = "remaining"
	}
	out.OkCount = 0
	out.WarnCount = 0
	out.ErrCount = 0
	for _, v := range views {
		switch strings.ToUpper(v.Status) {
		case string(core.StatusOK):
			out.OkCount++
		case string(core.StatusNearLimit):
			out.WarnCount++
		case string(core.StatusLimited), string(core.StatusError):
			out.ErrCount++
		}
	}
	out.ProviderCount = len(views)
	out.UnmappedCount, out.UnmappedPhrase = tui.WebUnmappedSummary(out.Snapshots)
	out.AvailableThemes = tui.AvailableThemeNames()
	if len(out.Snapshots) == 0 && !c.demo && strings.TrimSpace(out.Error) == "" {
		out.Error = "No usage snapshots found. No active provider accounts or telemetry records detected."
	}
	return out
}
