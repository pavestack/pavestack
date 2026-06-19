package config_test

import (
	"testing"

	"github.com/pavestack/service-template-api/internal/config"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Clear environment variables that might be set in the test runner
	t.Setenv("SERVICE_NAME", "")
	t.Setenv("LISTEN_ADDR", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("READY", "")

	cfg := config.Load()

	if cfg.ServiceName != "service-template-api" {
		t.Errorf("expected ServiceName 'service-template-api', got %q", cfg.ServiceName)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected ListenAddr ':8080', got %q", cfg.ListenAddr)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected LogLevel 'info', got %q", cfg.LogLevel)
	}
	if cfg.OTLPEndpoint != "" {
		t.Errorf("expected OTLPEndpoint to be empty, got %q", cfg.OTLPEndpoint)
	}
	if !cfg.Ready {
		t.Error("expected Ready to be true by default")
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("SERVICE_NAME", "custom-service")
	t.Setenv("LISTEN_ADDR", ":9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4318")
	t.Setenv("READY", "false")

	cfg := config.Load()

	if cfg.ServiceName != "custom-service" {
		t.Errorf("expected ServiceName 'custom-service', got %q", cfg.ServiceName)
	}
	if cfg.ListenAddr != ":9090" {
		t.Errorf("expected ListenAddr ':9090', got %q", cfg.ListenAddr)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug', got %q", cfg.LogLevel)
	}
	if cfg.OTLPEndpoint != "http://otel-collector:4318" {
		t.Errorf("expected OTLPEndpoint 'http://otel-collector:4318', got %q", cfg.OTLPEndpoint)
	}
	if cfg.Ready {
		t.Error("expected Ready to be false")
	}
}

func TestLoadConfigWithInvalidReadyEnv(t *testing.T) {
	t.Setenv("READY", "invalid-boolean")
	cfg := config.Load()

	// Should fallback to default (true)
	if !cfg.Ready {
		t.Error("expected Ready to fallback to true on invalid boolean")
	}
}
