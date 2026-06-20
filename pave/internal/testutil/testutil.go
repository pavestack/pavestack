package testutil

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

// SetupWorkspace creates and populates a mock repository structure.
// If a filesystem is provided as a variadic parameter, it uses it and populates the structure under a virtual root path like "/workspace".
// If no filesystem is provided, it creates a physical temporary directory using t.TempDir(), wraps it in afero.NewOsFs(), and populates it.
func SetupWorkspace(t *testing.T, fsys ...afero.Fs) (afero.Fs, string) {
	t.Helper()

	var fs afero.Fs
	var root string

	if len(fsys) > 0 && fsys[0] != nil {
		fs = fsys[0]
		root = "/workspace"
	} else {
		root = t.TempDir()
		fs = afero.NewOsFs()
	}

	// 1. Create directory structure
	templateDir := filepath.Join(root, "service-template-api")
	schemaDir := filepath.Join(root, "pave", "schemas")

	dirs := []string{
		schemaDir,
		filepath.Join(templateDir, "cmd", "server"),
		filepath.Join(templateDir, "internal", "server"),
		filepath.Join(templateDir, "deploy", "helm", "service-template-api", "templates"),
		filepath.Join(root, "platform-config"),
		filepath.Join(root, "pave"),
	}

	for _, dir := range dirs {
		if err := fs.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// 2. Schema and template files content
	schema := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "PavestackServiceRequest",
  "type": "object",
  "additionalProperties": false,
  "required": ["name", "team", "database"],
  "properties": {
    "name": {
      "type": "string",
      "pattern": "^[a-z][a-z0-9-]{1,48}[a-z0-9]$",
      "description": "DNS-safe service identifier"
    },
    "team": {
      "type": "string",
      "pattern": "^[a-z][a-z0-9-]{1,62}[a-z0-9]$",
      "description": "Owning team slug"
    },
    "database": {
      "type": "boolean",
      "description": "Whether the service requires a managed database"
    }
  }
}`

	files := map[string]string{
		filepath.Join(schemaDir, "service-request.schema.json"): schema,
		filepath.Join(templateDir, "go.mod"):                    "module github.com/pavestack/service-template-api\n\ngo 1.23\n",
		filepath.Join(templateDir, "README.md"):                 "# service-template-api\n",
		filepath.Join(templateDir, "cmd", "server", "main.go"): `package main

import "github.com/pavestack/service-template-api/internal/server"

func main() { server.Run() }
`,
		filepath.Join(templateDir, "internal", "server", "server.go"): `package server

const name = "service-template-api"
`,
		filepath.Join(templateDir, "catalog-info.yaml"): `apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: service-template-api
  annotations:
    pavestack.io/team: platform
spec:
  owner: team-platform
`,
		filepath.Join(templateDir, "scorecard.yaml"): `service: service-template-api
owner: team-platform
overall_score: 100
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
		if err := afero.WriteFile(fs, path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return fs, root
}
