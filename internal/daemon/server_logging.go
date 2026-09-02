package daemon

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/nurulislamz/agentusage/internal/observability"
)

func (s *Service) infof(event, format string, args ...any) {
	if s == nil {
		return
	}
	var msg string
	if strings.TrimSpace(format) == "" {
		if s.cfg.Verbose {
			log.Printf("daemon level=info event=%s", event)
		}
	} else {
		msg = fmt.Sprintf(format, args...)
		if s.cfg.Verbose {
			log.Printf("daemon level=info event=%s %s", event, msg)
		}
	}

	if observability.IsEnabled() {
		observability.EmitLog(context.Background(), slog.LevelInfo, "daemon", event, msg)
	}
}

func (s *Service) warnf(event, format string, args ...any) {
	if s == nil {
		return
	}
	var msg string
	if strings.TrimSpace(format) == "" {
		if s.cfg.Verbose {
			log.Printf("daemon level=warn event=%s", event)
		}
	} else {
		msg = fmt.Sprintf(format, args...)
		if s.cfg.Verbose {
			log.Printf("daemon level=warn event=%s %s", event, msg)
		}
	}

	if observability.IsEnabled() {
		observability.EmitLog(context.Background(), slog.LevelWarn, "daemon", event, msg)
	}
}

func (s *Service) shouldLog(key string, interval time.Duration) bool {
	if s == nil {
		return false
	}
	return s.logThrottle.Allow(key, interval, time.Now())
}
