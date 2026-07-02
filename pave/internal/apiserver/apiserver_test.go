package apiserver_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pavestack/pave/internal/apiserver"
	"go.uber.org/zap"
)

// setupRepo builds a minimal on-disk repo (real files, not afero) since
// apiserver.New shells out through the real scaffold/gitops packages which
// use os.* directly for the gitops half.
func setupRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	dirs := []string{
		filepath.Join(root, "pave", "schemas"),
		filepath.Join(root, "platform-config"),
		filepath.Join(root, "service-template-api", "cmd", "server"),
		filepath.Join(root, "service-template-api", "deploy", "helm", "service-template-api", "templates"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	schema := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "additionalProperties": false,
  "required": ["name", "team", "database"],
  "properties": {
    "name": {"type": "string", "pattern": "^[a-z][a-z0-9-]{1,48}[a-z0-9]$"},
    "team": {"type": "string", "pattern": "^[a-z][a-z0-9-]{1,62}[a-z0-9]$"},
    "database": {"type": "boolean"},
    "runtime": {"type": "string", "enum": ["go"]},
    "exposure": {"type": "string", "enum": ["internal", "public"]},
    "tier": {"type": "string", "enum": ["tier-1", "tier-2", "tier-3"]}
  }
}`
	files := map[string]string{
		filepath.Join(root, "pave", "schemas", "service-request.schema.json"):   schema,
		filepath.Join(root, "service-template-api", "go.mod"):                   "module github.com/pavestack/service-template-api\n\ngo 1.23\n",
		filepath.Join(root, "service-template-api", "README.md"):                "# service-template-api\n",
		filepath.Join(root, "service-template-api", "cmd", "server", "main.go"): "package main\n\nfunc main() {}\n",
		filepath.Join(root, "service-template-api", "catalog-info.yaml"): `apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: service-template-api
  description: Golden-path template
  annotations:
    github.com/project-slug: pavestack/pavestack
    pavestack.io/team: platform
spec:
  type: service
  lifecycle: production
  owner: team-platform
  system: pavestack
`,
		filepath.Join(root, "service-template-api", "scorecard.yaml"): `service: service-template-api
owner: team-platform
criteria:
  docs_present:
    weight: 20
    status: passing
    evidence: README.md
overall_score: 100
`,
		filepath.Join(root, "service-template-api", "deploy", "helm", "service-template-api", "Chart.yaml"):  "apiVersion: v2\nname: service-template-api\nversion: 0.1.0\n",
		filepath.Join(root, "service-template-api", "deploy", "helm", "service-template-api", "values.yaml"): "replicaCount: 2\nimage:\n  repository: example/service-template-api\n",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return root
}

func newTestServer(t *testing.T) (*apiserver.Server, string) {
	t.Helper()
	root := setupRepo(t)
	srv, err := apiserver.New(apiserver.Config{RepoRoot: root, DryRun: true}, zap.NewNop(), nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return srv, root
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealthz(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := doJSON(t, srv.Handler(), http.MethodGet, "/healthz", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestOpenAPISpecIsServed(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := doJSON(t, srv.Handler(), http.MethodGet, "/api/v1/openapi.json", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("expected valid JSON: %v", err)
	}
	if doc["openapi"] == nil {
		t.Error("expected an 'openapi' version field in the served spec")
	}
}

func TestListServicesFindsTemplate(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := doJSON(t, srv.Handler(), http.MethodGet, "/api/v1/services", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var services []apiserver.Service
	if err := json.Unmarshal(rec.Body.Bytes(), &services); err != nil {
		t.Fatal(err)
	}
	if len(services) != 1 || services[0].Name != "service-template-api" {
		t.Fatalf("expected exactly service-template-api, got %+v", services)
	}
	if services[0].CreatedVia != "manual" {
		t.Errorf("expected template (no created-via annotation) to read as manual, got %q", services[0].CreatedVia)
	}
	if services[0].Scorecard.OverallScore != 100 {
		t.Errorf("expected overall score 100, got %d", services[0].Scorecard.OverallScore)
	}
}

func TestCostEstimateVariesByTierAndExtras(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := doJSON(t, srv.Handler(), http.MethodGet, "/api/v1/cost-estimate?tier=tier-1&exposure=public&database=true", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp apiserver.CostEstimateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Breakdown) != 3 {
		t.Fatalf("expected 3 breakdown line items for tier-1+public+database, got %d", len(resp.Breakdown))
	}
	if resp.MonthlyUSDLow <= 0 || resp.MonthlyUSDHigh <= resp.MonthlyUSDLow {
		t.Errorf("unexpected cost range: %+v", resp)
	}
}

func TestCreateServiceJobLifecycleDryRun(t *testing.T) {
	srv, root := newTestServer(t)
	h := srv.Handler()

	createBody := map[string]any{
		"name":     "payments",
		"team":     "team-payments",
		"runtime":  "go",
		"exposure": "internal",
		"database": false,
		"tier":     "tier-2",
	}
	rec := doJSON(t, h, http.MethodPost, "/api/v1/services", createBody)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}

	var accepted struct {
		JobID     string `json:"jobId"`
		StatusURL string `json:"statusUrl"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &accepted); err != nil {
		t.Fatal(err)
	}
	if accepted.JobID == "" {
		t.Fatal("expected non-empty jobId")
	}

	var job apiserver.Job
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rec := doJSON(t, h, http.MethodGet, "/api/v1/jobs/"+accepted.JobID, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 polling job, got %d", rec.Code)
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &job); err != nil {
			t.Fatal(err)
		}
		if job.Status == apiserver.JobCompleted || job.Status == apiserver.JobFailed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if job.Status != apiserver.JobCompleted {
		t.Fatalf("expected job to complete, got status=%s error=%s steps=%+v", job.Status, job.Error, job.Steps)
	}
	if !job.DryRun {
		t.Error("expected DryRun=true")
	}

	if _, err := os.Stat(filepath.Join(root, "services", "payments-api")); err != nil {
		t.Errorf("expected scaffolded service dir to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "platform-config", "tenants", "payments", "tenant.yaml")); err != nil {
		t.Errorf("expected tenant.yaml to exist: %v", err)
	}
}

func TestCreateServiceRejectsInvalidName(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := doJSON(t, srv.Handler(), http.MethodPost, "/api/v1/services", map[string]any{
		"name": "Invalid Name!", "team": "team-payments", "database": false,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid name, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAccessRequestApprovalWorkflow(t *testing.T) {
	srv, _ := newTestServer(t)
	h := srv.Handler()

	rec := doJSON(t, h, http.MethodPost, "/api/v1/access-requests", map[string]any{
		"requester": "alice", "team": "team-payments", "namespace": "payments", "level": "write", "reason": "on-call rotation",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var created apiserver.AccessRequest
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Status != "pending" {
		t.Fatalf("expected new access request to be pending (no silent auto-grant), got %q", created.Status)
	}

	rec = doJSON(t, h, http.MethodPatch, "/api/v1/access-requests/"+created.ID, map[string]any{
		"action": "approve", "approver": "platform-lead",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var decided apiserver.AccessRequest
	if err := json.Unmarshal(rec.Body.Bytes(), &decided); err != nil {
		t.Fatal(err)
	}
	if decided.Status != "approved" || decided.Approver != "platform-lead" {
		t.Fatalf("expected approved by platform-lead, got %+v", decided)
	}

	rec = doJSON(t, h, http.MethodPatch, "/api/v1/access-requests/"+created.ID, map[string]any{
		"action": "approve",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when approver missing, got %d", rec.Code)
	}
}
