package daemon

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/telemetry"
)

// runWatchLoop watches local provider data directories for changes and triggers
// immediate collection when files are modified. This replaces fixed-interval
// polling with event-driven collection for local providers (claude_code, cursor,
// codex, gemini_cli, copilot, ollama, antigravity).
//
// Only top-level directories are watched (not individual files) to stay well
// within macOS kqueue descriptor limits.
func (s *Service) runWatchLoop(ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		s.warnf("watch_loop_start_error", "error=%v", err)
		return
	}
	defer watcher.Close()

	dirs := collectWatchDirs()
	watchCount := 0
	for _, dir := range dirs {
		if err := watcher.Add(dir); err == nil {
			watchCount++
		}
	}
	if watchCount == 0 {
		s.infof("watch_loop_skip", "no watchable directories found")
		return
	}
	s.infof("watch_loop_start", "directories=%d", watchCount)

	// Debounce non-antigravity changes into a single read-model refresh.
	var debounceTimer *time.Timer
	debounceInterval := 2 * time.Second

	for {
		select {
		case <-ctx.Done():
			s.infof("watch_loop_stop", "reason=context_done")
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}

			// Antigravity status-line writes need a full Fetch poll so
			// limit_snapshot gauges update. RequestPoll already coalesces.
			if isAntigravityStatusFile(event.Name) {
				s.RequestPoll()
				core.Tracef("[watch] antigravity status change: %s op=%s → poll kick", event.Name, event.Op)
				continue
			}

			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			eventName, eventOp := event.Name, event.Op
			debounceTimer = time.AfterFunc(debounceInterval, func() {
				s.markDataIngested()
				core.Tracef("[watch] change detected: %s op=%s", eventName, eventOp)
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			if s.shouldLog("watch_error", 30*time.Second) {
				s.warnf("watch_error", "error=%v", err)
			}
		}
	}
}

// isAntigravityStatusFile reports whether path is an OpenUsage Antigravity
// status-line state file (default or per-account variants).
func isAntigravityStatusFile(path string) bool {
	base := strings.ToLower(filepath.Base(strings.TrimSpace(path)))
	if !strings.HasSuffix(base, ".json") {
		return false
	}
	if strings.HasPrefix(base, ".") {
		return false // ignore atomic write temps like .antigravity-status-*.tmp
	}
	return strings.HasPrefix(base, "antigravity") && strings.Contains(base, "status")
}

// collectWatchDirs returns the set of directories to watch for changes.
// These are the top-level data directories for each local provider.
func collectWatchDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil
	}

	candidates := []string{
		filepath.Join(home, ".claude", "projects"),
		filepath.Join(home, ".config", "claude", "projects"),
		filepath.Join(home, ".codex", "sessions"),
		filepath.Join(home, ".copilot"),
		filepath.Join(home, ".gemini"),
	}

	if stateDir, err := telemetry.DefaultStateDir(); err == nil && strings.TrimSpace(stateDir) != "" {
		candidates = append(candidates, stateDir)
	}

	// Add platform-specific paths.
	if configDir, _ := os.UserConfigDir(); configDir != "" {
		candidates = append(candidates,
			filepath.Join(configDir, "Cursor", "User", "globalStorage"),
		)
	}

	var dirs []string
	for _, dir := range candidates {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}
