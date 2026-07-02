// Package config centralizes pave-api's environment-variable parsing (see
// .env.example at the pave module root for the documented list), mirroring
// service-template-api/internal/config's Load() pattern.
package config

import (
	"fmt"
	"os"

	"github.com/pavestack/pave/internal/workspace"
)

type Config struct {
	ServiceName  string
	ListenAddr   string
	LogLevel     string
	OTLPEndpoint string

	RepoRoot   string
	DryRun     bool
	CORSOrigin string
}

func Load() (Config, error) {
	repoRoot := os.Getenv("PAVE_API_REPO_ROOT")
	if repoRoot == "" {
		root, err := workspace.Root()
		if err != nil {
			return Config{}, fmt.Errorf("resolve repo root: %w (set PAVE_API_REPO_ROOT explicitly)", err)
		}
		repoRoot = root
	}

	// Default to dry-run: a demo/public deployment of pave-api should never
	// silently open real pull requests against a real repo unless an
	// operator explicitly opts in. Never change this default - see
	// AGENTS.md's "pave-api - safety defaults" section.
	dryRun := true
	if v := os.Getenv("PAVE_API_DRY_RUN"); v == "false" {
		dryRun = false
	}

	return Config{
		ServiceName:  env("PAVE_API_SERVICE_NAME", "pave-api"),
		ListenAddr:   ":" + env("PAVE_API_PORT", "8787"),
		LogLevel:     env("PAVE_API_LOG_LEVEL", "info"),
		OTLPEndpoint: env("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		RepoRoot:     repoRoot,
		DryRun:       dryRun,
		// "*" is rejected once cookie-based auth is in use (credentialed
		// CORS forbids a wildcard origin), so default to the portal's local
		// dev origin rather than allow-all.
		CORSOrigin: env("PAVE_API_CORS_ORIGIN", "http://localhost:5173"),
	}, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
