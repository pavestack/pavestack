package api

import (
	"encoding/json"
	"testing"
)

func TestOpenAPIYAMLIsNonEmpty(t *testing.T) {
	if len(OpenAPIYAML()) == 0 {
		t.Fatal("expected embedded openapi.yaml to be non-empty")
	}
}

func TestOpenAPIJSONIsValidAndMatchesYAML(t *testing.T) {
	raw, err := OpenAPIJSON()
	if err != nil {
		t.Fatalf("OpenAPIJSON: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("expected valid JSON, got: %v", err)
	}

	if doc["openapi"] != "3.0.3" {
		t.Errorf("expected openapi version 3.0.3, got %v", doc["openapi"])
	}
	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		t.Fatal("expected a paths object")
	}
	for _, want := range []string{"/healthz", "/api/v1/services", "/api/v1/access-requests", "/auth/github/login"} {
		if _, ok := paths[want]; !ok {
			t.Errorf("expected %s to be documented in paths", want)
		}
	}
}
