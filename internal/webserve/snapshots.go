package webserve

import (
	"context"
	"fmt"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/daemon"
	"github.com/nurulislamz/agentusage/internal/export"
	"github.com/nurulislamz/agentusage/internal/tui"
)

func (c *collector) fetchSnapshots(ctx context.Context, refresh bool, accountID string) ([]core.UsageSnapshot, string, error) {
	if c.snapshotFetch != nil {
		return c.snapshotFetch(ctx, refresh, accountID)
	}
	if c.demo {
		return demoSnapshots(c.now()), "demo", nil
	}

	cfg := configOrDefault(c.opts)
	tw := core.ParseTimeWindow(cfg.Data.TimeWindow)

	if c.source != export.SourceDirect && c.rt != nil {
		var frame daemon.SnapshotFrame
		if refresh {
			if accountID != "" {
				frame = c.rt.RefreshAccountForWindow(ctx, accountID, tw)
			} else {
				frame = c.rt.RefreshForWindow(ctx, tw)
			}
		} else {
			frame = c.rt.ReadWithFallbackForWindow(ctx, tw)
		}
		if len(frame.Snapshots) > 0 && daemon.SnapshotsHaveUsableData(frame.Snapshots) {
			if c.enrich != nil {
				c.enrich(ctx, frame.Snapshots, accountID)
			}
			return snapshotsForCache(cfg, frame.Snapshots), "daemon", nil
		}
	}

	snaps, resolved, err := export.Collect(ctx, c.source)
	if err != nil {
		return nil, "", fmt.Errorf("serve: collecting snapshots: %w", err)
	}
	snapMap := tui.SnapshotsToMap(snaps)
	if c.enrich != nil {
		c.enrich(ctx, snapMap, accountID)
	}
	return snapshotsForCache(cfg, snapMap), string(resolved), nil
}

// snapshotsForCache keeps every fetched account in the collector cache,
// including dashboard-hidden ones. Visibility (Enabled=false) is applied later
// in decorate/buildViews. Filtering here would drop hidden accounts from the
// cache, so Show could not restore them without a refetch and the providers
// panel would lose status for hidden rows.
func snapshotsForCache(cfg config.Config, snaps map[string]core.UsageSnapshot) []core.UsageSnapshot {
	cfg.Dashboard.Providers = nil
	return tui.NewWebProjectorFromConfig(cfg).OrderSnapshots(snaps)
}
