// Package openapi embeds the service's contract-first OpenAPI 3.1
// specification (openapi.yaml) and exposes it as JSON so the running
// service can serve it back at /openapi.json.
package openapi

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed openapi.yaml
var specYAML []byte

// YAML returns the raw embedded OpenAPI specification.
func YAML() []byte {
	return specYAML
}

// JSON converts the embedded OpenAPI specification to JSON.
func JSON() ([]byte, error) {
	var doc any
	if err := yaml.Unmarshal(specYAML, &doc); err != nil {
		return nil, fmt.Errorf("parse openapi.yaml: %w", err)
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshal openapi spec to json: %w", err)
	}
	return body, nil
}
