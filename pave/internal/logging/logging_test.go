package logging_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pavestack/pave/internal/logging"
	"go.opentelemetry.io/otel/trace"
)

func TestNewLogger(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "unknown"}

	for _, lvl := range levels {
		t.Run(lvl, func(t *testing.T) {
			logger := logging.New(lvl)
			if logger == nil {
				t.Fatal("expected non-nil logger")
			}
			logger.Info("test message", logging.String("key", "value"), logging.Error(errors.New("test error")))
		})
	}
}

func TestTraceContextWithoutSpan(t *testing.T) {
	if fields := logging.TraceContext(context.Background()); fields != nil {
		t.Errorf("expected nil fields for a context with no span, got %v", fields)
	}
}

func TestTraceContextWithSpan(t *testing.T) {
	traceID, err := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	if err != nil {
		t.Fatal(err)
	}
	spanID, err := trace.SpanIDFromHex("0102030405060708")
	if err != nil {
		t.Fatal(err)
	}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	fields := logging.TraceContext(ctx)
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields (trace_id, span_id), got %d: %v", len(fields), fields)
	}
}

func TestFromContextAttachesTraceFields(t *testing.T) {
	base := logging.New("error")

	if got := logging.FromContext(context.Background(), base); got == nil {
		t.Fatal("expected non-nil logger when context has no span")
	}

	traceID, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := trace.SpanIDFromHex("0102030405060708")
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	got := logging.FromContext(ctx, base)
	if got == nil {
		t.Fatal("expected non-nil logger when context has a span")
	}
	got.Info("request handled")
}
