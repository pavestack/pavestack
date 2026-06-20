package gitops_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pavestack/pave/internal/gitops"
	"github.com/pavestack/pave/internal/validate"
)

func TestWriteTenantManifestsCreatesStructure(t *testing.T) {
	root := t.TempDir()
	serviceDir := filepath.Join(root, "services", "payments-api")
	helmDir := filepath.Join(serviceDir, "deploy", "helm", "payments-api")
	if err := os.MkdirAll(helmDir, 0o755); err != nil {
		t.Fatal(err)
	}

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	if err := gitops.WriteTenantManifests(root, request, serviceDir); err != nil {
		t.Fatalf("WriteTenantManifests failed: %v", err)
	}

	tenantRoot := filepath.Join(root, "platform-config", "tenants", "payments")

	// Verify directory structure
	for _, sub := range []string{"base", "dev", "prod"} {
		path := filepath.Join(tenantRoot, sub)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected directory %s to exist: %v", sub, err)
		}
	}
}

func TestWriteTenantManifestsTenantYAML(t *testing.T) {
	root := t.TempDir()
	serviceDir := filepath.Join(root, "services", "payments-api")
	helmDir := filepath.Join(serviceDir, "deploy", "helm", "payments-api")
	if err := os.MkdirAll(helmDir, 0o755); err != nil {
		t.Fatal(err)
	}

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: true,
	}

	if err := gitops.WriteTenantManifests(root, request, serviceDir); err != nil {
		t.Fatalf("WriteTenantManifests failed: %v", err)
	}

	tenantYAML, err := os.ReadFile(filepath.Join(root, "platform-config", "tenants", "payments", "tenant.yaml"))
	if err != nil {
		t.Fatalf("failed to read tenant.yaml: %v", err)
	}

	content := string(tenantYAML)
	checks := []string{
		"namespace: payments",
		"releaseName: payments-api",
		"owner: team-payments",
		"database: true",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("tenant.yaml missing %q:\n%s", check, content)
		}
	}
}

func TestWriteTenantManifestsDevValues(t *testing.T) {
	root := t.TempDir()
	serviceDir := filepath.Join(root, "services", "payments-api")
	helmDir := filepath.Join(serviceDir, "deploy", "helm", "payments-api")
	if err := os.MkdirAll(helmDir, 0o755); err != nil {
		t.Fatal(err)
	}

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	if err := gitops.WriteTenantManifests(root, request, serviceDir); err != nil {
		t.Fatalf("WriteTenantManifests failed: %v", err)
	}

	devValues, err := os.ReadFile(filepath.Join(root, "platform-config", "tenants", "payments", "dev", "values.yaml"))
	if err != nil {
		t.Fatalf("failed to read dev/values.yaml: %v", err)
	}

	content := string(devValues)
	if !strings.Contains(content, "replicaCount: 1") {
		t.Error("dev values should have replicaCount: 1")
	}
	if !strings.Contains(content, "LOG_LEVEL: debug") {
		t.Error("dev values should have LOG_LEVEL: debug")
	}
	if !strings.Contains(content, "SERVICE_NAME: payments-api") {
		t.Error("dev values should reference payments-api")
	}
}

func TestWriteTenantManifestsProdValues(t *testing.T) {
	root := t.TempDir()
	serviceDir := filepath.Join(root, "services", "payments-api")
	helmDir := filepath.Join(serviceDir, "deploy", "helm", "payments-api")
	if err := os.MkdirAll(helmDir, 0o755); err != nil {
		t.Fatal(err)
	}

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	if err := gitops.WriteTenantManifests(root, request, serviceDir); err != nil {
		t.Fatalf("WriteTenantManifests failed: %v", err)
	}

	prodValues, err := os.ReadFile(filepath.Join(root, "platform-config", "tenants", "payments", "prod", "values.yaml"))
	if err != nil {
		t.Fatalf("failed to read prod/values.yaml: %v", err)
	}

	content := string(prodValues)
	if !strings.Contains(content, "replicaCount: 2") {
		t.Error("prod values should have replicaCount: 2")
	}
	if !strings.Contains(content, "LOG_LEVEL: info") {
		t.Error("prod values should have LOG_LEVEL: info")
	}
}

func TestWriteTenantManifestsApplication(t *testing.T) {
	root := t.TempDir()
	serviceDir := filepath.Join(root, "services", "payments-api")
	helmDir := filepath.Join(serviceDir, "deploy", "helm", "payments-api")
	if err := os.MkdirAll(helmDir, 0o755); err != nil {
		t.Fatal(err)
	}

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	if err := gitops.WriteTenantManifests(root, request, serviceDir); err != nil {
		t.Fatalf("WriteTenantManifests failed: %v", err)
	}

	for _, env := range []string{"dev", "prod"} {
		app, err := os.ReadFile(filepath.Join(root, "platform-config", "tenants", "payments", env, "application.yaml"))
		if err != nil {
			t.Fatalf("failed to read %s/application.yaml: %v", env, err)
		}

		content := string(app)
		if !strings.Contains(content, "name: payments-api-"+env) {
			t.Errorf("%s application missing correct name", env)
		}
		if !strings.Contains(content, "pavestack.io/tenant: payments") {
			t.Errorf("%s application missing tenant label", env)
		}
		if !strings.Contains(content, "pavestack.io/environment: "+env) {
			t.Errorf("%s application missing environment label", env)
		}
		if !strings.Contains(content, "namespace: payments") {
			t.Errorf("%s application missing destination namespace", env)
		}
	}
}

func TestWriteTenantManifestsBaseKustomization(t *testing.T) {
	root := t.TempDir()
	serviceDir := filepath.Join(root, "services", "payments-api")
	helmDir := filepath.Join(serviceDir, "deploy", "helm", "payments-api")
	if err := os.MkdirAll(helmDir, 0o755); err != nil {
		t.Fatal(err)
	}

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	if err := gitops.WriteTenantManifests(root, request, serviceDir); err != nil {
		t.Fatalf("WriteTenantManifests failed: %v", err)
	}

	kustomization, err := os.ReadFile(filepath.Join(root, "platform-config", "tenants", "payments", "base", "kustomization.yaml"))
	if err != nil {
		t.Fatalf("failed to read kustomization.yaml: %v", err)
	}

	content := string(kustomization)
	checks := []string{
		"namespace: payments",
		"../../../templates/namespace",
		"../../../templates/rbac",
		"../../../templates/network-policy",
		"../../../templates/resource-quota",
		"value: payments",
		"value: team-payments",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("kustomization.yaml missing %q", check)
		}
	}
}

func TestCreatePullRequestMissingGitOrGh(t *testing.T) {
	root := t.TempDir()
	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	// Mock PATH to be empty so git or gh cannot be found
	t.Setenv("PATH", "")

	err := gitops.CreatePullRequest(root, request, "")
	if err == nil {
		t.Fatal("expected error when git/gh are not in PATH")
	}
}

func TestCreatePullRequestFailsWhenNotGitRepo(t *testing.T) {
	root := t.TempDir()
	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	// Call it, git and gh should be in the actual PATH on this mac (or not, if not we still get error)
	// If they are in PATH, it will fail because it's not a git repo.
	// If they are not in PATH, it will fail because they are not found.
	// Either way, an error must be returned.
	err := gitops.CreatePullRequest(root, request, "")
	if err == nil {
		t.Fatal("expected error running CreatePullRequest in non-git directory")
	}
}

func TestTenantManifestRenderer(t *testing.T) {
	renderer := gitops.NewTenantManifestRenderer()

	t.Run("RenderTenant", func(t *testing.T) {
		got, err := renderer.RenderTenant("foo", "services/foo-api/deploy/helm/foo-api", "team-foo", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "namespace: foo") {
			t.Errorf("expected tenant namespace; got:\n%s", got)
		}
		if !strings.Contains(got, "helmPath: services/foo-api/deploy/helm/foo-api") {
			t.Errorf("expected helmPath; got:\n%s", got)
		}
		if !strings.Contains(got, "database: true") {
			t.Errorf("expected database: true; got:\n%s", got)
		}
	})

	t.Run("RenderBaseKustomization", func(t *testing.T) {
		got, err := renderer.RenderBaseKustomization("foo", "team-foo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "namespace: foo") {
			t.Errorf("expected namespace; got:\n%s", got)
		}
		if !strings.Contains(got, "value: team-foo") {
			t.Errorf("expected team-foo patch; got:\n%s", got)
		}
	})

	t.Run("RenderValues", func(t *testing.T) {
		got, err := renderer.RenderValues("foo", "123.dkr.ecr.us-east-1.amazonaws.com/pavestack/foo-api", 3, "warn")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "replicaCount: 3") {
			t.Errorf("expected replicaCount: 3; got:\n%s", got)
		}
		if !strings.Contains(got, "repository: 123.dkr.ecr.us-east-1.amazonaws.com/pavestack/foo-api") {
			t.Errorf("expected repository; got:\n%s", got)
		}
		if !strings.Contains(got, "LOG_LEVEL: warn") {
			t.Errorf("expected LOG_LEVEL: warn; got:\n%s", got)
		}
	})

	t.Run("RenderApplication", func(t *testing.T) {
		got, err := renderer.RenderApplication("foo", "services/foo-api/deploy/helm/foo-api", "staging")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "name: foo-api-staging") {
			t.Errorf("expected staging name; got:\n%s", got)
		}
		if !strings.Contains(got, "pavestack.io/environment: staging") {
			t.Errorf("expected staging env; got:\n%s", got)
		}
		if !strings.Contains(got, "- $values/platform-config/tenants/foo/staging/values.yaml") {
			t.Errorf("expected value file path; got:\n%s", got)
		}
	})
}

func TestVersionControlValidateTools(t *testing.T) {
	vc := gitops.NewVersionControl(t.TempDir())
	t.Run("With Empty PATH", func(t *testing.T) {
		t.Setenv("PATH", "")
		err := vc.ValidateTools()
		if err == nil {
			t.Fatal("expected error with empty PATH")
		}
	})
}
