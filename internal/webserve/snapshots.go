package webserve

import (
	"context"
	"fmt"

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
	projector := tui.NewWebProjectorFromConfig(cfg)

	if c.source != export.SourceDirect && c.rt != nil {
		var frame daemon.SnapshotFrame
		if refresh && accountID == "" {
			frame = c.rt.RefreshForWindow(ctx, tw)
		} else {
			frame = c.rt.ReadWithFallbackForWindow(ctx, tw)
		}
		if len(frame.Snapshots) > 0 {
			if c.enrich != nil {
				c.enrich(ctx, frame.Snapshots, accountID)
			}
			return projector.OrderSnapshots(frame.Snapshots), "daemon", nil
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
	return projector.OrderSnapshots(snapMap), string(resolved), nil
}
