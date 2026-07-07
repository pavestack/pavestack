package server

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	openapi "github.com/pavestack/service-template-api"
	"github.com/pavestack/service-template-api/internal/config"
	"github.com/pavestack/service-template-api/internal/logging"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type Server struct {
	cfg         config.Config
	logger      *zap.Logger
	ready       atomic.Bool
	openapiJSON []byte
}

func New(cfg config.Config, logger *zap.Logger) *Server {
	s := &Server{cfg: cfg, logger: logger}
	s.ready.Store(cfg.Ready)

	body, err := openapi.JSON()
	if err != nil {
		logger.Error("failed to render openapi spec", logging.Error(err))
	} else {
		s.openapiJSON = body
	}

	return s
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /ready", s.handleReady)
	mux.HandleFunc("GET /openapi.json", s.handleOpenAPI)
	return otelhttp.NewHandler(mux, "http.server")
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": s.cfg.ServiceName,
	})
}

func (s *Server) handleReady(w http.ResponseWriter, _ *http.Request) {
	if !s.ready.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "not_ready",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

// handleOpenAPI serves the service's OpenAPI 3.1 contract as JSON, converted
// at startup from the embedded openapi.yaml (see openapi.go at the repo root).
func (s *Server) handleOpenAPI(w http.ResponseWriter, _ *http.Request) {
	if s.openapiJSON == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "openapi spec unavailable",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(s.openapiJSON)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
