// Package api embeds pave-api's OpenAPI contract so it can be served at
// runtime (GET /api/v1/openapi.json) without a separate build/copy step.
// See pave/API_VERSIONING.md for how this contract is versioned.
package api

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed openapi.yaml
var openAPIYAML []byte

// OpenAPIYAML returns the raw OpenAPI 3.0 spec as authored.
func OpenAPIYAML() []byte {
	return openAPIYAML
}

// OpenAPIJSON converts the embedded YAML spec to JSON - most OpenAPI
// tooling (Swagger UI, codegen) expects JSON. yaml.v3 unmarshals mapping
// nodes into map[string]interface{}, which encoding/json marshals
// natively, so this is a direct, lossless conversion.
func OpenAPIJSON() ([]byte, error) {
	var doc any
	if err := yaml.Unmarshal(openAPIYAML, &doc); err != nil {
		return nil, fmt.Errorf("parse embedded openapi.yaml: %w", err)
	}
	return json.Marshal(doc)
}
