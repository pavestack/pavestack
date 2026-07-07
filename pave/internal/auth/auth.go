// Package auth authenticates pave-api callers against GitHub, since this
// platform is already an all-GitHub shop (owners are team slugs, `pave`
// itself shells out to `gh` for PRs) - see docs/adr for the "why GitHub as
// the identity provider, not a new IdP/DB" decision. It issues its own
// signed session cookie after a standard OAuth 2.0 Authorization Code
// exchange with GitHub, and checks GitHub org/team membership for
// authorization (e.g. who may approve an access request).
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	sessionCookieName = "pave_session"
	stateCookieName   = "pave_oauth_state"
	sessionTTL        = 12 * time.Hour
	stateTTL          = 10 * time.Minute
)

// Config controls OAuth client credentials, cookie behavior, and which
// GitHub org/team gates access-request approval.
type Config struct {
	ClientID      string
	ClientSecret  string
	SessionSecret []byte

	// GitHubOrg is the org whose team membership is checked; ApproverTeam
	// is the team slug within it required to approve/deny access requests.
	GitHubOrg    string
	ApproverTeam string

	// BaseURL is pave-api's own externally reachable origin, used to build
	// the OAuth redirect_uri (e.g. http://localhost:8787). PortalURL is
	// where the browser is sent after a successful login/logout.
	BaseURL   string
	PortalURL string

	// CookieSecure sets the Secure flag on cookies. Must be true for any
	// non-localhost deployment; false is only for http://localhost dev.
	CookieSecure bool

	// Overridable for tests; default to real github.com endpoints in
	// NewService when left empty.
	AuthorizeURL string
	TokenURL     string
	APIBaseURL   string
}

type ctxKey int

const identityKey ctxKey = iota

// IdentityFromContext returns the identity RequireAuth/RequireTeam
// attached to the request context, if any.
func IdentityFromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey).(Identity)
	return id, ok
}

// Service implements the GitHub OAuth login flow and the
// RequireAuth/RequireTeam middleware that consumes its session cookie.
type Service struct {
	cfg    Config
	client *http.Client
}

func NewService(cfg Config) *Service {
	if cfg.AuthorizeURL == "" {
		cfg.AuthorizeURL = "https://github.com/login/oauth/authorize"
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = "https://github.com/login/oauth/access_token"
	}
	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = "https://api.github.com"
	}
	return &Service{cfg: cfg, client: &http.Client{Timeout: 10 * time.Second}}
}

// RegisterRoutes mounts the login/callback/logout/me endpoints on mux.
func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/github/login", s.HandleLogin)
	mux.HandleFunc("GET /auth/github/callback", s.HandleCallback)
	mux.HandleFunc("POST /auth/logout", s.HandleLogout)
	mux.Handle("GET /auth/me", s.RequireAuth(http.HandlerFunc(s.HandleMe)))
}

func (s *Service) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := randomToken()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/auth/github",
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(stateTTL),
	})

	u, err := url.Parse(s.cfg.AuthorizeURL)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	q := u.Query()
	q.Set("client_id", s.cfg.ClientID)
	q.Set("redirect_uri", s.redirectURI())
	q.Set("scope", "read:org")
	q.Set("state", state)
	u.RawQuery = q.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)
}

func (s *Service) HandleCallback(w http.ResponseWriter, r *http.Request) {
	// Always clear the one-time state cookie, whether or not it validates.
	defer http.SetCookie(w, &http.Cookie{
		Name: stateCookieName, Value: "", Path: "/auth/github", MaxAge: -1,
	})

	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil || stateCookie.Value == "" ||
		!hmac.Equal([]byte(stateCookie.Value), []byte(r.URL.Query().Get("state"))) {
		writeJSONError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		writeJSONError(w, http.StatusBadRequest, "missing code")
		return
	}

	token, err := s.exchangeCode(r.Context(), code)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "github token exchange failed")
		return
	}

	login, err := s.fetchLogin(r.Context(), token)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "github user lookup failed")
		return
	}

	teams, err := s.fetchOrgTeams(r.Context(), token)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "github team lookup failed")
		return
	}

	cookieValue, err := signSession(s.cfg.SessionSecret, sessionPayload{
		Login: login, Teams: teams, ExpiresAt: time.Now().Add(sessionTTL),
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    cookieValue,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(sessionTTL),
	})

	http.Redirect(w, r, s.portalRedirectTarget(), http.StatusFound)
}

func (s *Service) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", MaxAge: -1})
	if s.cfg.PortalURL == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, s.cfg.PortalURL, http.StatusFound)
}

func (s *Service) HandleMe(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	writeJSON(w, http.StatusOK, identity)
}

func (s *Service) redirectURI() string {
	return strings.TrimRight(s.cfg.BaseURL, "/") + "/auth/github/callback"
}

func (s *Service) portalRedirectTarget() string {
	if s.cfg.PortalURL == "" {
		return "/"
	}
	return s.cfg.PortalURL
}

func (s *Service) exchangeCode(ctx context.Context, code string) (string, error) {
	form := url.Values{
		"client_id":     {s.cfg.ClientID},
		"client_secret": {s.cfg.ClientSecret},
		"code":          {code},
		"redirect_uri":  {s.redirectURI()},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %d", resp.StatusCode)
	}

	var body struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return "", err
	}
	if body.Error != "" || body.AccessToken == "" {
		return "", fmt.Errorf("github oauth error: %s", body.Error)
	}
	return body.AccessToken, nil
}

func (s *Service) fetchLogin(ctx context.Context, token string) (string, error) {
	var user struct {
		Login string `json:"login"`
	}
	if err := s.getJSON(ctx, "/user", token, &user); err != nil {
		return "", err
	}
	if user.Login == "" {
		return "", errors.New("github user response missing login")
	}
	return user.Login, nil
}

// fetchOrgTeams returns the caller's team slugs within cfg.GitHubOrg.
// Reads a single page of up to 100 teams across all of the user's orgs -
// an internal-platform user realistically belongs to far fewer than that,
// so full pagination isn't worth the complexity here.
func (s *Service) fetchOrgTeams(ctx context.Context, token string) ([]string, error) {
	var teams []struct {
		Slug         string `json:"slug"`
		Organization struct {
			Login string `json:"login"`
		} `json:"organization"`
	}
	if err := s.getJSON(ctx, "/user/teams?per_page=100", token, &teams); err != nil {
		return nil, err
	}
	var slugs []string
	for _, t := range teams {
		if strings.EqualFold(t.Organization.Login, s.cfg.GitHubOrg) {
			slugs = append(slugs, t.Slug)
		}
	}
	return slugs, nil
}

func (s *Service) getJSON(ctx context.Context, path, token string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cfg.APIBaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned %d", path, resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out)
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
