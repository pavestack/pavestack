package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pavestack/service-template-api/internal/config"
	"github.com/pavestack/service-template-api/internal/logging"
	"github.com/pavestack/service-template-api/internal/server"
	"github.com/pavestack/service-template-api/internal/telemetry"
	"go.uber.org/zap"
)

// App manages the application lifecycle.
type App struct {
	cfg        config.Config
	logger     *zap.Logger
	httpServer *http.Server
	Listener   net.Listener
}

// New creates a new App instance.
func New() *App {
	return &App{}
}

// Run executes the application lifecycle.
func (a *App) Run(ctx context.Context) error {
	// 1. Config loading
	a.cfg = config.Load()

	// 2. Logging initialization
	a.logger = logging.New(a.cfg.LogLevel)
	defer func() {
		_ = a.logger.Sync()
	}()

	// 3. Telemetry initialization
	shutdownTelemetry, err := telemetry.Init(ctx, a.cfg.ServiceName, a.cfg.OTLPEndpoint)
	if err != nil {
		a.logger.Error("failed to initialize telemetry", logging.Error(err))
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTelemetry(shutdownCtx); err != nil {
			a.logger.Error("telemetry shutdown failed", logging.Error(err))
		}
	}()

	// 4. Starting the HTTP server
	srv := server.New(a.cfg, a.logger)
	a.httpServer = &http.Server{
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	var ln net.Listener
	if a.Listener != nil {
		ln = a.Listener
	} else {
		var err error
		ln, err = net.Listen("tcp", a.cfg.ListenAddr)
		if err != nil {
			a.logger.Error("failed to listen", logging.Error(err))
			return fmt.Errorf("failed to listen: %w", err)
		}
		a.Listener = ln
	}
	a.httpServer.Addr = ln.Addr().String()

	serverErrChan := make(chan error, 1)
	go func() {
		a.logger.Info("server listening", logging.String("addr", a.httpServer.Addr))
		if err := a.httpServer.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

	// 5. Listening for OS signals AND context cancellation
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	select {
	case sig := <-stop:
		a.logger.Info("received signal, shutting down", logging.String("signal", sig.String()))
	case <-ctx.Done():
		a.logger.Info("context cancelled, shutting down")
	case err := <-serverErrChan:
		a.logger.Error("server failed, shutting down", logging.Error(err))
		return fmt.Errorf("server failed: %w", err)
	}

	// 6. Graceful shutdown of the server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("graceful shutdown failed", logging.Error(err))
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	return nil
}
