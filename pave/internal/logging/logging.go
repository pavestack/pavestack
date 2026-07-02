// Package logging provides pave-api's structured (zap) logger, mirroring
// service-template-api/internal/logging so the two Go services in this
// monorepo produce the same shape of log line and the same trace_id/span_id
// correlation - see the "Observability - three pillars, one trace ID"
// section of AGENTS.md.
package logging

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(level string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(parseLevel(level))
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.MessageKey = "message"
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func String(key, value string) zap.Field {
	return zap.String(key, value)
}

func Error(err error) zap.Field {
	return zap.Error(err)
}

// TraceContext returns trace_id/span_id fields for the span active in ctx,
// or nil if ctx carries no valid span.
func TraceContext(ctx context.Context) []zap.Field {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return nil
	}
	return []zap.Field{
		zap.String("trace_id", sc.TraceID().String()),
		zap.String("span_id", sc.SpanID().String()),
	}
}

// FromContext returns logger with trace_id/span_id fields attached when
// ctx carries a valid span, otherwise it returns logger unchanged.
func FromContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	if fields := TraceContext(ctx); fields != nil {
		return logger.With(fields...)
	}
	return logger
}
