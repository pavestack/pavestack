package apiserver

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// securityHeadersMiddleware sets response headers that don't depend on
// any handler-specific logic. pave-api only ever returns JSON (or, for
// /auth/*, redirects) - never HTML/JS it renders itself - so a maximally
// restrictive CSP is safe and there's no script/style allowlist to
// maintain.
func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

// ipRateLimiter is a per-client-IP token bucket. Entries are swept
// opportunistically (no background goroutine, so there's nothing to leak
// across the many short-lived Server instances tests create) whenever the
// map grows past a threshold.
type ipRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rateEntry
	rate     rate.Limit
	burst    int
	ttl      time.Duration
}

type rateEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newIPRateLimiter(r rate.Limit, burst int) *ipRateLimiter {
	return &ipRateLimiter{
		limiters: make(map[string]*rateEntry),
		rate:     r,
		burst:    burst,
		ttl:      10 * time.Minute,
	}
}

const rateLimiterSweepThreshold = 1000

func (l *ipRateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	e, ok := l.limiters[key]
	if !ok {
		e = &rateEntry{limiter: rate.NewLimiter(l.rate, l.burst)}
		l.limiters[key] = e
		if len(l.limiters) > rateLimiterSweepThreshold {
			for k, v := range l.limiters {
				if now.Sub(v.lastSeen) > l.ttl {
					delete(l.limiters, k)
				}
			}
		}
	}
	e.lastSeen = now
	return e.limiter.Allow()
}

// rateLimit wraps next with a 429 once limiter's per-IP budget is spent.
// Applied per-route (not globally) since login/callback warrant a much
// tighter budget than the general mutating endpoints.
func (s *Server) rateLimit(limiter *ipRateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.allow(clientKey(r)) {
			w.Header().Set("Retry-After", "1")
			writeError(w, http.StatusTooManyRequests, errRateLimited)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
