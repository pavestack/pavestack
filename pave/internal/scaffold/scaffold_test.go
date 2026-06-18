package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pavestack/pave/internal/scaffold"
	"github.com/pavestack/pave/internal/validate"
)

func setupRepoRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

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
		filepath.Join(templateDir, "go.mod"):   "module github.com/pavestack/service-template-api\n\ngo 1.23\n",
		filepath.Join(templateDir, "README.md"): "# service-template-api\n",
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
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create dirs expected by pave repoRoot detection
	for _, name := range []string{"platform-config", "pave"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	return root
}

func TestCreateServiceCopiesTemplate(t *testing.T) {
	root := setupRepoRoot(t)

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	serviceDir, err := scaffold.CreateService(root, request)
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	expected := filepath.Join(root, "services", "payments-api")
	if serviceDir != expected {
		t.Errorf("expected service dir %s, got %s", expected, serviceDir)
	}

	if _, err := os.Stat(serviceDir); err != nil {
		t.Fatalf("service directory does not exist: %v", err)
	}
}

func TestCreateServiceReplaceNames(t *testing.T) {
	root := setupRepoRoot(t)

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	serviceDir, err := scaffold.CreateService(root, request)
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	// Check go.mod has replaced module path
	gomod, err := os.ReadFile(filepath.Join(serviceDir, "go.mod"))
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}
	if !strings.Contains(string(gomod), "github.com/pavestack/services/payments-api") {
		t.Errorf("go.mod module path not replaced: %s", string(gomod))
	}

	// Check catalog-info.yaml has replaced owner
	catalog, err := os.ReadFile(filepath.Join(serviceDir, "catalog-info.yaml"))
	if err != nil {
		t.Fatalf("failed to read catalog-info.yaml: %v", err)
	}
	if !strings.Contains(string(catalog), "team-payments") {
		t.Errorf("catalog owner not replaced: %s", string(catalog))
	}
}

func TestCreateServiceRenamesHelmChart(t *testing.T) {
	root := setupRepoRoot(t)

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	serviceDir, err := scaffold.CreateService(root, request)
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	newChart := filepath.Join(serviceDir, "deploy", "helm", "payments-api")
	if _, err := os.Stat(newChart); err != nil {
		t.Fatalf("helm chart not renamed: %v", err)
	}

	oldChart := filepath.Join(serviceDir, "deploy", "helm", "service-template-api")
	if _, err := os.Stat(oldChart); err == nil {
		t.Error("old helm chart directory still exists")
	}
}

func TestCreateServiceWithDatabase(t *testing.T) {
	root := setupRepoRoot(t)

	request := validate.ServiceRequest{
		Name:     "orders",
		Team:     "team-commerce",
		Database: true,
	}

	serviceDir, err := scaffold.CreateService(root, request)
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	readme, err := os.ReadFile(filepath.Join(serviceDir, "README.md"))
	if err != nil {
		t.Fatalf("failed to read README: %v", err)
	}
	if !strings.Contains(string(readme), "Database") {
		t.Error("database stub not appended to README")
	}
}

func TestCreateServiceWritesMetadata(t *testing.T) {
	root := setupRepoRoot(t)

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	serviceDir, err := scaffold.CreateService(root, request)
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	metaPath := filepath.Join(serviceDir, ".pavestack", "service-request.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("metadata not written: %v", err)
	}
	if !strings.Contains(string(data), `"name": "payments"`) {
		t.Errorf("metadata does not contain service name: %s", string(data))
	}
}
