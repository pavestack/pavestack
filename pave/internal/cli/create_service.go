package cli

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/pavestack/pave/internal/cost"
	"github.com/pavestack/pave/internal/gitops"
	"github.com/pavestack/pave/internal/scaffold"
	"github.com/pavestack/pave/internal/validate"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type createServiceOptions struct {
	Name     string
	Team     string
	Database bool
	Runtime  string
	Exposure string
	Tier     string
	NoPR     bool
	Branch   string
}

func newCreateServiceCmd() *cobra.Command {
	opts := &createServiceOptions{}

	cmd := &cobra.Command{
		Use:   "create-service",
		Short: "Scaffold a new internal API service and GitOps manifests",
		RunE: func(cmd *cobra.Command, _ []string) error {
			databaseProvided := cmd.Flags().Changed("database")
			if err := promptMissing(cmd.InOrStdin(), opts, databaseProvided); err != nil {
				return err
			}

			root, err := repoRoot()
			if err != nil {
				return err
			}

			request := validate.ServiceRequest{
				Name:     opts.Name,
				Team:     opts.Team,
				Database: opts.Database,
				Runtime:  opts.Runtime,
				Exposure: opts.Exposure,
				Tier:     opts.Tier,
			}
			request.ApplyDefaults()

			estimate := cost.Estimate(cost.Tier(request.Tier), request.Exposure, request.Database)
			fmt.Fprintf(cmd.OutOrStdout(), "\nEstimated cost (%s, %s, database=%t): $%.0f-%.0f/month\n", estimate.Tier, request.Exposure, request.Database, estimate.MonthlyUSDLow, estimate.MonthlyUSDHigh)
			for _, item := range estimate.Breakdown {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s: $%.0f-%.0f/month\n", item.Item, item.MonthlyUSDLow, item.MonthlyUSD)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  (%s)\n\n", estimate.Disclaimer)

			fs := afero.NewOsFs()
			schemaPath := filepath.Join(root, "pave", "schemas", "service-request.schema.json")
			schemaBytes, err := afero.ReadFile(fs, schemaPath)
			if err != nil {
				return fmt.Errorf("load schema: %w", err)
			}
			v, err := validate.NewValidator(fs, schemaBytes)
			if err != nil {
				return err
			}
			if err := v.Validate(root, request); err != nil {
				return err
			}

			serviceDir, err := scaffold.CreateService(afero.NewOsFs(), root, request)
			if err != nil {
				return err
			}

			if err := gitops.WriteTenantManifests(root, request, serviceDir); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created service at %s\n", serviceDir)
			fmt.Fprintf(cmd.OutOrStdout(), "Created GitOps tenant at platform-config/tenants/%s\n", request.Name)

			if !opts.NoPR {
				if err := gitops.CreatePullRequest(root, request, opts.Branch); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: PR creation skipped: %v\n", err)
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Next: commit, push, and let Argo CD reconcile after merge.")
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Service name (DNS-safe slug)")
	cmd.Flags().StringVar(&opts.Team, "team", "", "Owning team slug")
	cmd.Flags().BoolVar(&opts.Database, "database", false, "Provision managed database")
	cmd.Flags().StringVar(&opts.Runtime, "runtime", cost.DefaultRuntime, "Golden-path runtime (only 'go' is scaffoldable today)")
	cmd.Flags().StringVar(&opts.Exposure, "exposure", cost.DefaultExposure, "Service exposure: internal (ClusterIP) or public (ALB)")
	cmd.Flags().StringVar(&opts.Tier, "tier", string(cost.DefaultTier), "Reliability/sizing tier: tier-1 (critical), tier-2 (standard), tier-3 (low-traffic)")
	cmd.Flags().BoolVar(&opts.NoPR, "no-pr", false, "Skip automatic pull request creation")
	cmd.Flags().StringVar(&opts.Branch, "branch", "", "Branch name for pull request")

	return cmd
}

func promptMissing(in io.Reader, opts *createServiceOptions, databaseProvided bool) error {
	reader := bufio.NewReader(in)

	if opts.Name == "" {
		fmt.Print("Service name: ")
		value, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		opts.Name = strings.TrimSpace(value)
	}

	if opts.Team == "" {
		fmt.Print("Team owner: ")
		value, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		opts.Team = strings.TrimSpace(value)
	}

	if !databaseProvided {
		fmt.Print("Requires database? [y/N]: ")
		value, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		value = strings.TrimSpace(strings.ToLower(value))
		if value == "y" || value == "yes" || value == "true" {
			opts.Database = true
		}
	}

	return nil
}
