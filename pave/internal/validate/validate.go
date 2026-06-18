package validate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

type ServiceRequest struct {
	Name     string `json:"name"`
	Team     string `json:"team"`
	Database bool   `json:"database"`
}

func ValidateServiceRequest(repoRoot string, request ServiceRequest) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	schemaPath := filepath.Join(repoRoot, "pave", "schemas", "service-request.schema.json")
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("load schema: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("service-request.schema.json", bytes.NewReader(schemaBytes)); err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}

	schema, err := compiler.Compile("service-request.schema.json")
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}

	var document any
	if err := json.Unmarshal(payload, &document); err != nil {
		return fmt.Errorf("decode request: %w", err)
	}

	if err := schema.Validate(document); err != nil {
		return fmt.Errorf("request validation failed: %w", err)
	}

	serviceDir := filepath.Join(repoRoot, "services", request.Name+"-api")
	if pathExists(serviceDir) {
		return fmt.Errorf("service directory already exists: %s", serviceDir)
	}

	tenantDir := filepath.Join(repoRoot, "platform-config", "tenants", request.Name)
	if pathExists(tenantDir) {
		return fmt.Errorf("tenant already exists: %s", tenantDir)
	}

	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
