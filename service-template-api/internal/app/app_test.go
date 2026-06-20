package app_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/pavestack/service-template-api/internal/app"
)

func TestAppLifecycle(t *testing.T) {
	// Create a TCP listener on a random available port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	addr := ln.Addr().String()

	// Instantiate the App module.
	a := app.New()
	a.Listener = ln

	// Set environment variables for the test to ensure config parses correctly.
	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("LOG_LEVEL", "debug")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runErrChan := make(chan error, 1)
	go func() {
		runErrChan <- a.Run(ctx)
	}()

	// Perform an HTTP GET request to verify the server is handling requests.
	// Use a retry loop to give the server a moment to start serving.
	client := &http.Client{Timeout: 2 * time.Second}
	url := fmt.Sprintf("http://%s/health", addr)

	var resp *http.Response
	var getErr error
	for i := 0; i < 10; i++ {
		resp, getErr = client.Get(url)
		if getErr == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if getErr != nil {
		t.Fatalf("failed to query /health: %v", getErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.Status)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", body["status"])
	}
	if body["service"] != "test-service" {
		t.Errorf("expected service 'test-service', got %q", body["service"])
	}

	// Cancel the context to initiate graceful shutdown.
	cancel()

	// Wait for Run to complete and check for errors.
	select {
	case err := <-runErrChan:
		if err != nil {
			t.Errorf("app Run returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("app did not shut down within timeout after context cancellation")
	}
}
