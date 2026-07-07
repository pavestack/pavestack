package config_test

import (
	"testing"

	"github.com/pavestack/pave/internal/config"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"PAVE_API_REPO_ROOT",
		"PAVE_API_PORT",
		"PAVE_API_DRY_RUN",
		"PAVE_API_CORS_ORIGIN",
		"PAVE_API_SERVICE_NAME",
		"PAVE_API_LOG_LEVEL",
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"PAVE_API_DISABLE_AUTH",
		"PAVE_API_BASE_URL",
		"PAVE_API_PORTAL_URL",
		"PAVE_API_SESSION_SECRET",
		"PAVE_API_GITHUB_CLIENT_ID",
		"PAVE_API_GITHUB_CLIENT_SECRET",
		"PAVE_API_GITHUB_ORG",
		"PAVE_API_APPROVER_TEAM",
	} {
		t.Setenv(key, "")
	}
}

func TestLoadDefaultConfig(t *testing.T) {
	clearEnv(t)
	t.Setenv("PAVE_API_REPO_ROOT", t.TempDir())
	t.Setenv("PAVE_API_DISABLE_AUTH", "true")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.ServiceName != "pave-api" {
		t.Errorf("expected ServiceName 'pave-api', got %q", cfg.ServiceName)
	}
	if cfg.ListenAddr != ":8787" {
		t.Errorf("expected ListenAddr ':8787', got %q", cfg.ListenAddr)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected LogLevel 'info', got %q", cfg.LogLevel)
	}
	if cfg.OTLPEndpoint != "" {
		t.Errorf("expected OTLPEndpoint to be empty, got %q", cfg.OTLPEndpoint)
	}
	if !cfg.DryRun {
		t.Error("expected DryRun to default to true - never change this default")
	}
	if cfg.CORSOrigin == "*" {
		t.Error("expected CORSOrigin to never default to a wildcard")
	}
	if cfg.PortalURL == "" {
		t.Error("expected a default PortalURL")
	}
	if cfg.CookieSecure {
		t.Error("expected CookieSecure to be false for the default http://localhost BaseURL")
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	clearEnv(t)
	t.Setenv("PAVE_API_REPO_ROOT", t.TempDir())
	t.Setenv("PAVE_API_PORT", "9090")
	t.Setenv("PAVE_API_DRY_RUN", "false")
	t.Setenv("PAVE_API_CORS_ORIGIN", "https://portal.pavestack.io")
	t.Setenv("PAVE_API_SERVICE_NAME", "pave-api-custom")
	t.Setenv("PAVE_API_LOG_LEVEL", "debug")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4318")
	t.Setenv("PAVE_API_BASE_URL", "https://pave-api.pavestack.io")
	t.Setenv("PAVE_API_SESSION_SECRET", "s3cr3t")
	t.Setenv("PAVE_API_GITHUB_CLIENT_ID", "client-id")
	t.Setenv("PAVE_API_GITHUB_CLIENT_SECRET", "client-secret")
	t.Setenv("PAVE_API_GITHUB_ORG", "acme")
	t.Setenv("PAVE_API_APPROVER_TEAM", "sre")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.ListenAddr != ":9090" {
		t.Errorf("expected ListenAddr ':9090', got %q", cfg.ListenAddr)
	}
	if cfg.DryRun {
		t.Error("expected DryRun to be false when PAVE_API_DRY_RUN=false")
	}
	if cfg.CORSOrigin != "https://portal.pavestack.io" {
		t.Errorf("expected CORSOrigin override, got %q", cfg.CORSOrigin)
	}
	if cfg.ServiceName != "pave-api-custom" {
		t.Errorf("expected ServiceName override, got %q", cfg.ServiceName)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug', got %q", cfg.LogLevel)
	}
	if cfg.OTLPEndpoint != "http://otel-collector:4318" {
		t.Errorf("expected OTLPEndpoint override, got %q", cfg.OTLPEndpoint)
	}
	if !cfg.CookieSecure {
		t.Error("expected CookieSecure to be true for an https:// BaseURL")
	}
	if cfg.GitHubOrg != "acme" || cfg.ApproverTeam != "sre" {
		t.Errorf("expected GitHub org/team overrides, got org=%q team=%q", cfg.GitHubOrg, cfg.ApproverTeam)
	}
	if string(cfg.SessionSecret) != "s3cr3t" {
		t.Errorf("expected SessionSecret to be set, got %q", cfg.SessionSecret)
	}
}

func TestLoadFailsClosedWithoutOAuthConfigOrDisableAuth(t *testing.T) {
	clearEnv(t)
	t.Setenv("PAVE_API_REPO_ROOT", t.TempDir())

	if _, err := config.Load(); err == nil {
		t.Fatal("expected Load to fail closed when OAuth isn't configured and auth isn't explicitly disabled")
	}
}

func TestLoadMissingRepoRootWithoutMarkerFails(t *testing.T) {
	clearEnv(t)
	t.Setenv("PAVESTACK_ROOT", "")
	t.Setenv("PAVE_API_DISABLE_AUTH", "true")
	t.Chdir(t.TempDir())

	if _, err := config.Load(); err == nil {
		t.Fatal("expected error resolving repo root outside a pavestack checkout")
	}
}
