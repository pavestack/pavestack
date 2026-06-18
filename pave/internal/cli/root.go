package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "pave",
	Short:   "Pavestack self-service developer CLI",
	Long:    "Pavestack CLI scaffolds services and GitOps manifests. It never mutates live clusters directly.",
	Version: Version,
}

func Execute() error {
	rootCmd.AddCommand(newCreateServiceCmd())
	return rootCmd.Execute()
}

func repoRoot() (string, error) {
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
