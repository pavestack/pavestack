package validate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pavestack/pave/internal/validate"
	"github.com/spf13/afero"
)

func TestValidateServiceRequest(t *testing.T) {
	root := t.TempDir()
	copySchema(t, root)

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: true,
	}

	if err := validate.ValidateServiceRequest(root, request); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
}

func TestValidateServiceRequestRejectsInvalidName(t *testing.T) {
	root := t.TempDir()
	copySchema(t, root)

	request := validate.ServiceRequest{
		Name:     "Payments",
		Team:     "team-payments",
		Database: false,
	}

	if err := validate.ValidateServiceRequest(root, request); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateServiceRequestRejectsExistingServiceDir(t *testing.T) {
	root := t.TempDir()
	copySchema(t, root)

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	// Create service directory
	serviceDir := filepath.Join(root, "services", "payments-api")
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := validate.ValidateServiceRequest(root, request); err == nil {
		t.Fatal("expected validation error due to existing service dir")
	}
}

func TestValidateServiceRequestRejectsExistingTenantDir(t *testing.T) {
	root := t.TempDir()
	copySchema(t, root)

	request := validate.ServiceRequest{
		Name:     "payments",
		Team:     "team-payments",
		Database: false,
	}

	// Create tenant directory
	tenantDir := filepath.Join(root, "platform-config", "tenants", "payments")
	if err := os.MkdirAll(tenantDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := validate.ValidateServiceRequest(root, request); err == nil {
		t.Fatal("expected validation error due to existing tenant dir")
	}
}

func copySchema(t *testing.T, root string) {
	t.Helper()
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
}

func TestValidator(t *testing.T) {
	fs := afero.NewMemMapFs()
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

	v, err := validate.NewValidator(fs, []byte(schema))
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	t.Run("ValidRequest", func(t *testing.T) {
		req := validate.ServiceRequest{
			Name:     "payments",
			Team:     "team-payments",
			Database: true,
		}
		if err := v.Validate("/workspace", req); err != nil {
			t.Fatalf("expected valid request, got %v", err)
		}
	})

	t.Run("InvalidName", func(t *testing.T) {
		req := validate.ServiceRequest{
			Name:     "Payments",
			Team:     "team-payments",
			Database: true,
		}
		if err := v.Validate("/workspace", req); err == nil {
			t.Fatal("expected validation error for uppercase Name")
		}
	})

	t.Run("ExistingServiceDir", func(t *testing.T) {
		req := validate.ServiceRequest{
			Name:     "payments",
			Team:     "team-payments",
			Database: true,
		}
		serviceDir := "/workspace/services/payments-api"
		if err := fs.MkdirAll(serviceDir, 0o755); err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = fs.RemoveAll(serviceDir)
		}()

		if err := v.Validate("/workspace", req); err == nil {
			t.Fatal("expected validation error due to existing service dir")
		}
	})

	t.Run("ExistingTenantDir", func(t *testing.T) {
		req := validate.ServiceRequest{
			Name:     "payments",
			Team:     "team-payments",
			Database: true,
		}
		tenantDir := "/workspace/platform-config/tenants/payments"
		if err := fs.MkdirAll(tenantDir, 0o755); err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = fs.RemoveAll(tenantDir)
		}()

		if err := v.Validate("/workspace", req); err == nil {
			t.Fatal("expected validation error due to existing tenant dir")
		}
	})
}
