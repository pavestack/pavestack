package apiserver_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/pavestack/pave/internal/apiserver"
	"github.com/pavestack/pave/internal/auth"
	"go.uber.org/zap"
)

func jsonBody(t *testing.T, v any) io.Reader {
	t.Helper()
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		t.Fatal(err)
	}
	return &buf
}

// fakeGitHub stands in for github.com/login/oauth/* and api.github.com,
// exactly like internal/auth's own tests, so the full login -> callback ->
// authenticated-request flow can be exercised end-to-end through the real
// apiserver.Server without ever making a real network call.
func fakeGitHub(t *testing.T, login string, teams []string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "fake-token"})
	})
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"login": login})
	})
	mux.HandleFunc("/user/teams", func(w http.ResponseWriter, r *http.Request) {
		type teamResp struct {
			Slug         string `json:"slug"`
			Organization struct {
				Login string `json:"login"`
			} `json:"organization"`
		}
		resp := make([]teamResp, 0, len(teams))
		for _, slug := range teams {
			var tr teamResp
			tr.Slug = slug
			tr.Organization.Login = "pavestack"
			resp = append(resp, tr)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// loginAs drives the full OAuth login+callback flow against h and returns
// the resulting session cookie, ready to attach to subsequent requests.
func loginAs(t *testing.T, h http.Handler) *http.Cookie {
	t.Helper()

	loginReq := httptest.NewRequest(http.MethodGet, "/auth/github/login", nil)
	loginRec := httptest.NewRecorder()
	h.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusFound {
		t.Fatalf("login: expected 302, got %d", loginRec.Code)
	}
	loc, err := url.Parse(loginRec.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	state := loc.Query().Get("state")

	cbReq := httptest.NewRequest(http.MethodGet, "/auth/github/callback?code=abc123&state="+state, nil)
	for _, c := range loginRec.Result().Cookies() {
		cbReq.AddCookie(c)
	}
	cbRec := httptest.NewRecorder()
	h.ServeHTTP(cbRec, cbReq)
	if cbRec.Code != http.StatusFound {
		t.Fatalf("callback: expected 302, got %d: %s", cbRec.Code, cbRec.Body.String())
	}

	for _, c := range cbRec.Result().Cookies() {
		if c.Name == "pave_session" {
			return c
		}
	}
	t.Fatal("callback did not set a session cookie")
	return nil
}

func newAuthedTestServer(t *testing.T, login string, teams []string) (http.Handler, string) {
	t.Helper()
	root := setupRepo(t)
	gh := fakeGitHub(t, login, teams)

	authSvc := auth.NewService(auth.Config{
		ClientID:      "client-id",
		ClientSecret:  "client-secret",
		SessionSecret: []byte("integration-test-secret"),
		GitHubOrg:     "pavestack",
		ApproverTeam:  "platform",
		BaseURL:       "http://pave-api.local",
		PortalURL:     "http://portal.local/",
		AuthorizeURL:  gh.URL + "/login/oauth/authorize",
		TokenURL:      gh.URL + "/login/oauth/access_token",
		APIBaseURL:    gh.URL,
	})

	srv, err := apiserver.New(apiserver.Config{RepoRoot: root, DryRun: true, ApproverTeam: "platform"}, zap.NewNop(), authSvc)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return srv.Handler(), root
}

func TestCreateServiceRequiresAuthWhenConfigured(t *testing.T) {
	h, _ := newAuthedTestServer(t, "octocat", nil)

	rec := doJSON(t, h, http.MethodPost, "/api/v1/services", map[string]any{
		"name": "payments", "team": "team-payments", "database": false,
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated create-service, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateServiceSucceedsWithAnyAuthenticatedSession(t *testing.T) {
	h, _ := newAuthedTestServer(t, "octocat", []string{"some-other-team"})
	session := loginAs(t, h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/services", jsonBody(t, map[string]any{
		"name": "payments", "team": "team-payments", "runtime": "go", "exposure": "internal", "database": false, "tier": "tier-2",
	}))
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for an authenticated caller regardless of team, got %d: %s", rec.Code, rec.Body.String())
	}

	var accepted struct {
		JobID string `json:"jobId"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &accepted); err != nil {
		t.Fatal(err)
	}
	waitForJobDone(t, h, session, accepted.JobID)
}

// waitForJobDone polls until the async scaffold job finishes, so the test's
// t.TempDir() cleanup doesn't race the job's goroutine still writing files.
func waitForJobDone(t *testing.T, h http.Handler, session *http.Cookie, jobID string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/"+jobID, nil)
		if session != nil {
			req.AddCookie(session)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		var job apiserver.Job
		if err := json.Unmarshal(rec.Body.Bytes(), &job); err == nil {
			if job.Status == apiserver.JobCompleted || job.Status == apiserver.JobFailed {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("job %s did not finish before test deadline", jobID)
}

func TestDecideAccessRequestRequiresApproverTeam(t *testing.T) {
	h, _ := newAuthedTestServer(t, "mallory", []string{"not-platform"})
	session := loginAs(t, h)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/access-requests", jsonBody(t, map[string]any{
		"requester": "alice", "team": "team-payments", "namespace": "payments", "level": "write", "reason": "on-call",
	}))
	createReq.AddCookie(session)
	createResp := httptest.NewRecorder()
	h.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 creating access request, got %d: %s", createResp.Code, createResp.Body.String())
	}
	var created apiserver.AccessRequest
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	// mallory is authenticated but not on the "platform" team, so deciding
	// must be forbidden even though she supplies a plausible approver name.
	decideReq := httptest.NewRequest(http.MethodPatch, "/api/v1/access-requests/"+created.ID,
		jsonBody(t, map[string]any{"action": "approve", "approver": "mallory"}))
	decideReq.AddCookie(session)
	decideResp := httptest.NewRecorder()
	h.ServeHTTP(decideResp, decideReq)

	if decideResp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for a non-approver-team session, got %d: %s", decideResp.Code, decideResp.Body.String())
	}
}

func TestDecideAccessRequestUsesSessionIdentityNotBodyApprover(t *testing.T) {
	h, _ := newAuthedTestServer(t, "platform-lead", []string{"platform"})
	session := loginAs(t, h)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/access-requests", jsonBody(t, map[string]any{
		"requester": "alice", "team": "team-payments", "namespace": "payments", "level": "write", "reason": "on-call",
	}))
	createReq.AddCookie(session)
	createResp := httptest.NewRecorder()
	h.ServeHTTP(createResp, createReq)
	var created apiserver.AccessRequest
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	// Body claims a different approver than the session identity - the
	// session identity must win, closing the self-approval gap.
	decideReq := httptest.NewRequest(http.MethodPatch, "/api/v1/access-requests/"+created.ID,
		jsonBody(t, map[string]any{"action": "approve", "approver": "someone-else"}))
	decideReq.AddCookie(session)
	decideResp := httptest.NewRecorder()
	h.ServeHTTP(decideResp, decideReq)

	if decideResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", decideResp.Code, decideResp.Body.String())
	}
	var decided apiserver.AccessRequest
	if err := json.Unmarshal(decideResp.Body.Bytes(), &decided); err != nil {
		t.Fatal(err)
	}
	if decided.Approver != "platform-lead" {
		t.Errorf("expected approver to be the verified session identity 'platform-lead', got %q", decided.Approver)
	}
}
