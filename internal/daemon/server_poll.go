package daemon

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func (s *Service) runPollLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()

	s.infof("poll_loop_start", "interval=%s", s.cfg.PollInterval)
	s.pollProviders(ctx)
	for {
		select {
		case <-ctx.Done():
			s.infof("poll_loop_stop", "reason=context_done")
			return
		case <-ticker.C:
			s.pollProviders(ctx)
		case <-s.pollKick:
			s.infof("poll_kick", "reason=on_demand")
			s.pollProviders(ctx)
		}
	}
}

func (s *Service) pollProviders(ctx context.Context) {
	s.pollProvidersTargeted(ctx, "", false, false)
}

func (s *Service) pollProvidersTargeted(ctx context.Context, targetAccountID string, manual bool, force bool) {
	if s == nil || (s.quotaIngest == nil && s.pollScheduler == nil) {
		return
	}
	if s.pollScheduler == nil {
		s.pollMu.Lock()
		defer s.pollMu.Unlock()
		s.doPollTargeted(ctx, targetAccountID, manual, force)
		return
	}

	genID, isLeader, done := s.pollScheduler.BeginGeneration()
	if !isLeader {
		select {
		case <-done:
			return
		case <-ctx.Done():
			return
		}
	}
	defer s.pollScheduler.EndGeneration(genID)

	execCtx := s.serviceContext(ctx)
	s.doPollTargeted(execCtx, targetAccountID, manual, force)
}

func (s *Service) doPoll(ctx context.Context) {
	s.doPollTargeted(ctx, "", false, false)
}

func (s *Service) doPollTargeted(ctx context.Context, targetAccountID string, manual bool, force bool) {
	s.pollMu.Lock()
	defer s.pollMu.Unlock()
	started := time.Now()

	accounts, modelNorm, err := loadAccountsAndNormFunc()
	if err != nil {
		if s.shouldLog("poll_config_warning", 20*time.Second) {
			s.warnf("poll_config_warning", "error=%v", err)
		}
		return
	}
	if len(accounts) == 0 {
		if s.shouldLog("poll_no_accounts", 30*time.Second) {
			s.infof("poll_skipped", "reason=no_enabled_accounts")
		}
		return
	}

	if targetAccountID != "" {
		filtered := make([]core.AccountConfig, 0, 1)
		for _, acct := range accounts {
			if strings.EqualFold(acct.ID, targetAccountID) {
				filtered = append(filtered, acct)
				break
			}
		}
		if len(filtered) > 0 {
			accounts = filtered
		}
	}

	type providerResult struct {
		accountID string
		snapshot  core.UsageSnapshot
	}

	results := make(chan providerResult, len(accounts))
	var wg sync.WaitGroup

	for _, acct := range accounts {
		wg.Add(1)
		go func(account core.AccountConfig) {
			defer wg.Done()
			if snap := s.pollSingleAccount(ctx, account, modelNorm, manual, force); snap != nil {
				results <- providerResult{accountID: account.ID, snapshot: *snap}
			}
		}(acct)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	snapshots := make(map[string]core.UsageSnapshot, len(accounts))
	statusCounts := map[core.Status]int{}
	errorCount := 0
	for result := range results {
		snapshots[result.accountID] = result.snapshot
		statusCounts[result.snapshot.Status]++
		if result.snapshot.Status == core.StatusError {
			errorCount++
		}
	}
	if len(snapshots) == 0 {
		return
	}

	var ingestErr error
	if s.quotaIngest != nil {
		ingestCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
		defer cancel()
		ingestErr = s.ingestQuotaSnapshots(ingestCtx, snapshots)
		if ingestErr != nil && s.shouldLog("poll_ingest_warning", 10*time.Second) {
			s.warnf("poll_ingest_warning", "error=%v", ingestErr)
		}
	}
	if ingestErr == nil && len(snapshots) > 0 {
		s.markDataIngested()
		s.publishReadModelSync(ctx)
	}

	durationMs := time.Since(started).Milliseconds()
	if ingestErr != nil || errorCount > 0 || s.shouldLog("poll_cycle_info", 45*time.Second) {
		s.infof(
			"poll_cycle",
			"duration_ms=%d accounts=%d snapshots=%d status_ok=%d status_auth=%d status_limited=%d status_error=%d status_unknown=%d ingest_error=%t",
			durationMs,
			len(accounts),
			len(snapshots),
			statusCounts[core.StatusOK],
			statusCounts[core.StatusAuth],
			statusCounts[core.StatusLimited],
			statusCounts[core.StatusError],
			statusCounts[core.StatusUnknown],
			ingestErr != nil,
		)
	}
}

// skipUnchangedProvider checks if a provider's data source has changed since the last
// fetch. Returns the cached snapshot if unchanged, nil if a fresh Fetch() is needed.
func (s *Service) skipUnchangedProvider(provider core.UsageProvider, acct core.AccountConfig) *core.UsageSnapshot {
	detector, ok := provider.(core.ChangeDetector)
	if !ok {
		return nil // provider doesn't support change detection, always fetch
	}

	s.pollStateMu.Lock()
	state := s.pollState[acct.ID]
	s.pollStateMu.Unlock()

	if state == nil || !state.hasSnap {
		return nil // no previous fetch, must run
	}

	now := s.now()
	if snapshotResetPassed(state.lastSnap, state.lastFetchAt, now) {
		core.Tracef("[poll] %s/%s: forcing refresh because a reset boundary passed after %s", acct.Provider, acct.ID, state.lastFetchAt.Format(time.RFC3339))
		return nil
	}

	changed, err := detector.HasChanged(acct, state.lastFetchAt)
	if err != nil || changed {
		return nil // error or changed — run Fetch()
	}

	core.Tracef("[poll] %s/%s: skipped (no change since %s)", acct.Provider, acct.ID, state.lastFetchAt.Format(time.RFC3339))
	snap := state.lastSnap
	return &snap
}

func (s *Service) pollSingleAccount(ctx context.Context, account core.AccountConfig, modelNorm core.ModelNormalizationConfig, manual bool, force bool) *core.UsageSnapshot {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	provider, ok := s.providerByID[account.Provider]
	if !ok {
		return &core.UsageSnapshot{
			ProviderID: account.Provider,
			AccountID:  account.ID,
			Timestamp:  s.now().UTC(),
			Status:     core.StatusError,
			Message:    fmt.Sprintf("no provider adapter registered for %q (restart/reinstall telemetry daemon if recently added)", account.Provider),
		}
	}

	_, hasDetector := provider.(core.ChangeDetector)

	if manual {
		if !s.pollScheduler.ShouldPollManual(account.ID, force) {
			s.pollStateMu.Lock()
			state := s.pollState[account.ID]
			s.pollStateMu.Unlock()
			if state != nil && state.hasSnap {
				return &state.lastSnap
			}
		}
	} else {
		// Adaptive backoff: skip providers that are in a backoff window.
		if !s.pollScheduler.ShouldPoll(account.ID, hasDetector) {
			s.pollStateMu.Lock()
			state := s.pollState[account.ID]
			s.pollStateMu.Unlock()
			if state != nil && state.hasSnap {
				return &state.lastSnap
			}
		}

		// Check if provider data has changed since last fetch (optional interface).
		if cached := s.skipUnchangedProvider(provider, account); cached != nil {
			s.pollScheduler.RecordPoll(account.ID, false)
			return cached
		}
	}

	fetchStart := time.Now()
	fetchCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	snap, fetchErr := provider.Fetch(fetchCtx, account)
	if snap.Timestamp.IsZero() || snap.Timestamp.Before(fetchStart) {
		snap.Timestamp = s.now().UTC()
	}
	fetchDurationMs := time.Since(fetchStart).Milliseconds()
	if fetchErr != nil {
		s.warnf(
			"provider_fetch_error",
			"provider=%s account_id=%s duration_ms=%d error=%v",
			account.Provider, account.ID, fetchDurationMs, fetchErr,
		)
		s.pollStateMu.Lock()
		prevState := s.pollState[account.ID]
		s.pollStateMu.Unlock()
		if prevState != nil && prevState.hasSnap && (prevState.lastSnap.Status == core.StatusOK || prevState.lastSnap.Status == core.StatusLimited) {
			cached := prevState.lastSnap.DeepClone()
			cached.EnsureMaps()
			cached.Status = core.StatusError
			cached.Message = fetchErr.Error()
			cached.Diagnostics["last_error"] = fetchErr.Error()
			cached.Diagnostics["last_error_at"] = s.now().UTC().Format(time.RFC3339)
			snap = cached
		} else {
			snap = core.UsageSnapshot{
				ProviderID: account.Provider,
				AccountID:  account.ID,
				Timestamp:  s.now().UTC(),
				Status:     core.StatusError,
				Message:    fetchErr.Error(),
			}
		}
	} else {
		if s.shouldLog("provider_fetch_"+account.ID, 60*time.Second) {
			s.infof(
				"provider_fetch_success",
				"provider=%s account_id=%s duration_ms=%d status=%s",
				account.Provider, account.ID, fetchDurationMs, snap.Status,
			)
		}
		if snap.Status == core.StatusLimited {
			resetAt := findResetBoundary(snap)
			if !resetAt.IsZero() && resetAt.After(s.now()) {
				maxRateLimit := s.now().Add(2 * time.Hour)
				if resetAt.After(maxRateLimit) {
					resetAt = maxRateLimit
				}
				s.pollScheduler.RecordRateLimit(account.ID, resetAt)
			} else if resetAt.IsZero() {
				// No specific rate limit header provided; back off for 1 minute
				s.pollScheduler.RecordRateLimit(account.ID, s.now().Add(time.Minute))
			}
		} else if snap.Status == core.StatusOK {
			s.pollScheduler.ClearRateLimit(account.ID)
		} else if snap.Status == core.StatusError {
			s.pollStateMu.Lock()
			prevState := s.pollState[account.ID]
			s.pollStateMu.Unlock()
			if prevState != nil && prevState.hasSnap && (prevState.lastSnap.Status == core.StatusOK || prevState.lastSnap.Status == core.StatusLimited) {
				cached := prevState.lastSnap.DeepClone()
				cached.EnsureMaps()
				cached.Status = core.StatusError
				if snap.Message != "" {
					cached.Message = snap.Message
				}
				cached.Diagnostics["last_error"] = snap.Message
				cached.Diagnostics["last_error_at"] = s.now().UTC().Format(time.RFC3339)
				snap = cached
			}
		}
	}
	snap = core.NormalizeUsageSnapshotWithConfig(snap, modelNorm)

	// Track whether data actually changed for adaptive backoff.
	changed := s.pollScheduler.SnapshotChanged(account.ID, snap)
	s.pollScheduler.RecordPoll(account.ID, changed)

	// Record successful fetch or error state for future change detection.
	s.pollStateMu.Lock()
	s.pollState[account.ID] = &providerPollState{
		lastFetchAt: s.now(),
		lastSnap:    snap,
		hasSnap:     true,
	}
	s.pollStateMu.Unlock()

	return &snap
}

func isBillingOrQuotaResetKey(key string) bool {
	lower := strings.ToLower(key)
	for _, term := range []string{"billing", "cycle", "month", "plan", "credit", "quota"} {
		if strings.Contains(lower, term) {
			return true
		}
	}
	return false
}

func findResetBoundary(snap core.UsageSnapshot) time.Time {
	var earliest time.Time
	for key, resetAt := range snap.Resets {
		if resetAt.IsZero() {
			continue
		}
		if isBillingOrQuotaResetKey(key) {
			continue
		}
		if earliest.IsZero() || resetAt.Before(earliest) {
			earliest = resetAt
		}
	}
	return earliest
}

func (s *Service) WithClock(c core.Clock) *Service {
	if s != nil {
		s.clock = c
		if s.pollScheduler != nil {
			s.pollScheduler.SetClock(c)
		}
	}
	return s
}

func snapshotResetPassed(snap core.UsageSnapshot, since, now time.Time) bool {
	if since.IsZero() || len(snap.Resets) == 0 {
		return false
	}
	for _, resetAt := range snap.Resets {
		if resetAt.IsZero() {
			continue
		}
		if resetAt.After(since) && !resetAt.After(now) {
			return true
		}
	}
	return false
}
