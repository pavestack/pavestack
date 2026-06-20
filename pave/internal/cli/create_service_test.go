package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pavestack/pave/internal/testutil"
)

func TestCreateServiceCmd(t *testing.T) {
	_, root := testutil.SetupWorkspace(t)
	t.Setenv("PAVESTACK_ROOT", root)

	cmd := newCreateServiceCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&outBuf)

	cmd.SetArgs([]string{
		"--name", "payments",
		"--team", "team-payments",
		"--database=true",
		"--no-pr",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "Created service at") {
		t.Errorf("expected output to contain 'Created service at', got %q", output)
	}
	if !strings.Contains(output, "Created GitOps tenant at") {
		t.Errorf("expected output to contain 'Created GitOps tenant at', got %q", output)
	}

	// Verify the scaffolded directories actually exist in the temp root
	serviceDir := filepath.Join(root, "services", "payments-api")
	if _, err := os.Stat(serviceDir); err != nil {
		t.Fatalf("expected scaffolded service directory to exist, got %v", err)
	}

	tenantDir := filepath.Join(root, "platform-config", "tenants", "payments")
	if _, err := os.Stat(tenantDir); err != nil {
		t.Fatalf("expected tenant directory to exist, got %v", err)
	}
}

func TestCreateServiceCmdValidationFailure(t *testing.T) {
	_, root := testutil.SetupWorkspace(t)
	t.Setenv("PAVESTACK_ROOT", root)

	cmd := newCreateServiceCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&outBuf)

	// Payments with capital P is invalid according to schema regex
	cmd.SetArgs([]string{
		"--name", "Payments",
		"--team", "team-payments",
		"--database=false",
		"--no-pr",
	})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected execution error due to name validation failure")
	}
}

func TestPromptMissing(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		initialOpts      createServiceOptions
		databaseProvided bool
		expectedOpts     createServiceOptions
		expectErr        bool
	}{
		{
			name:             "Prompt all values - database yes",
			input:            "my-service\nmy-team\ny\n",
			initialOpts:      createServiceOptions{},
			databaseProvided: false,
			expectedOpts: createServiceOptions{
				Name:     "my-service",
				Team:     "my-team",
				Database: true,
			},
			expectErr: false,
		},
		{
			name:             "Prompt all values - database no",
			input:            "my-service\nmy-team\nn\n",
			initialOpts:      createServiceOptions{},
			databaseProvided: false,
			expectedOpts: createServiceOptions{
				Name:     "my-service",
				Team:     "my-team",
				Database: false,
			},
			expectErr: false,
		},
		{
			name:             "Prompt partial - only database prompt needed",
			input:            "yes\n",
			initialOpts:      createServiceOptions{Name: "my-service", Team: "my-team"},
			databaseProvided: false,
			expectedOpts: createServiceOptions{
				Name:     "my-service",
				Team:     "my-team",
				Database: true,
			},
			expectErr: false,
		},
		{
			name:             "Prompt partial - only team and database",
			input:            "my-team\nno\n",
			initialOpts:      createServiceOptions{Name: "my-service"},
			databaseProvided: false,
			expectedOpts: createServiceOptions{
				Name:     "my-service",
				Team:     "my-team",
				Database: false,
			},
			expectErr: false,
		},
		{
			name:             "Database already provided - no prompt for db",
			input:            "my-service\nmy-team\n",
			initialOpts:      createServiceOptions{Database: true},
			databaseProvided: true,
			expectedOpts: createServiceOptions{
				Name:     "my-service",
				Team:     "my-team",
				Database: true,
			},
			expectErr: false,
		},
		{
			name:             "Empty inputs are accepted but trimmed",
			input:            "  my-service  \n  my-team  \n  true  \n",
			initialOpts:      createServiceOptions{},
			databaseProvided: false,
			expectedOpts: createServiceOptions{
				Name:     "my-service",
				Team:     "my-team",
				Database: true,
			},
			expectErr: false,
		},
		{
			name:             "Input EOF before completing",
			input:            "my-service\n",
			initialOpts:      createServiceOptions{},
			databaseProvided: false,
			expectedOpts:     createServiceOptions{Name: "my-service"},
			expectErr:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := strings.NewReader(tc.input)
			opts := tc.initialOpts
			err := promptMissing(in, &opts, tc.databaseProvided)
			if tc.expectErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if opts.Name != tc.expectedOpts.Name {
					t.Errorf("expected name %q, got %q", tc.expectedOpts.Name, opts.Name)
				}
				if opts.Team != tc.expectedOpts.Team {
					t.Errorf("expected team %q, got %q", tc.expectedOpts.Team, opts.Team)
				}
				if opts.Database != tc.expectedOpts.Database {
					t.Errorf("expected database %v, got %v", tc.expectedOpts.Database, opts.Database)
				}
			}
		})
	}
}
