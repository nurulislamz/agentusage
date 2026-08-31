package webserve

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/daemon"
	"github.com/nurulislamz/agentusage/internal/export"
	"github.com/nurulislamz/agentusage/internal/providers/antigravity"
	"github.com/nurulislamz/agentusage/internal/providers/cursor"
	"github.com/nurulislamz/agentusage/internal/providers/opencode"
	"github.com/nurulislamz/agentusage/internal/tui"
	"github.com/nurulislamz/agentusage/internal/version"
)

type collector struct {
	mu       sync.Mutex
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
	tw := core.ParseTimeWindow(cfg.Data.TimeWindow)
	rt := daemon.NewViewRuntime(nil, daemon.ResolveSocketPath(), core.DebugEnabled())
	rt.SetTimeWindow(tw)

	cachedAccounts := core.MergeAccounts(cfg.Accounts, cfg.AutoDetectedAccounts)
	cursorProv := cursor.New()
	antigravityProv := antigravity.New()
	opencodeProv := opencode.New()

	enrich := func(ctx context.Context, snaps map[string]core.UsageSnapshot, accountID string) {
		if len(snaps) == 0 {
			return
		}
		targetSnaps := snaps
		accountID = strings.TrimSpace(accountID)
		if accountID != "" {
			targetSnaps = make(map[string]core.UsageSnapshot)
			if snap, ok := snaps[accountID]; ok {
				targetSnaps[accountID] = snap
			} else {
				return
			}
		}
		enrichCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
		defer cancel()
		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			cursorProv.EnrichSnapshots(enrichCtx, cachedAccounts, targetSnaps)
		}()
		go func() {
			defer wg.Done()
			antigravityProv.EnrichSnapshots(enrichCtx, cachedAccounts, targetSnaps)
		}()
		go func() {
			defer wg.Done()
			opencodeProv.EnrichSnapshots(enrichCtx, cachedAccounts, targetSnaps)
		}()
		wg.Wait()
		if accountID != "" {
			if snap, ok := targetSnaps[accountID]; ok {
				snaps[accountID] = snap
			}
		}
	}

	c := &collector{
		ttl:    time.Duration(refresh) * time.Second,
		source: src,
		demo:   opts.Demo,
		now:    now,
		opts:   opts,
		rt:     rt,
		enrich: enrich,
		meta: collectorMeta{
			version:        strings.TrimSpace(opts.Version),
			timeWindow:     strings.TrimSpace(opts.TimeWindow),
			theme:          strings.TrimSpace(opts.Theme),
			refreshSeconds: refresh,
			catalog:        providerCatalog(),
		},
		collect: opts.Collect,
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

	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	if !refresh && !c.cachedAt.IsZero() && c.ttl > 0 && now.Sub(c.cachedAt) < c.ttl {
		return c.cached, nil
	}

	env, err := c.fetch(refresh, accountID)
	if err != nil {
		return Envelope{}, err
	}
	env = c.decorate(env)
	c.cached = env
	c.cachedAt = now
	return env, nil
}

func (c *collector) fetch(refresh bool, accountID string) (Envelope, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	snaps, source, err := c.fetchSnapshots(ctx, refresh, accountID)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{
		Source:    source,
		Snapshots: snaps,
	}, nil
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
	return out
}
