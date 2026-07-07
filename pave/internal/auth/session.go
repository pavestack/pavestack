package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Identity is who pave-api believes made a request, established via a
// verified GitHub OAuth session. Teams is the subset of the configured
// GitHub org's teams the user belongs to, fetched once at login time.
type Identity struct {
	Login string   `json:"login"`
	Teams []string `json:"teams"`
}

// HasTeam reports whether the identity belongs to the given team slug.
func (i Identity) HasTeam(team string) bool {
	for _, t := range i.Teams {
		if t == team {
			return true
		}
	}
	return false
}

type sessionPayload struct {
	Login     string    `json:"login"`
	Teams     []string  `json:"teams"`
	ExpiresAt time.Time `json:"expires_at"`
}

var errInvalidSession = errors.New("invalid or expired session")

// signSession encodes payload as base64url(json) + "." +
// base64url(HMAC-SHA256(secret, json)) - a minimal, dependency-free
// signed-cookie format (deliberately not a JWT: pave-api is both issuer
// and sole verifier, so there's no need for JWT's algorithm-negotiation
// surface, which is a common source of auth bugs).
func signSession(secret []byte, p sessionPayload) (string, error) {
	raw, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("marshal session: %w", err)
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "." + sig, nil
}

// verifySession checks the HMAC signature and expiry, returning the
// Identity it carries only if both are valid.
func verifySession(secret []byte, value string) (Identity, error) {
	dot := -1
	for i := len(value) - 1; i >= 0; i-- {
		if value[i] == '.' {
			dot = i
			break
		}
	}
	if dot < 0 {
		return Identity{}, errInvalidSession
	}
	payload, sig := value[:dot], value[dot+1:]

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return Identity{}, errInvalidSession
	}

	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return Identity{}, errInvalidSession
	}
	var p sessionPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return Identity{}, errInvalidSession
	}
	if time.Now().After(p.ExpiresAt) {
		return Identity{}, errInvalidSession
	}
	return Identity{Login: p.Login, Teams: p.Teams}, nil
}
