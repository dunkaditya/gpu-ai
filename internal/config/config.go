// Package config loads gpuctl configuration from environment variables.
package config

import (
	"encoding/hex"
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

	// WGEncryptionKey is the hex-encoded AES-256-GCM key for encrypting
	// WireGuard private keys at rest. Must be exactly 64 hex characters (32 bytes).
	WGEncryptionKey string

	// WGEncryptionKeyBytes is the decoded 32-byte encryption key derived from
	// WGEncryptionKey. Not loaded from env; computed during Load().
	WGEncryptionKeyBytes []byte

	// WGProxyEndpoint is the public IP:port of the WireGuard proxy server.
	// Example: "203.0.113.1:51820".
	WGProxyEndpoint string

	// WGProxyPublicKey is the proxy server's WireGuard public key (base64).
	WGProxyPublicKey string

	// WGInterfaceName is the WireGuard interface name on the proxy server.
	// Default: "wg0".
	WGInterfaceName string

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

	// PricingMarkupPct is the percentage markup applied to provider prices for retail pricing.
	// Default: 15.0 (15% markup). Set via PRICING_MARKUP_PCT env var.
	PricingMarkupPct float64
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

	// WireGuard config is optional. All three WG vars must be present together or all absent.
	wgEncryptionKey := os.Getenv("WG_ENCRYPTION_KEY")
	wgProxyEndpoint := os.Getenv("WG_PROXY_ENDPOINT")
	wgProxyPublicKey := os.Getenv("WG_PROXY_PUBLIC_KEY")

	wgVars := map[string]string{
		"WG_ENCRYPTION_KEY":   wgEncryptionKey,
		"WG_PROXY_ENDPOINT":   wgProxyEndpoint,
		"WG_PROXY_PUBLIC_KEY": wgProxyPublicKey,
	}
	var wgPresent, wgMissing []string
	for name, val := range wgVars {
		if val != "" {
			wgPresent = append(wgPresent, name)
		} else {
			wgMissing = append(wgMissing, name)
		}
	}

	if len(wgPresent) > 0 && len(wgMissing) > 0 {
		return nil, fmt.Errorf("incomplete WireGuard configuration: have %s but missing %s (all three WG vars must be set together or all absent)",
			strings.Join(wgPresent, ", "), strings.Join(wgMissing, ", "))
	}

	// Validate and decode WG_ENCRYPTION_KEY only when present.
	var wgEncryptionKeyBytes []byte
	if wgEncryptionKey != "" {
		if len(wgEncryptionKey) != 64 {
			return nil, fmt.Errorf("WG_ENCRYPTION_KEY must be exactly 64 hex characters (32 bytes), got %d", len(wgEncryptionKey))
		}
		var err error
		wgEncryptionKeyBytes, err = hex.DecodeString(wgEncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("WG_ENCRYPTION_KEY is not valid hex: %w", err)
		}
	} else {
		slog.Info("WireGuard config not set, privacy layer disabled")
	}

	if os.Getenv("STRIPE_API_KEY") == "" {
		slog.Info("Stripe not configured, billing metering disabled")
	}

	// Parse pricing markup percentage (default 15.0%).
	pricingMarkupPctStr := getEnvDefault("PRICING_MARKUP_PCT", "15.0")
	pricingMarkupPct, err := strconv.ParseFloat(pricingMarkupPctStr, 64)
	if err != nil {
		return nil, fmt.Errorf("PRICING_MARKUP_PCT must be a valid number: %w", err)
	}

	return &Config{
		Port:                 getEnvDefault("GPUCTL_PORT", "9090"),
		DatabaseURL:          databaseURL,
		RedisURL:             redisURL,
		InternalAPIToken:     internalAPIToken,
		RunPodAPIKey:         os.Getenv("RUNPOD_API_KEY"),
		WGEncryptionKey:      wgEncryptionKey,
		WGEncryptionKeyBytes: wgEncryptionKeyBytes,
		WGProxyEndpoint:      wgProxyEndpoint,
		WGProxyPublicKey:     wgProxyPublicKey,
		WGInterfaceName:      getEnvDefault("WG_INTERFACE_NAME", "wg0"),
		ClerkSecretKey:       os.Getenv("CLERK_SECRET_KEY"),
		GpuctlPublicURL:     os.Getenv("GPUCTL_PUBLIC_URL"),
		StripeAPIKey:         os.Getenv("STRIPE_API_KEY"),
		StripeMeterEventName: os.Getenv("STRIPE_METER_EVENT_NAME"),
		PricingMarkupPct:     pricingMarkupPct,
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
