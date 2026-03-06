package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gpuai/gpuctl/internal/api"
	"github.com/gpuai/gpuctl/internal/availability"
	"github.com/gpuai/gpuctl/internal/billing"
	"github.com/gpuai/gpuctl/internal/health"
)

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	logLevel := fs.String("log-level", "info", "Log level: debug, info, warn, error")
	logFormat := fs.String("log-format", "json", "Log format: json, text")
	fs.Parse(args)

	// Configure logger based on flags.
	var level slog.Level
	switch strings.ToLower(*logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if *logFormat == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Create context that cancels on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	deps := setupCommonDeps(ctx, logger)
	defer deps.DB.Close()
	defer deps.Redis.Close()

	// Create availability cache and poller.
	availCache := newAvailCache(deps.Redis)
	availPoller := availability.NewPoller(
		deps.Registry,
		availCache,
		30*time.Second,
		deps.Config.PricingMarkupPct,
		logger,
	)
	go availPoller.Start(ctx)
	slog.Info("availability poller started", "interval", "30s", "markup_pct", deps.Config.PricingMarkupPct)

	// Create API server.
	srv := api.NewServer(api.ServerDeps{
		DB:         deps.DB,
		Redis:      deps.Redis,
		Config:     deps.Config,
		Engine:     deps.Engine,
		AvailCache: availCache,
	})

	// Wire SSE status events from provisioning engine to API server.
	deps.Engine.SetOnStatusChange(srv.PublishStatusChange)

	// Create billing service (Stripe metering).
	billingSvc := billing.NewBillingService(deps.Config.StripeAPIKey, deps.Config.StripeMeterEventName, logger)

	// Create and start billing ticker.
	billingTicker := billing.NewBillingTicker(billing.TickerDeps{
		DB:     deps.DB,
		Engine: deps.Engine,
		Stripe: billingSvc,
		Logger: logger,
	})
	go billingTicker.Start(ctx)
	slog.Info("billing ticker started")

	// Create and start health monitor.
	healthMonitor := health.NewMonitor(health.MonitorDeps{
		DB:       deps.DB,
		Registry: deps.Registry,
		Logger:   logger,
		Interval: 60 * time.Second,
		OnEvent:  srv.PublishOrgEvent,
	})
	go healthMonitor.Start(ctx)
	slog.Info("health monitor started", "interval", "60s")

	httpServer := &http.Server{
		Addr:         ":" + deps.Config.Port,
		Handler:      srv.Handler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // Disabled for SSE long-lived connections.
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background.
	go func() {
		slog.Info("gpuctl starting", "port", deps.Config.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("gpuctl stopped")
}
