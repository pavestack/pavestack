package apiserver

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pavestack/pave/internal/auth"
	"github.com/pavestack/pave/internal/cost"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

// Config controls how the server resolves state and whether it actually
// shells out to git/gh, or simulates that final step.
type Config struct {
	RepoRoot   string
	DryRun     bool
	CORSOrigin string // a specific origin - never "*", see New()
	RuntimeDir string // where access-requests.json lives; default <RepoRoot>/.pave-api

	// ApproverTeam is the GitHub team (within the auth Service's configured
	// org) required to approve/deny access requests when auth is enabled.
	// Ignored when authSvc passed to New is nil.
	ApproverTeam string
}

type Server struct {
	cfg      Config
	logger   *zap.Logger
	auth     *auth.Service
	jobs     *JobStore
	requests *AccessRequestStore
	mux      *http.ServeMux

	mutationLimiter *ipRateLimiter
	authLimiter     *ipRateLimiter
}

// New builds a Server. logger may be nil (a no-op logger is used instead).
// authSvc may also be nil, which runs pave-api with no authentication at
// all on the mutating endpoints - only ever appropriate for local dev/CI,
// wired through PAVE_API_DISABLE_AUTH in internal/config, never silently.
func New(cfg Config, logger *zap.Logger, authSvc *auth.Service) (*Server, error) {
	if cfg.RuntimeDir == "" {
		cfg.RuntimeDir = filepath.Join(cfg.RepoRoot, ".pave-api")
	}
	if cfg.CORSOrigin == "" || cfg.CORSOrigin == "*" {
		// Never allow "*": this API is called with credentials (session
		// cookies) from the portal, and credentialed CORS requests reject
		// a wildcard origin anyway - so "*" would just break auth, not
		// relax it.
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
		auth:     authSvc,
		jobs:     NewJobStore(cfg.RepoRoot, cfg.DryRun),
		requests: requests,
		mux:      http.NewServeMux(),

		// Mutating endpoints: generous enough for normal portal use, tight
		// enough to blunt a scripted hammering of create-service/access-request.
		mutationLimiter: newIPRateLimiter(2, 10),
		// Login/callback: much tighter, since these are also the surface
		// an attacker would hit to brute-force or abuse the OAuth dance.
		authLimiter: newIPRateLimiter(1, 5),
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
// logging middleware above it), then security headers, then CORS, then
// the routed mux. Auth/team checks and rate limiting are applied
// per-route in routes() instead of globally, since GET endpoints stay
// public while mutating endpoints don't, and login/callback need a much
// tighter rate-limit budget than the rest.
func (s *Server) Handler() http.Handler {
	h := s.withCORS(s.mux)
	h = s.securityHeadersMiddleware(h)
	h = s.recoveryMiddleware(h)
	h = s.loggingMiddleware(h)
	h = s.requestIDMiddleware(h)
	return otelhttp.NewHandler(h, "http.server")
}

func (s *Server) routes() {
	// Public, read-only: catalog/cost data already lives in
	// catalog-info.yaml/scorecard.yaml committed to this repo, so gating
	// reads behind auth wouldn't add real confidentiality - see
	// AGENTS.md's "Portal data model" section.
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /api/v1/services", s.handleListServices)
	s.mux.HandleFunc("GET /api/v1/services/{name}", s.handleGetService)
	s.mux.HandleFunc("GET /api/v1/jobs/{id}", s.handleGetJob)
	s.mux.HandleFunc("GET /api/v1/cost-estimate", s.handleCostEstimate)
	s.mux.HandleFunc("GET /api/v1/access-requests", s.handleListAccessRequests)

	// Mutating: require a verified session when auth is configured (see
	// protect/protectApprover), and are always rate-limited regardless.
	s.mux.Handle("POST /api/v1/services",
		s.rateLimit(s.mutationLimiter, s.protect(http.HandlerFunc(s.handleCreateService))))
	s.mux.Handle("POST /api/v1/access-requests",
		s.rateLimit(s.mutationLimiter, s.protect(http.HandlerFunc(s.handleCreateAccessRequest))))
	s.mux.Handle("PATCH /api/v1/access-requests/{id}",
		s.rateLimit(s.mutationLimiter, s.protectApprover(http.HandlerFunc(s.handleDecideAccessRequest))))

	if s.auth != nil {
		s.mux.Handle("GET /auth/github/login", s.rateLimit(s.authLimiter, http.HandlerFunc(s.auth.HandleLogin)))
		s.mux.Handle("GET /auth/github/callback", s.rateLimit(s.authLimiter, http.HandlerFunc(s.auth.HandleCallback)))
		s.mux.HandleFunc("POST /auth/logout", s.auth.HandleLogout)
		s.mux.Handle("GET /auth/me", s.auth.RequireAuth(http.HandlerFunc(s.auth.HandleMe)))
	}
}

// protect requires a verified session when auth is configured; with no
// auth Service (local dev/CI via PAVE_API_DISABLE_AUTH) it's a no-op,
// preserving this package's previous unauthenticated behavior exactly.
func (s *Server) protect(next http.Handler) http.Handler {
	if s.auth == nil {
		return next
	}
	return s.auth.RequireAuth(next)
}

// protectApprover requires membership in cfg.ApproverTeam when auth is
// configured; see protect for the no-auth fallback.
func (s *Server) protectApprover(next http.Handler) http.Handler {
	if s.auth == nil {
		return next
	}
	return s.auth.RequireTeam(s.cfg.ApproverTeam)(next)
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", s.cfg.CORSOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
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
		Action string `json:"action"`
		// Approver is only used as a fallback when auth is disabled
		// (PAVE_API_DISABLE_AUTH=true, local dev/CI only) - see below.
		Approver string `json:"approver"`
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	// The verified session identity is always the approver of record when
	// auth is enabled - never the client-supplied field, which would let
	// any caller self-approve by naming themselves as approver. This is
	// what protectApprover's RequireTeam check is actually gating.
	approver := body.Approver
	if identity, ok := auth.IdentityFromContext(r.Context()); ok {
		approver = identity.Login
	}
	if approver == "" {
		writeError(w, http.StatusBadRequest, errApproverRequired)
		return
	}

	updated, err := s.requests.Decide(id, body.Action, approver, body.Note)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
