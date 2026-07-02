// Package app manages pave-api's process lifecycle - config loading,
// logging/telemetry init, and graceful shutdown - mirroring
// service-template-api/internal/app so both Go services in this monorepo
// start up and shut down the same way.
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

	"github.com/pavestack/pave/internal/apiserver"
	"github.com/pavestack/pave/internal/auth"
	"github.com/pavestack/pave/internal/config"
	"github.com/pavestack/pave/internal/logging"
	"github.com/pavestack/pave/internal/telemetry"
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
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	a.cfg = cfg

	a.logger = logging.New(cfg.LogLevel)
	defer func() {
		_ = a.logger.Sync()
	}()

	shutdownTelemetry, err := telemetry.Init(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
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

	var authSvc *auth.Service
	if cfg.DisableAuth {
		a.logger.Warn("PAVE_API_DISABLE_AUTH=true - running with no authentication on mutating endpoints; local dev/CI only")
	} else {
		authSvc = auth.NewService(auth.Config{
			ClientID:      cfg.GitHubClientID,
			ClientSecret:  cfg.GitHubClientSecret,
			SessionSecret: cfg.SessionSecret,
			GitHubOrg:     cfg.GitHubOrg,
			ApproverTeam:  cfg.ApproverTeam,
			BaseURL:       cfg.BaseURL,
			PortalURL:     cfg.PortalURL,
			CookieSecure:  cfg.CookieSecure,
		})
	}

	srv, err := apiserver.New(apiserver.Config{
		RepoRoot:     cfg.RepoRoot,
		DryRun:       cfg.DryRun,
		CORSOrigin:   cfg.CORSOrigin,
		ApproverTeam: cfg.ApproverTeam,
	}, a.logger, authSvc)
	if err != nil {
		a.logger.Error("failed to build apiserver", logging.Error(err))
		return fmt.Errorf("failed to build apiserver: %w", err)
	}

	if !apiserver.GitOpsToolsAvailable() {
		a.logger.Warn("git and/or gh not found on PATH - open_pr steps will fail (or, in dry-run mode, are already simulated)")
	}

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
		ln, err = net.Listen("tcp", cfg.ListenAddr)
		if err != nil {
			a.logger.Error("failed to listen", logging.Error(err))
			return fmt.Errorf("failed to listen: %w", err)
		}
		a.Listener = ln
	}
	a.httpServer.Addr = ln.Addr().String()

	serverErrChan := make(chan error, 1)
	go func() {
		a.logger.Info("pave-api listening",
			logging.String("addr", a.httpServer.Addr),
			logging.String("repoRoot", cfg.RepoRoot),
			zap.Bool("dryRun", cfg.DryRun),
		)
		if err := a.httpServer.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("graceful shutdown failed", logging.Error(err))
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	return nil
}
