package cli

import (
	"github.com/pavestack/pave/internal/workspace"
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
	return workspace.Root()
}
