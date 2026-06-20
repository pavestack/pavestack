package validate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/spf13/afero"
)

type ServiceRequest struct {
	Name     string `json:"name"`
	Team     string `json:"team"`
	Database bool   `json:"database"`
}

type Validator struct {
	fs     afero.Fs
	schema *jsonschema.Schema
}

func NewValidator(fs afero.Fs, schemaBytes []byte) (*Validator, error) {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("service-request.schema.json", bytes.NewReader(schemaBytes)); err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}

	schema, err := compiler.Compile("service-request.schema.json")
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}

	return &Validator{
		fs:     fs,
		schema: schema,
	}, nil
}

func (v *Validator) Validate(repoRoot string, request ServiceRequest) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	var document any
	if err := json.Unmarshal(payload, &document); err != nil {
		return fmt.Errorf("decode request: %w", err)
	}

	if err := v.schema.Validate(document); err != nil {
		return fmt.Errorf("request validation failed: %w", err)
	}

	serviceDir := filepath.Join(repoRoot, "services", request.Name+"-api")
	if pathExists(v.fs, serviceDir) {
		return fmt.Errorf("service directory already exists: %s", serviceDir)
	}

	tenantDir := filepath.Join(repoRoot, "platform-config", "tenants", request.Name)
	if pathExists(v.fs, tenantDir) {
		return fmt.Errorf("tenant already exists: %s", tenantDir)
	}

	return nil
}

func ValidateServiceRequest(repoRoot string, request ServiceRequest) error {
	fs := afero.NewOsFs()
	schemaPath := filepath.Join(repoRoot, "pave", "schemas", "service-request.schema.json")
	schemaBytes, err := afero.ReadFile(fs, schemaPath)
	if err != nil {
		return fmt.Errorf("load schema: %w", err)
	}

	v, err := NewValidator(fs, schemaBytes)
	if err != nil {
		return err
	}

	return v.Validate(repoRoot, request)
}

func pathExists(fs afero.Fs, path string) bool {
	_, err := fs.Stat(path)
	return err == nil
}
