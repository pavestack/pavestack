// Package config centralizes pave-api's environment-variable parsing (see
// .env.example at the pave module root for the documented list), mirroring
// service-template-api/internal/config's Load() pattern.
package config

import (
	"fmt"
	"os"
	"strings"

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

	// DisableAuth runs pave-api with no authentication on mutating
	// endpoints at all - only ever appropriate for local dev or a CI job
	// that never leaves localhost. It is never the default; see Load.
	DisableAuth        bool
	BaseURL            string // pave-api's own externally reachable origin
	PortalURL          string // where the browser is sent after login/logout
	CookieSecure       bool   // derived from BaseURL's scheme
	SessionSecret      []byte
	GitHubClientID     string
	GitHubClientSecret string
	GitHubOrg          string
	ApproverTeam       string
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

	disableAuth := os.Getenv("PAVE_API_DISABLE_AUTH") == "true"
	sessionSecret := os.Getenv("PAVE_API_SESSION_SECRET")
	clientID := os.Getenv("PAVE_API_GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("PAVE_API_GITHUB_CLIENT_SECRET")

	// Fail closed: an unconfigured OAuth app must not silently mean "no
	// auth" - that's exactly the kind of quiet security regression
	// PAVE_API_DRY_RUN's explicit-opt-in default already guards against
	// for the GitOps side. PAVE_API_DISABLE_AUTH is the equivalent
	// explicit opt-out here, for local dev/CI only.
	if !disableAuth && (sessionSecret == "" || clientID == "" || clientSecret == "") {
		return Config{}, fmt.Errorf(
			"GitHub OAuth is not configured - set PAVE_API_SESSION_SECRET, " +
				"PAVE_API_GITHUB_CLIENT_ID, and PAVE_API_GITHUB_CLIENT_SECRET, " +
				"or set PAVE_API_DISABLE_AUTH=true to run without authentication " +
				"(local dev/CI only - never for a network-reachable deployment)")
	}

	baseURL := env("PAVE_API_BASE_URL", "http://localhost:8787")

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

		DisableAuth:        disableAuth,
		BaseURL:            baseURL,
		PortalURL:          env("PAVE_API_PORTAL_URL", "http://localhost:5173"),
		CookieSecure:       strings.HasPrefix(baseURL, "https://"),
		SessionSecret:      []byte(sessionSecret),
		GitHubClientID:     clientID,
		GitHubClientSecret: clientSecret,
		GitHubOrg:          env("PAVE_API_GITHUB_ORG", "pavestack"),
		ApproverTeam:       env("PAVE_API_APPROVER_TEAM", "platform"),
	}, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
