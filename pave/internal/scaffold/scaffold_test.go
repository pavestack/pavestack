package scaffold_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/pavestack/pave/internal/scaffold"
	"github.com/pavestack/pave/internal/testutil"
	"github.com/pavestack/pave/internal/validate"
	"github.com/spf13/afero"
)

func TestCreateServiceCopiesTemplate(t *testing.T) {
	fsys, root := testutil.SetupWorkspace(t, afero.NewMemMapFs())

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	serviceDir, err := scaffold.CreateService(fsys, root, request)
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	expected := filepath.Join(root, "services", "payments-api")
	if serviceDir != expected {
		t.Errorf("expected service dir %s, got %s", expected, serviceDir)
	}

	if _, err := fsys.Stat(serviceDir); err != nil {
		t.Fatalf("service directory does not exist: %v", err)
	}
}

func TestCreateServiceReplaceNames(t *testing.T) {
	fsys, root := testutil.SetupWorkspace(t, afero.NewMemMapFs())

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	serviceDir, err := scaffold.CreateService(fsys, root, request)
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	// Check go.mod has replaced module path
	gomod, err := afero.ReadFile(fsys, filepath.Join(serviceDir, "go.mod"))
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}
	if !strings.Contains(string(gomod), "github.com/pavestack/services/payments-api") {
		t.Errorf("go.mod module path not replaced: %s", string(gomod))
	}

	// Check catalog-info.yaml has replaced owner
	catalog, err := afero.ReadFile(fsys, filepath.Join(serviceDir, "catalog-info.yaml"))
	if err != nil {
		t.Fatalf("failed to read catalog-info.yaml: %v", err)
	}
	if !strings.Contains(string(catalog), "team-payments") {
		t.Errorf("catalog owner not replaced: %s", string(catalog))
	}
}

func TestCreateServiceRenamesHelmChart(t *testing.T) {
	fsys, root := testutil.SetupWorkspace(t, afero.NewMemMapFs())

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	serviceDir, err := scaffold.CreateService(fsys, root, request)
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	newChart := filepath.Join(serviceDir, "deploy", "helm", "payments-api")
	if _, err := fsys.Stat(newChart); err != nil {
		t.Fatalf("helm chart not renamed: %v", err)
	}

	oldChart := filepath.Join(serviceDir, "deploy", "helm", "service-template-api")
	if _, err := fsys.Stat(oldChart); err == nil {
		t.Error("old helm chart directory still exists")
	}
}

func TestCreateServiceWithDatabase(t *testing.T) {
	fsys, root := testutil.SetupWorkspace(t, afero.NewMemMapFs())

	request := validate.ServiceRequest{
		Name:     "orders",
		Team:     "team-commerce",
		Database: true,
	}

	serviceDir, err := scaffold.CreateService(fsys, root, request)
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	readme, err := afero.ReadFile(fsys, filepath.Join(serviceDir, "README.md"))
	if err != nil {
		t.Fatalf("failed to read README: %v", err)
	}
	if !strings.Contains(string(readme), "Database") {
		t.Error("database stub not appended to README")
	}
}

func TestCreateServiceTemplatesOpenAPISpec(t *testing.T) {
	fsys, root := testutil.SetupWorkspace(t, afero.NewMemMapFs())

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	serviceDir, err := scaffold.CreateService(fsys, root, request)
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	spec, err := afero.ReadFile(fsys, filepath.Join(serviceDir, "openapi.yaml"))
	if err != nil {
		t.Fatalf("openapi.yaml not scaffolded: %v", err)
	}

	content := string(spec)
	if strings.Contains(content, "service-template-api") {
		t.Errorf("openapi.yaml still references the template name: %s", content)
	}
	if !strings.Contains(content, "title: payments-api") {
		t.Errorf("openapi.yaml title not substituted: %s", content)
	}
	if !strings.Contains(content, "payments-api HTTP server") {
		t.Errorf("openapi.yaml server description not substituted: %s", content)
	}
}

func TestCreateServiceWritesMetadata(t *testing.T) {
	fsys, root := testutil.SetupWorkspace(t, afero.NewMemMapFs())

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	serviceDir, err := scaffold.CreateService(fsys, root, request)
	if err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	metaPath := filepath.Join(serviceDir, ".pavestack", "service-request.json")
	data, err := afero.ReadFile(fsys, metaPath)
	if err != nil {
		t.Fatalf("metadata not written: %v", err)
	}
	if !strings.Contains(string(data), `"name": "payments"`) {
		t.Errorf("metadata does not contain service name: %s", string(data))
	}
}
