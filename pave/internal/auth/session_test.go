package auth

import (
	"testing"
	"time"
)

func TestSignAndVerifySessionRoundTrip(t *testing.T) {
	secret := []byte("test-secret")
	p := sessionPayload{Login: "octocat", Teams: []string{"platform"}, ExpiresAt: time.Now().Add(time.Hour)}

	cookie, err := signSession(secret, p)
	if err != nil {
		t.Fatalf("signSession: %v", err)
	}

	identity, err := verifySession(secret, cookie)
	if err != nil {
		t.Fatalf("verifySession: %v", err)
	}
	if identity.Login != "octocat" || !identity.HasTeam("platform") {
		t.Errorf("unexpected identity: %+v", identity)
	}
}

func TestVerifySessionRejectsTamperedPayload(t *testing.T) {
	secret := []byte("test-secret")
	cookie, err := signSession(secret, sessionPayload{Login: "octocat", ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatalf("signSession: %v", err)
	}

	tampered := cookie[:len(cookie)-1] + "x"
	if tampered == cookie {
		t.Fatal("test setup: tampering did not change the cookie")
	}
	if _, err := verifySession(secret, tampered); err == nil {
		t.Error("expected verification to fail for a tampered cookie")
	}
}

func TestVerifySessionRejectsWrongSecret(t *testing.T) {
	cookie, err := signSession([]byte("secret-a"), sessionPayload{Login: "octocat", ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatalf("signSession: %v", err)
	}
	if _, err := verifySession([]byte("secret-b"), cookie); err == nil {
		t.Error("expected verification to fail for the wrong secret")
	}
}

func TestVerifySessionRejectsExpired(t *testing.T) {
	secret := []byte("test-secret")
	cookie, err := signSession(secret, sessionPayload{Login: "octocat", ExpiresAt: time.Now().Add(-time.Minute)})
	if err != nil {
		t.Fatalf("signSession: %v", err)
	}
	if _, err := verifySession(secret, cookie); err == nil {
		t.Error("expected verification to fail for an expired session")
	}
}

func TestVerifySessionRejectsMalformed(t *testing.T) {
	cases := []string{"", "no-dot-here", ".", "abc.", ".xyz"}
	for _, c := range cases {
		if _, err := verifySession([]byte("secret"), c); err == nil {
			t.Errorf("expected verification to fail for malformed value %q", c)
		}
	}
}

func TestIdentityHasTeam(t *testing.T) {
	id := Identity{Login: "octocat", Teams: []string{"platform", "payments"}}
	if !id.HasTeam("platform") {
		t.Error("expected HasTeam(platform) to be true")
	}
	if id.HasTeam("not-a-member") {
		t.Error("expected HasTeam(not-a-member) to be false")
	}
}
