package apiserver_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pavestack/pave/internal/apiserver"
	"go.uber.org/zap"
)

func TestRequestIDIsEchoedFromCaller(t *testing.T) {
	root := setupRepo(t)
	srv, err := apiserver.New(apiserver.Config{RepoRoot: root, DryRun: true}, zap.NewNop())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-Id", "caller-supplied-id")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-Id"); got != "caller-supplied-id" {
		t.Errorf("expected caller-supplied request ID to be echoed back, got %q", got)
	}
}

func TestRequestIDIsGeneratedWhenAbsent(t *testing.T) {
	root := setupRepo(t)
	srv, err := apiserver.New(apiserver.Config{RepoRoot: root, DryRun: true}, zap.NewNop())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-Id"); got == "" {
		t.Error("expected a generated X-Request-Id header")
	}
}

func TestRequestBodyOverLimitIsRejected(t *testing.T) {
	root := setupRepo(t)
	srv, err := apiserver.New(apiserver.Config{RepoRoot: root, DryRun: true}, zap.NewNop())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	huge := strings.Repeat("a", (1<<20)+1)
	body := `{"requester":"` + huge + `","team":"t","namespace":"n","level":"read"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/access-requests", strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for oversized body, got %d: %s", rec.Code, rec.Body.String())
	}
}
