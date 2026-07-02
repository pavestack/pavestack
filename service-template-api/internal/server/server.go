package server

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/pavestack/service-template-api/internal/config"
	"github.com/pavestack/service-template-api/internal/logging"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type Server struct {
	cfg    config.Config
	logger *zap.Logger
	ready  atomic.Bool
}

func New(cfg config.Config, logger *zap.Logger) *Server {
	s := &Server{cfg: cfg, logger: logger}
	s.ready.Store(cfg.Ready)
	return s
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /ready", s.handleReady)
	// otelhttp must be the outermost handler: it injects the active span
	// into the request context before calling the next handler, and
	// loggingMiddleware needs that span already present (via r.Context())
	// to attach trace_id/span_id to its log line - all three observability
	// pillars correlated by one ID, end to end. otelhttp also records the
	// http.server.request.duration metric (via the global
	// TracerProvider/MeterProvider installed by internal/telemetry) for
	// every request.
	return otelhttp.NewHandler(s.loggingMiddleware(mux), "http.server")
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		logging.FromContext(r.Context(), s.logger).Info("http request",
			logging.String("method", r.Method),
			logging.String("path", r.URL.Path),
			zap.Int("status", rec.status),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
