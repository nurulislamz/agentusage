package core

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/nurulislamz/agentusage/internal/observability"
)

var (
	traceEnabled     bool
	traceEnabledOnce sync.Once
)

func isTraceEnabled() bool {
	traceEnabledOnce.Do(func() {
		traceEnabled = os.Getenv("AGENTUSAGE_DEBUG") != ""
	})
	return traceEnabled
}

// DebugEnabled reports whether AGENTUSAGE_DEBUG is enabled.
func DebugEnabled() bool {
	return isTraceEnabled()
}

// Tracef logs a formatted message to stderr when AGENTUSAGE_DEBUG is set,
// and emits to observability when enabled.
func Tracef(format string, args ...any) {
	if DebugEnabled() {
		log.Printf("[trace] "+format, args...)
	}
	if observability.IsEnabled() {
		msg := format
		if len(args) > 0 {
			msg = fmt.Sprintf(format, args...)
		}
		observability.EmitLog(context.Background(), slog.LevelDebug, "core", "trace", msg)
	}
}

