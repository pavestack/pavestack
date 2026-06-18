package server_test

import (
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
