package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// fakeGitHub stands in for github.com/login/oauth/* and api.github.com so
// tests never make a real network call.
func fakeGitHub(t *testing.T, login string, teams []struct{ Slug, Org string }) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("code") == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "fake-token"})
	})
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fake-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"login": login})
	})
	mux.HandleFunc("/user/teams", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fake-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		type teamResp struct {
			Slug         string `json:"slug"`
			Organization struct {
				Login string `json:"login"`
			} `json:"organization"`
		}
		resp := make([]teamResp, 0, len(teams))
		for _, tm := range teams {
			var tr teamResp
			tr.Slug = tm.Slug
			tr.Organization.Login = tm.Org
			resp = append(resp, tr)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newTestService(t *testing.T, gh *httptest.Server) *Service {
	t.Helper()
	return NewService(Config{
		ClientID:      "client-id",
		ClientSecret:  "client-secret",
		SessionSecret: []byte("session-secret"),
		GitHubOrg:     "pavestack",
		ApproverTeam:  "platform",
		BaseURL:       "http://pave-api.local",
		PortalURL:     "http://portal.local/",
		AuthorizeURL:  gh.URL + "/login/oauth/authorize",
		TokenURL:      gh.URL + "/login/oauth/access_token",
		APIBaseURL:    gh.URL,
	})
}

func TestHandleLoginRedirectsWithStateCookie(t *testing.T) {
	gh := fakeGitHub(t, "octocat", nil)
	s := newTestService(t, gh)

	req := httptest.NewRequest(http.MethodGet, "/auth/github/login", nil)
	rec := httptest.NewRecorder()
	s.HandleLogin(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
	loc, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if loc.Query().Get("client_id") != "client-id" {
		t.Errorf("expected client_id in authorize URL, got %q", loc.Query().Get("client_id"))
	}
	if loc.Query().Get("state") == "" {
		t.Error("expected non-empty state in authorize URL")
	}

	var stateCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == stateCookieName {
			stateCookie = c
		}
	}
	if stateCookie == nil || stateCookie.Value == "" {
		t.Fatal("expected state cookie to be set")
	}
	if stateCookie.Value != loc.Query().Get("state") {
		t.Error("expected state cookie to match the state param sent to GitHub")
	}
}

func TestFullLoginCallbackFlowIssuesSessionAndTeams(t *testing.T) {
	gh := fakeGitHub(t, "octocat", []struct{ Slug, Org string }{
		{Slug: "platform", Org: "pavestack"},
		{Slug: "payments", Org: "pavestack"},
		{Slug: "someteam", Org: "a-different-org"},
	})
	s := newTestService(t, gh)

	loginReq := httptest.NewRequest(http.MethodGet, "/auth/github/login", nil)
	loginRec := httptest.NewRecorder()
	s.HandleLogin(loginRec, loginReq)
	loc, _ := url.Parse(loginRec.Header().Get("Location"))
	state := loc.Query().Get("state")

	cbReq := httptest.NewRequest(http.MethodGet, "/auth/github/callback?code=abc123&state="+state, nil)
	for _, c := range loginRec.Result().Cookies() {
		cbReq.AddCookie(c)
	}
	cbRec := httptest.NewRecorder()
	s.HandleCallback(cbRec, cbReq)

	if cbRec.Code != http.StatusFound {
		t.Fatalf("expected 302 after successful callback, got %d: %s", cbRec.Code, cbRec.Body.String())
	}
	if got := cbRec.Header().Get("Location"); got != "http://portal.local/" {
		t.Errorf("expected redirect to portal URL, got %q", got)
	}

	var sessionCookie *http.Cookie
	for _, c := range cbRec.Result().Cookies() {
		if c.Name == sessionCookieName {
			sessionCookie = c
		}
	}
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatal("expected session cookie to be set")
	}

	identity, err := verifySession(s.cfg.SessionSecret, sessionCookie.Value)
	if err != nil {
		t.Fatalf("verifySession: %v", err)
	}
	if identity.Login != "octocat" {
		t.Errorf("expected login octocat, got %q", identity.Login)
	}
	if !identity.HasTeam("platform") || !identity.HasTeam("payments") {
		t.Errorf("expected platform+payments teams from pavestack org, got %v", identity.Teams)
	}
	if identity.HasTeam("someteam") {
		t.Error("expected team from a different org to be filtered out")
	}
}

func TestHandleCallbackRejectsMismatchedState(t *testing.T) {
	gh := fakeGitHub(t, "octocat", nil)
	s := newTestService(t, gh)

	req := httptest.NewRequest(http.MethodGet, "/auth/github/callback?code=abc123&state=attacker-supplied", nil)
	req.AddCookie(&http.Cookie{Name: stateCookieName, Value: "real-state"})
	rec := httptest.NewRecorder()
	s.HandleCallback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for mismatched state, got %d", rec.Code)
	}
}

func TestHandleCallbackRejectsMissingStateCookie(t *testing.T) {
	gh := fakeGitHub(t, "octocat", nil)
	s := newTestService(t, gh)

	req := httptest.NewRequest(http.MethodGet, "/auth/github/callback?code=abc123&state=whatever", nil)
	rec := httptest.NewRecorder()
	s.HandleCallback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 with no state cookie at all, got %d", rec.Code)
	}
}

func TestRequireAuthRejectsMissingSession(t *testing.T) {
	gh := fakeGitHub(t, "octocat", nil)
	s := newTestService(t, gh)

	called := false
	h := s.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if called {
		t.Error("expected inner handler not to run")
	}
}

func TestRequireAuthAllowsValidSession(t *testing.T) {
	gh := fakeGitHub(t, "octocat", nil)
	s := newTestService(t, gh)

	var gotIdentity Identity
	h := s.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIdentity, _ = IdentityFromContext(r.Context())
	}))

	cookieValue, err := signSession(s.cfg.SessionSecret, sessionPayload{Login: "octocat", ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: cookieValue})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotIdentity.Login != "octocat" {
		t.Errorf("expected identity to be attached to context, got %+v", gotIdentity)
	}
}

func TestRequireTeamRejectsNonMember(t *testing.T) {
	gh := fakeGitHub(t, "octocat", nil)
	s := newTestService(t, gh)

	h := s.RequireTeam("platform")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("expected inner handler not to run for a non-member")
	}))

	cookieValue, _ := signSession(s.cfg.SessionSecret, sessionPayload{
		Login: "mallory", Teams: []string{"some-other-team"}, ExpiresAt: time.Now().Add(time.Hour),
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/access-requests/ar_1", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: cookieValue})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-member, got %d", rec.Code)
	}
}

func TestRequireTeamAllowsMember(t *testing.T) {
	gh := fakeGitHub(t, "octocat", nil)
	s := newTestService(t, gh)

	var gotIdentity Identity
	h := s.RequireTeam("platform")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIdentity, _ = IdentityFromContext(r.Context())
	}))

	cookieValue, _ := signSession(s.cfg.SessionSecret, sessionPayload{
		Login: "platform-lead", Teams: []string{"platform"}, ExpiresAt: time.Now().Add(time.Hour),
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/access-requests/ar_1", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: cookieValue})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotIdentity.Login != "platform-lead" {
		t.Errorf("expected identity to be attached to context, got %+v", gotIdentity)
	}
}

func TestHandleLogoutClearsSessionCookie(t *testing.T) {
	gh := fakeGitHub(t, "octocat", nil)
	s := newTestService(t, gh)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	s.HandleLogout(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected redirect to portal on logout, got %d", rec.Code)
	}
	var cleared *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookieName {
			cleared = c
		}
	}
	if cleared == nil || cleared.MaxAge >= 0 {
		t.Fatal("expected session cookie to be cleared (MaxAge < 0)")
	}
}
