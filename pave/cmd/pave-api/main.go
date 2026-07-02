// Command pave-api is the HTTP backend for the Pavestack self-service
// portal's "Create service" and "Request access" flows. It is a thin HTTP
// shell around the same scaffold/gitops/validate packages `pave
// create-service` uses directly - see pave/internal/apiserver.
//
// Running this against your own checkout of the Pavestack monorepo will
// really scaffold services and (unless PAVE_API_DRY_RUN=true, which is the
// default) really open pull requests via the `gh` CLI, exactly like the
// CLI does. Point PAVE_API_REPO_ROOT at a scratch clone if you want to try
// it without touching a real checkout. See .env.example for the full list
// of environment variables this reads.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pavestack/pave/internal/app"
)

func main() {
	if err := app.New().Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
