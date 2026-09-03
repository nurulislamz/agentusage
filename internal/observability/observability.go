package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/nurulislamz/agentusage/internal/version"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	Enabled     bool              `json:"enabled"`
	Endpoint    string            `json:"endpoint,omitempty"`
	Insecure    bool              `json:"insecure,omitempty"`
	ServiceName string            `json:"service_name,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

var (
	mu             sync.RWMutex
	enabled        bool
	logger         *slog.Logger
	tracerProvider *sdktrace.TracerProvider
	loggerProvider *sdklog.LoggerProvider
	tracer         trace.Tracer
)

var (
	sensitiveKVPattern     = regexp.MustCompile(`(?i)(api[_-]?key|token|cookie|secret|auth|password)[\s:=]+([^\s,;&]+)`)
	sensitiveBearerPattern = regexp.MustCompile(`(?i)(bearer)[\s:=]+([^\s,;&]+)`)
	sensitiveAuthPattern   = regexp.MustCompile(`(?i)(authorization)[\s:=]+(?:bearer\s+)?([^\s,;&]+)`)
)


func ResolveConfig(cfg Config) Config {
	if v := os.Getenv("AGENTUSAGE_OTEL_ENABLED"); v != "" {
		cfg.Enabled = v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
	}
	if v := os.Getenv("AGENTUSAGE_OTEL_ENDPOINT"); v != "" {
		cfg.Endpoint = v
	}
	if v := os.Getenv("AGENTUSAGE_OTEL_INSECURE"); v != "" {
		cfg.Insecure = v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
	}
	if v := os.Getenv("AGENTUSAGE_OTEL_SERVICE_NAME"); v != "" {
		cfg.ServiceName = v
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "agentusage"
	}
	return cfg
}

func Init(ctx context.Context, cfg Config) error {
	mu.Lock()
	defer mu.Unlock()

	cfg = ResolveConfig(cfg)
	if !cfg.Enabled || cfg.Endpoint == "" {
		enabled = false
		return nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(version.Version),
			attribute.String("service.commit", version.CommitHash),
			attribute.String("service.build_date", version.BuildDate),
		),
	)
	if err != nil {
		res = resource.Default()
	}

	// Trace Exporter
	traceOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		traceOpts = append(traceOpts, otlptracehttp.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		traceOpts = append(traceOpts, otlptracehttp.WithHeaders(cfg.Headers))
	}
	traceExp, err := otlptracehttp.New(ctx, traceOpts...)
	if err != nil {
		return fmt.Errorf("init otlp trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	tracerProvider = tp
	tracer = tp.Tracer("agentusage")

	// Log Exporter
	logOpts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		logOpts = append(logOpts, otlploghttp.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		logOpts = append(logOpts, otlploghttp.WithHeaders(cfg.Headers))
	}
	logExp, err := otlploghttp.New(ctx, logOpts...)
	if err != nil {
		return fmt.Errorf("init otlp log exporter: %w", err)
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)
	loggerProvider = lp
	logger = otelslog.NewLogger("agentusage", otelslog.WithLoggerProvider(lp))
	enabled = true

	return nil
}

func IsEnabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return enabled
}

func Tracer() trace.Tracer {
	mu.RLock()
	defer mu.RUnlock()
	if tracer == nil {
		return otel.GetTracerProvider().Tracer("agentusage")
	}
	return tracer
}

func EmitLog(ctx context.Context, level slog.Level, component, event, msg string, attrs ...slog.Attr) {
	mu.RLock()
	l := logger
	on := enabled
	mu.RUnlock()

	if !on || l == nil {
		return
	}

	safeMsg := RedactSensitive(msg)
	allAttrs := []slog.Attr{
		slog.String("component", component),
		slog.String("event", event),
	}
	for _, a := range attrs {
		allAttrs = append(allAttrs, slog.Attr{
			Key:   a.Key,
			Value: slog.StringValue(RedactSensitive(a.Value.String())),
		})
	}

	l.LogAttrs(ctx, level, safeMsg, allAttrs...)
}

func EmitInfo(component, event, format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	EmitLog(context.Background(), slog.LevelInfo, component, event, msg)
}

func EmitWarn(component, event, format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	EmitLog(context.Background(), slog.LevelWarn, component, event, msg)
}

func EmitError(component, event, format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	EmitLog(context.Background(), slog.LevelError, component, event, msg)
}

func EmitDebug(component, event, format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	EmitLog(context.Background(), slog.LevelDebug, component, event, msg)
}


func RedactSensitive(s string) string {
	if s == "" {
		return ""
	}
	s = sensitiveAuthPattern.ReplaceAllString(s, `$1=[REDACTED]`)
	s = sensitiveBearerPattern.ReplaceAllString(s, `$1=[REDACTED]`)
	return sensitiveKVPattern.ReplaceAllString(s, `$1=[REDACTED]`)
}


func Flush(ctx context.Context) error {
	mu.RLock()
	tp := tracerProvider
	lp := loggerProvider
	mu.RUnlock()

	if tp != nil {
		_ = tp.ForceFlush(ctx)
	}
	if lp != nil {
		_ = lp.ForceFlush(ctx)
	}
	return nil
}

func ResetForTesting() {
	mu.Lock()
	defer mu.Unlock()
	enabled = false
	logger = nil
	tracerProvider = nil
	loggerProvider = nil
	tracer = nil
}

func Shutdown(ctx context.Context) error {

	mu.Lock()
	defer mu.Unlock()

	enabled = false
	var errs []string
	if tracerProvider != nil {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err.Error())
		}
		tracerProvider = nil
	}
	if loggerProvider != nil {
		if err := loggerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err.Error())
		}
		loggerProvider = nil
	}
	logger = nil
	if len(errs) > 0 {
		return fmt.Errorf("shutdown error: %s", strings.Join(errs, ", "))
	}
	return nil
}
