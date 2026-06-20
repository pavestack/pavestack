package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// 1. Create schemas directory and schema file
	schemaDir := filepath.Join(root, "pave", "schemas")
	if err := os.MkdirAll(schemaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	schema := `{
	  "$schema": "http://json-schema.org/draft-07/schema#",
	  "type": "object",
	  "required": ["name", "team", "database"],
	  "properties": {
	    "name": { "type": "string", "pattern": "^[a-z][a-z0-9-]{1,48}[a-z0-9]$" },
	    "team": { "type": "string", "pattern": "^[a-z][a-z0-9-]{1,62}[a-z0-9]$" },
	    "database": { "type": "boolean" }
	  }
	}`
	if err := os.WriteFile(filepath.Join(schemaDir, "service-request.schema.json"), []byte(schema), 0o644); err != nil {
		t.Fatal(err)
	}

	// 2. Create service-template-api mock directory structure
	templateDir := filepath.Join(root, "service-template-api")
	dirs := []string{
		filepath.Join(templateDir, "cmd", "server"),
		filepath.Join(templateDir, "internal", "server"),
		filepath.Join(templateDir, "deploy", "helm", "service-template-api", "templates"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	files := map[string]string{
		filepath.Join(templateDir, "go.mod"):    "module github.com/pavestack/service-template-api\n\ngo 1.23\n",
		filepath.Join(templateDir, "README.md"): "# service-template-api\n",
		filepath.Join(templateDir, "cmd", "server", "main.go"): `package main
func main() {}
`,
		filepath.Join(templateDir, "internal", "server", "server.go"): `package server
const name = "service-template-api"
`,
		filepath.Join(templateDir, "catalog-info.yaml"): `apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: service-template-api
spec:
  owner: team-platform
`,
		filepath.Join(templateDir, "deploy", "helm", "service-template-api", "Chart.yaml"): `apiVersion: v2
name: service-template-api
version: 0.1.0
`,
		filepath.Join(templateDir, "deploy", "helm", "service-template-api", "values.yaml"): `replicaCount: 2
image:
  repository: 123456789012.dkr.ecr.us-east-1.amazonaws.com/pavestack/service-template-api
env:
  SERVICE_NAME: service-template-api
`,
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// 3. Create other required workspace directories
	for _, name := range []string{"platform-config", "pave"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	return root
}

func TestCreateServiceCmd(t *testing.T) {
	root := setupTestWorkspace(t)
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
	root := setupTestWorkspace(t)
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
