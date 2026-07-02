// Command pave-api is the HTTP backend for the Pavestack self-service
// portal's "Create service" and "Request access" flows. It is a thin HTTP
// shell around the same scaffold/gitops/validate packages `pave
// create-service` uses directly - see pave/internal/apiserver.
//
// Running this against your own checkout of the Pavestack monorepo will
// really scaffold services and (unless PAVE_API_DRY_RUN=true, which is the
// default) really open pull requests via the `gh` CLI, exactly like the
// CLI does. Point PAVESTACK_ROOT at a scratch clone if you want to try it
// without touching a real checkout.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pavestack/pave/internal/apiserver"
	"github.com/pavestack/pave/internal/workspace"
)

func main() {
	repoRoot := os.Getenv("PAVE_API_REPO_ROOT")
	if repoRoot == "" {
		root, err := workspace.Root()
		if err != nil {
			log.Fatalf("resolve repo root: %v (set PAVE_API_REPO_ROOT explicitly)", err)
		}
		repoRoot = root
	}

	port := os.Getenv("PAVE_API_PORT")
	if port == "" {
		port = "8787"
	}

	// Default to dry-run: a demo/public deployment of pave-api should never
	// silently open real pull requests against a real repo unless an
	// operator explicitly opts in.
	dryRun := true
	if v := os.Getenv("PAVE_API_DRY_RUN"); v == "false" {
		dryRun = false
	}

	cfg := apiserver.Config{
		RepoRoot:   repoRoot,
		DryRun:     dryRun,
		CORSOrigin: envOr("PAVE_API_CORS_ORIGIN", "*"),
	}

	srv, err := apiserver.New(cfg)
	if err != nil {
		log.Fatalf("start pave-api: %v", err)
	}

	if !apiserver.GitOpsToolsAvailable() {
		log.Println("warning: git and/or gh not found on PATH - open_pr steps will fail (or, in dry-run mode, are already simulated)")
	}

	addr := ":" + port
	log.Printf("pave-api listening on %s (repoRoot=%s, dryRun=%t)", addr, repoRoot, dryRun)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
