package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/O-Midey/switchboard/internal/admin"
	"github.com/O-Midey/switchboard/internal/backend"
	"github.com/O-Midey/switchboard/internal/config"
	"github.com/O-Midey/switchboard/internal/health"
	proxyhandler "github.com/O-Midey/switchboard/internal/proxy"
	"github.com/O-Midey/switchboard/internal/router"
)

func main() {
	if err := run(); err != nil {
		slog.Error("switchboard stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", envOr("SWITCHBOARD_CONFIG", config.DefaultPath), "path to JSON configuration")
	flag.Parse()
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel()}))
	slog.SetDefault(log)
	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	targets := make([]*backend.Target, 0, len(cfg.Backends))
	for _, backendConfig := range cfg.Backends {
		target, targetErr := proxyhandler.NewTarget(backendConfig.ID, backendConfig.URL, cfg.MetricsWindow.Value(), log)
		if targetErr != nil {
			return fmt.Errorf("initialize backend %q: %w", backendConfig.ID, targetErr)
		}
		targets = append(targets, target)
	}
	selectedRouter := router.New(cfg.RoutingStrategy)
	proxyServer := server(cfg.ProxyAddress, proxyhandler.NewHandler(selectedRouter, targets, log))
	adminServer := server(cfg.AdminAddress, admin.NewHandler(targets, selectedRouter.Name()))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	checker := health.NewChecker(log)
	for index, target := range targets {
		go checker.Start(ctx, target, cfg.Backends[index])
	}
	errCh := make(chan error, 2)
	startServer(log, "proxy", proxyServer, errCh)
	startServer(log, "admin", adminServer, errCh)
	select {
	case <-ctx.Done():
		log.Info("shutdown requested", "signal", ctx.Err())
	case serveErr := <-errCh:
		stop()
		if serveErr != nil {
			log.Error("listener stopped unexpectedly", "error", serveErr)
		}
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout.Value())
	defer cancel()
	return errors.Join(proxyServer.Shutdown(shutdownCtx), adminServer.Shutdown(shutdownCtx))
}

func server(address string, handler http.Handler) *http.Server {
	return &http.Server{Addr: address, Handler: handler, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
}

func startServer(log *slog.Logger, name string, server *http.Server, result chan<- error) {
	go func() {
		log.Info("listener started", "name", name, "address", server.Addr)
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			result <- nil
			return
		}
		result <- fmt.Errorf("%s listener: %w", name, err)
	}()
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func logLevel() slog.Level {
	if os.Getenv("SWITCHBOARD_LOG_LEVEL") == "debug" {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}
