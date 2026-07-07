package apiserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/pavestack/pave/internal/logging"
	"go.uber.org/zap"
)

type ctxKey int

const requestIDKey ctxKey = iota

// requestIDMiddleware assigns (or propagates, if the caller already sent
// one) a request ID, attaches it to the request context so downstream
// middleware/handlers can log it, and echoes it back in the response so a
// caller can correlate their request to a specific log line.
func (s *Server) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}

func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func requestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// loggingMiddleware logs one structured line per request, correlated to
// the active OTel trace/span (see logging.FromContext) and to the
// request ID assigned above.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		logging.FromContext(r.Context(), s.logger).Info("http request",
			logging.String("method", r.Method),
			logging.String("path", r.URL.Path),
			logging.String("request_id", requestIDFromContext(r.Context())),
			zap.Int("status", rec.status),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

// recoveryMiddleware converts a panic in any inner handler into a logged
// 500 response instead of crashing the whole process - previously a single
// panicking handler would take down pave-api entirely.
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logging.FromContext(r.Context(), s.logger).Error("panic recovered",
					zap.Any("panic", rec),
					zap.String("stack", string(debug.Stack())),
					logging.String("request_id", requestIDFromContext(r.Context())),
				)
				writeError(w, http.StatusInternalServerError, errInternal)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
