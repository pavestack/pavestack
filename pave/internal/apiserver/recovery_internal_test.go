package apiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

// White-box test (package apiserver, not apiserver_test) so it can reach
// the unexported recoveryMiddleware directly with a handler that panics on
// purpose - the routed mux has no such handler, so this can't be exercised
// through the public Handler() surface.
func TestRecoveryMiddlewareConvertsPanicToInternalError(t *testing.T) {
	s := &Server{logger: zap.NewNop()}

	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/whatever", nil)
	rec := httptest.NewRecorder()

	// Must not panic out of the test itself.
	s.recoveryMiddleware(panicking).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 after recovered panic, got %d", rec.Code)
	}
}
