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

	if err := shutdown(ctx); err != nil {
		t.Errorf("expected no error on shutdown, got %v", err)
	}
}
