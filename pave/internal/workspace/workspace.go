// Package workspace resolves the Pavestack monorepo root. It is shared by
// the pave CLI (internal/cli) and the pave-api backend (cmd/pave-api) so the
// two entry points that both scaffold services agree on where "the repo" is
// without duplicating the directory-walk logic.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

// Root resolves the Pavestack repository root: $PAVESTACK_ROOT if set,
// otherwise the nearest ancestor of the current working directory that
// contains platform-config/, service-template-api/, and pave/.
func Root() (string, error) {
	if value := os.Getenv("PAVESTACK_ROOT"); value != "" {
		return value, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	for {
		if isRepoRoot(cwd) {
			return cwd, nil
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}

	return "", fmt.Errorf("could not find Pavestack repository root; set PAVESTACK_ROOT")
}

func isRepoRoot(path string) bool {
	required := []string{"platform-config", "service-template-api", "pave"}
	for _, name := range required {
		if _, err := os.Stat(filepath.Join(path, name)); err != nil {
			return false
		}
	}
	return true
}
