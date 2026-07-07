package apiserver

import "testing"

func TestIPRateLimiterEnforcesBurst(t *testing.T) {
	// rate=0 means no refill ever, so behavior is fully deterministic:
	// exactly `burst` calls succeed, everything after is blocked.
	l := newIPRateLimiter(0, 3)

	for i := 0; i < 3; i++ {
		if !l.allow("1.2.3.4") {
			t.Fatalf("expected call %d within burst to be allowed", i)
		}
	}
	if l.allow("1.2.3.4") {
		t.Fatal("expected call beyond burst to be blocked")
	}
}

func TestIPRateLimiterTracksKeysIndependently(t *testing.T) {
	l := newIPRateLimiter(0, 1)

	if !l.allow("1.2.3.4") {
		t.Fatal("expected first call from 1.2.3.4 to be allowed")
	}
	if l.allow("1.2.3.4") {
		t.Fatal("expected second call from 1.2.3.4 to be blocked")
	}
	if !l.allow("5.6.7.8") {
		t.Fatal("expected first call from a different IP to be allowed independently")
	}
}
