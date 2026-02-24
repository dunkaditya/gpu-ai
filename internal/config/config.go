// Package config loads gpuctl configuration from environment variables.
package config

import (
	"fmt"
	"os"
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

	return &Config{
		Port:             getEnvDefault("GPUCTL_PORT", "9090"),
		DatabaseURL:      databaseURL,
		RedisURL:         redisURL,
		InternalAPIToken: internalAPIToken,
		RunPodAPIKey:     os.Getenv("RUNPOD_API_KEY"),
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
