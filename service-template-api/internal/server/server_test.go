package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pavestack/service-template-api/internal/config"
	"github.com/pavestack/service-template-api/internal/logging"
	"github.com/pavestack/service-template-api/internal/server"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := config.Config{ServiceName: "test-service", Ready: true}
	srv := server.New(cfg, logging.New("error"))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReadyEndpoint(t *testing.T) {
	cfg := config.Config{ServiceName: "test-service", Ready: true}
	srv := server.New(cfg, logging.New("error"))

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReadyEndpointNotReady(t *testing.T) {
	cfg := config.Config{ServiceName: "test-service", Ready: false}
	srv := server.New(cfg, logging.New("error"))

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestOpenAPIEndpoint(t *testing.T) {
	cfg := config.Config{ServiceName: "test-service", Ready: true}
	srv := server.New(cfg, logging.New("error"))

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected content-type application/json, got %q", ct)
	}

	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}

	if version, _ := doc["openapi"].(string); version != "3.1.0" {
		t.Fatalf("expected openapi field 3.1.0, got %q", version)
	}
}
