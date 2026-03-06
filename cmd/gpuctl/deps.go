package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/gpuai/gpuctl/internal/availability"
	"github.com/gpuai/gpuctl/internal/config"
	"github.com/gpuai/gpuctl/internal/db"
	"github.com/gpuai/gpuctl/internal/provider"
	"github.com/gpuai/gpuctl/internal/provider/runpod"
	"github.com/gpuai/gpuctl/internal/provision"
	"github.com/gpuai/gpuctl/internal/wireguard"
	"github.com/redis/go-redis/v9"
	"golang.zx2c4.com/wireguard/wgctrl"
)

// commonDeps holds shared infrastructure dependencies used by multiple subcommands.
type commonDeps struct {
	Config   *config.Config
	DB       *db.Pool
	Redis    *redis.Client
	Registry *provider.Registry
	WGMgr    *wireguard.Manager
	IPAM     *wireguard.IPAM
	Engine   *provision.Engine
	Logger   *slog.Logger
}

// setupCommonDeps initializes config, DB, Redis, provider registry, WireGuard, and engine.
func setupCommonDeps(ctx context.Context, logger *slog.Logger) *commonDeps {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Connect to Postgres with retry.
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

	// Connect to Redis with retry.
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("failed to parse redis URL", "error", err)
		os.Exit(1)
	}
	redisClient := redis.NewClient(opt)

	err = db.ConnectWithRetry(ctx, "redis", 5, func(ctx context.Context) error {
		return redisClient.Ping(ctx).Err()
	})
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	// Set up provider registry.
	providerRegistry := provider.NewRegistry()
	if cfg.RunPodAPIKey != "" {
		runpodAdapter := runpod.NewAdapter(cfg.RunPodAPIKey)
		providerRegistry.Register(runpodAdapter)
		slog.Info("registered provider", "name", "runpod")
	}

	// Initialize WireGuard Manager and IPAM if configured.
	var wgMgr *wireguard.Manager
	var ipam *wireguard.IPAM

	if cfg.WGEncryptionKeyBytes != nil {
		wgClient, err := wgctrl.New()
		if err != nil {
			slog.Error("failed to create WireGuard client", "error", err)
			os.Exit(1)
		}
		// Note: wgClient.Close() is deferred by the caller.

		wgMgr = wireguard.NewManager(wgClient, nil, cfg.WGInterfaceName, logger)
		slog.Info("wireguard manager initialized", "interface", cfg.WGInterfaceName)

		ipam, err = wireguard.NewIPAM("10.0.0.0/16", logger)
		if err != nil {
			slog.Error("failed to initialize IPAM", "error", err)
			os.Exit(1)
		}
		slog.Info("wireguard IPAM initialized", "subnet", "10.0.0.0/16")
	} else {
		slog.Info("wireguard not configured, privacy layer disabled")
	}

	// Create provisioning engine.
	engine := provision.NewEngine(provision.EngineDeps{
		Registry:  providerRegistry,
		DB:        dbPool,
		Config:    cfg,
		Logger:    logger,
		WGManager: wgMgr,
		IPAM:      ipam,
	})

	return &commonDeps{
		Config:   cfg,
		DB:       dbPool,
		Redis:    redisClient,
		Registry: providerRegistry,
		WGMgr:    wgMgr,
		IPAM:     ipam,
		Engine:   engine,
		Logger:   logger,
	}
}

// newAvailCache creates an availability cache with the standard TTL.
func newAvailCache(redisClient *redis.Client) *availability.Cache {
	return availability.NewCache(redisClient, 35*time.Second)
}
