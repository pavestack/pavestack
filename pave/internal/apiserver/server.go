package apiserver

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pavestack/pave/internal/cost"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

// Config controls how the server resolves state and whether it actually
// shells out to git/gh, or simulates that final step.
type Config struct {
	RepoRoot   string
	DryRun     bool
	CORSOrigin string // "*" or a specific origin
	RuntimeDir string // where access-requests.json lives; default <RepoRoot>/.pave-api
}

type Server struct {
	cfg      Config
	logger   *zap.Logger
	jobs     *JobStore
	requests *AccessRequestStore
	mux      *http.ServeMux
}

// New builds a Server. logger may be nil (a no-op logger is used instead),
// which keeps existing callers like tests that don't care about log output
// working without change.
func New(cfg Config, logger *zap.Logger) (*Server, error) {
	if cfg.RuntimeDir == "" {
		cfg.RuntimeDir = filepath.Join(cfg.RepoRoot, ".pave-api")
	}
	if cfg.CORSOrigin == "" {
		// Never default to "*": this API is called with credentials
		// (session cookies) from the portal, and credentialed CORS
		// requests reject a wildcard origin anyway.
		cfg.CORSOrigin = "http://localhost:5173"
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	requests, err := NewAccessRequestStore(cfg.RuntimeDir)
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg:      cfg,
		logger:   logger,
		jobs:     NewJobStore(cfg.RepoRoot, cfg.DryRun),
		requests: requests,
		mux:      http.NewServeMux(),
	}
	s.routes()
	return s, nil
}

// GitOpsToolsAvailable reports whether git and gh are on PATH, which the
// caller can log at startup - if they're missing, every create-service job
// will fall back to dry-run for its open_pr step regardless of DryRun.
func GitOpsToolsAvailable() bool {
	_, gitErr := exec.LookPath("git")
	_, ghErr := exec.LookPath("gh")
	return gitErr == nil && ghErr == nil
}

// maxRequestBodyBytes bounds JSON request bodies so a large/malicious body
// can't exhaust memory before json.Decode ever gets a chance to reject it.
const maxRequestBodyBytes = 1 << 20 // 1 MiB

// Handler wraps the routed mux with the middleware chain, outermost first:
// otelhttp must be outermost so it injects the active span into the
// request context before requestID/logging run (mirrors
// service-template-api/internal/server - see its Handler() comment for
// why), then request-ID assignment, then request logging, then panic
// recovery (closest to the actual handlers so it catches anything that
// panics in CORS/routing/business logic and still gets logged by the
// logging middleware above it), then CORS, then the routed mux.
func (s *Server) Handler() http.Handler {
	h := s.withCORS(s.mux)
	h = s.recoveryMiddleware(h)
	h = s.loggingMiddleware(h)
	h = s.requestIDMiddleware(h)
	return otelhttp.NewHandler(h, "http.server")
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /api/v1/services", s.handleListServices)
	s.mux.HandleFunc("GET /api/v1/services/{name}", s.handleGetService)
	s.mux.HandleFunc("POST /api/v1/services", s.handleCreateService)
	s.mux.HandleFunc("GET /api/v1/jobs/{id}", s.handleGetJob)
	s.mux.HandleFunc("GET /api/v1/cost-estimate", s.handleCostEstimate)
	s.mux.HandleFunc("GET /api/v1/access-requests", s.handleListAccessRequests)
	s.mux.HandleFunc("POST /api/v1/access-requests", s.handleCreateAccessRequest)
	s.mux.HandleFunc("PATCH /api/v1/access-requests/{id}", s.handleDecideAccessRequest)
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", s.cfg.CORSOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleListServices(w http.ResponseWriter, r *http.Request) {
	services, err := LoadCatalog(s.cfg.RepoRoot)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, services)
}

func (s *Server) handleGetService(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	svc, ok := LoadOne(s.cfg.RepoRoot, name)
	if !ok {
		writeError(w, http.StatusNotFound, errServiceNotFound(name))
		return
	}
	writeJSON(w, http.StatusOK, svc)
}

func (s *Server) handleCreateService(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	job, err := s.jobs.Submit(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"jobId":     job.JobID,
		"statusUrl": "/api/v1/jobs/" + job.JobID,
	})
}

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, ok := s.jobs.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, errJobNotFound(id))
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) handleCostEstimate(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	tier := q.Get("tier")
	exposure := q.Get("exposure")
	if exposure == "" {
		exposure = "internal"
	}
	database := strings.EqualFold(q.Get("database"), "true")

	est := cost.Estimate(cost.ResolveTier(tier), exposure, database)
	resp := CostEstimateResponse{
		MonthlyUSDLow:  est.MonthlyUSDLow,
		MonthlyUSDHigh: est.MonthlyUSDHigh,
		Currency:       est.Currency,
		Disclaimer:     est.Disclaimer,
	}
	for _, item := range est.Breakdown {
		resp.Breakdown = append(resp.Breakdown, CostLineItem{
			Item:          item.Item,
			MonthlyUSDLow: item.MonthlyUSDLow,
			MonthlyUSD:    item.MonthlyUSD,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleListAccessRequests(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.requests.List())
}

func (s *Server) handleCreateAccessRequest(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	var req AccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Requester == "" || req.Namespace == "" || req.Level == "" {
		writeError(w, http.StatusBadRequest, errMissingFields)
		return
	}

	created, err := s.requests.Create(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleDecideAccessRequest(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	id := r.PathValue("id")
	var body struct {
		Action   string `json:"action"`
		Approver string `json:"approver"`
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if body.Approver == "" {
		writeError(w, http.StatusBadRequest, errApproverRequired)
		return
	}

	updated, err := s.requests.Decide(id, body.Action, body.Approver, body.Note)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Log is a tiny helper so main.go doesn't need its own log import just to
// announce the listen address.
func Log(format string, args ...any) {
	log.Printf(format, args...)
}
