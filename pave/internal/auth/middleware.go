package auth

import (
	"context"
	"fmt"
	"net/http"
)

// RequireAuth rejects requests without a valid session cookie (401) and
// otherwise attaches the verified Identity to the request context.
func (s *Service) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := s.identityFromRequest(r)
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), identityKey, identity)))
	})
}

// RequireTeam rejects requests without a valid session cookie (401) or
// without membership in the given GitHub team (403). It only checks
// group membership within the session's cached team list captured at
// login time - a team change on GitHub takes effect on the caller's next
// login, not instantly, which is an acceptable tradeoff for a session
// with a bounded (sessionTTL) lifetime.
func (s *Service) RequireTeam(team string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := s.identityFromRequest(r)
			if !ok {
				writeJSONError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			if !identity.HasTeam(team) {
				writeJSONError(w, http.StatusForbidden, fmt.Sprintf("requires membership in the %q team", team))
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), identityKey, identity)))
		})
	}
}

func (s *Service) identityFromRequest(r *http.Request) (Identity, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return Identity{}, false
	}
	identity, err := verifySession(s.cfg.SessionSecret, cookie.Value)
	if err != nil {
		return Identity{}, false
	}
	return identity, true
}
