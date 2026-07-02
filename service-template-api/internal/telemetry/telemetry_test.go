package telemetry_test

import (
	"context"
	"testing"

	"github.com/pavestack/service-template-api/internal/telemetry"
)

func TestInitTelemetryWithoutEndpoint(t *testing.T) {
	ctx := context.Background()
	shutdown, err := telemetry.Init(ctx, "test-service", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected non-nil shutdown function")
	}

	// Call shutdown and ensure it doesn't fail
	if err := shutdown(ctx); err != nil {
		t.Errorf("expected no error on shutdown, got %v", err)
	}
}

func TestInitTelemetryWithEndpoint(t *testing.T) {
	ctx := context.Background()
	// Test initialization with endpoint, since it's just setting up the HTTP exporter client setup
	// We can pass a dummy localhost HTTP endpoint.
	shutdown, err := telemetry.Init(ctx, "test-service", "http://localhost:4318")
	if err != nil {
		t.Fatalf("expected no error initializing with endpoint URL format, got %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected non-nil shutdown function")
	}

	// Init itself never dials the endpoint - only the trace/metric
	// exporters' shutdown-time flush does, and there is no collector
	// listening on this dummy localhost port in a unit test. That flush
	// failure is expected here (and is the reason app.go treats a
	// shutdownTelemetry error as log-and-continue, not fatal); the
	// assertion that matters is that shutdown returns rather than hangs.
	_ = shutdown(ctx)
}
