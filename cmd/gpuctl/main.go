package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gpuai/gpuctl/internal/api"
	"github.com/gpuai/gpuctl/internal/config"
	"github.com/gpuai/gpuctl/internal/db"
	"github.com/gpuai/gpuctl/internal/provider"
	"github.com/gpuai/gpuctl/internal/provider/runpod"
	"github.com/gpuai/gpuctl/internal/provision"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Set up structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Create context that cancels on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Connect to Postgres with retry
	var dbPool *db.Pool
	err = db.ConnectWithRetry(ctx, "postgres", 5, func(ctx context.Context) error {
		pool, poolErr := db.NewPool(ctx, cfg.DatabaseURL)
		if poolErr != nil {
			return poolErr
		}
		dbPool = pool
		return nil
	})
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Connect to Redis with retry
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("failed to parse redis URL", "error", err)
		os.Exit(1)
	}
	redisClient := redis.NewClient(opt)
	defer redisClient.Close()

	err = db.ConnectWithRetry(ctx, "redis", 5, func(ctx context.Context) error {
		return redisClient.Ping(ctx).Err()
	})
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	// Set up provider registry and register providers.
	providerRegistry := provider.NewRegistry()
	if cfg.RunPodAPIKey != "" {
		runpodAdapter := runpod.NewAdapter(cfg.RunPodAPIKey)
		providerRegistry.Register(runpodAdapter)
		slog.Info("registered provider", "name", "runpod")
	}

	// Create provisioning engine.
	engine := provision.NewEngine(provision.EngineDeps{
		Registry: providerRegistry,
		DB:       dbPool,
		Config:   cfg,
		Logger:   logger,
	})

	// Create API server
	srv := api.NewServer(api.ServerDeps{
		DB:     dbPool,
		Redis:  redisClient,
		Config: cfg,
		Engine: engine,
	})

	httpServer := &http.Server{
		Addr:        ":" + cfg.Port,
		Handler:     srv.Handler(),
		ReadTimeout: 10 * time.Second,
		// WriteTimeout disabled (0) for SSE long-lived connections.
		// Per-handler timeouts should be added for production non-SSE routes.
		WriteTimeout: 0,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background
	go func() {
		slog.Info("gpuctl starting", "port", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
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
