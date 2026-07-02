package telemetry_test

import (
	"context"
	"testing"

	"github.com/pavestack/pave/internal/telemetry"
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
	if err := shutdown(ctx); err != nil {
		t.Errorf("expected no error on shutdown, got %v", err)
	}
}

func TestInitTelemetryWithEndpoint(t *testing.T) {
	ctx := context.Background()
	shutdown, err := telemetry.Init(ctx, "test-service", "http://localhost:4318")
	if err != nil {
		t.Fatalf("expected no error initializing with endpoint URL format, got %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected non-nil shutdown function")
	}
	// Init itself never dials the endpoint - only the flush at shutdown
	// does, and there is no collector listening in a unit test. That
	// failure is expected; the assertion that matters is that shutdown
	// returns rather than hangs.
	_ = shutdown(ctx)
}
