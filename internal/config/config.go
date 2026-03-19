// Package config loads gpuctl configuration from environment variables.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration values for gpuctl.
type Config struct {
	// Port is the HTTP listen port. Default "9090".
	Port string

	// DatabaseURL is the PostgreSQL connection string. Required.
	DatabaseURL string

	// RedisURL is the Redis connection string. Required.
	RedisURL string

	// InternalAPIToken is the shared secret used by cloud-init callbacks. Required.
	InternalAPIToken string

	// RunPodAPIKey is the RunPod API key for GPU provisioning. Optional.
	// The RunPod adapter is only created if this key is present.
	RunPodAPIKey string

	// ClerkSecretKey is the Clerk API secret key for JWT verification.
	// Optional -- empty disables Clerk auth for local dev.
	ClerkSecretKey string

	// GpuctlPublicURL is the public base URL of the gpuctl server
	// (e.g., "https://api.gpu.ai"). Used to construct callback URLs reachable
	// from GPU instances. Optional -- if empty, callback URLs fall back to
	// branded hostname (dev only).
	GpuctlPublicURL string

	// StripeAPIKey is the Stripe secret key for billing metering. Optional.
	// Billing metering is disabled if not set.
	StripeAPIKey string

	// StripeMeterEventName is the Stripe Billing Meter event name (e.g., "gpu_seconds").
	// Must match the meter configured in Stripe Dashboard.
	StripeMeterEventName string

	// StripeWebhookSecret is the signing secret for Stripe webhook signature verification.
	// Required for processing checkout.session.completed and payment_intent events.
	StripeWebhookSecret string

	// PricingMarkupPct is the percentage markup applied to provider prices for retail pricing.
	// Default: 15.0 (15% markup). Set via PRICING_MARKUP_PCT env var.
	PricingMarkupPct float64

	// FRPBindPort is the port that the embedded FRP server (frps) listens on
	// for frpc client connections. Default: 7000.
	FRPBindPort int

	// FRPToken is the shared auth token for frps<->frpc authentication.
	// Optional -- if empty, FRP tunneling is disabled.
	FRPToken string

	// FRPAllowPorts is the allowed remote port range for FRP proxies.
	// Default: "10000-10255".
	FRPAllowPorts string
}

// Load reads configuration from environment variables, validates required
// fields, and returns a Config or an error describing all missing variables.
func Load() (*Config, error) {
	var missing []string

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		missing = append(missing, "REDIS_URL")
	}

	internalAPIToken := os.Getenv("INTERNAL_API_TOKEN")
	if internalAPIToken == "" {
		missing = append(missing, "INTERNAL_API_TOKEN")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if internalAPIToken == "change-me" {
		return nil, fmt.Errorf("INTERNAL_API_TOKEN must be changed from default value 'change-me'")
	}

	if os.Getenv("STRIPE_API_KEY") == "" {
		slog.Info("Stripe not configured, billing metering disabled")
	}

	// Parse pricing markup percentage (default 6.0%).
	pricingMarkupPctStr := getEnvDefault("PRICING_MARKUP_PCT", "6.0")
	pricingMarkupPct, err := strconv.ParseFloat(pricingMarkupPctStr, 64)
	if err != nil {
		return nil, fmt.Errorf("PRICING_MARKUP_PCT must be a valid number: %w", err)
	}

	// FRP tunneling config (optional, replaces WireGuard for SSH access).
	frpToken := os.Getenv("FRP_TOKEN")
	frpAllowPorts := getEnvDefault("FRP_ALLOW_PORTS", "10000-10255")

	frpBindPortStr := getEnvDefault("FRP_BIND_PORT", "7000")
	frpBindPort, err := strconv.Atoi(frpBindPortStr)
	if err != nil {
		return nil, fmt.Errorf("FRP_BIND_PORT must be a valid integer: %w", err)
	}

	if frpToken == "" {
		slog.Info("FRP tunneling not configured, tunnel layer disabled")
	}

	return &Config{
		Port:                 getEnvDefault("GPUCTL_PORT", "9090"),
		DatabaseURL:          databaseURL,
		RedisURL:             redisURL,
		InternalAPIToken:     internalAPIToken,
		RunPodAPIKey:         os.Getenv("RUNPOD_API_KEY"),
		ClerkSecretKey:       os.Getenv("CLERK_SECRET_KEY"),
		GpuctlPublicURL:      os.Getenv("GPUCTL_PUBLIC_URL"),
		StripeAPIKey:         os.Getenv("STRIPE_API_KEY"),
		StripeMeterEventName: os.Getenv("STRIPE_METER_EVENT_NAME"),
		StripeWebhookSecret:  os.Getenv("STRIPE_WEBHOOK_SECRET"),
		PricingMarkupPct:     pricingMarkupPct,
		FRPBindPort:          frpBindPort,
		FRPToken:             frpToken,
		FRPAllowPorts:        frpAllowPorts,
	}, nil
}

// getEnvDefault returns the value of the environment variable named by key,
// or defaultVal if the variable is not set or is empty.
func getEnvDefault(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}
