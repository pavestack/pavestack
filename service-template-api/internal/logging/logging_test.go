package logging_test

import (
	"errors"
	"testing"

	"github.com/pavestack/service-template-api/internal/logging"
)

func TestNewLogger(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "unknown"}

	for _, lvl := range levels {
		t.Run(lvl, func(t *testing.T) {
			logger := logging.New(lvl)
			if logger == nil {
				t.Fatal("expected non-nil logger")
			}
			// Verify fields can be logged without panic
			logger.Info("test message", logging.String("key", "value"), logging.Error(errors.New("test error")))
		})
	}
}
