package competitor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "competitor:prices:"

// Cache stores scraped competitor prices in Redis.
type Cache struct {
	redis *redis.Client
	ttl   time.Duration
}

// NewCache creates a new competitor price cache.
// ttl is how long prices persist in Redis (should be generous, e.g. 24h,
// so stale data survives scraper failures).
func NewCache(redisClient *redis.Client, ttl time.Duration) *Cache {
	return &Cache{redis: redisClient, ttl: ttl}
}

// Set stores a single provider's prices in Redis.
func (c *Cache) Set(ctx context.Context, provider string, prices Prices) error {
	data, err := json.Marshal(prices)
	if err != nil {
		return fmt.Errorf("marshal %s prices: %w", provider, err)
	}
	return c.redis.Set(ctx, keyPrefix+provider, data, c.ttl).Err()
}

// GetAll reads all cached competitor prices, keyed by provider name.
// Missing providers are silently omitted (not an error).
func (c *Cache) GetAll(ctx context.Context) (map[string]Prices, error) {
	providers := []string{"Lambda", "CoreWeave", "AWS"}
	result := make(map[string]Prices, len(providers))

	for _, name := range providers {
		data, err := c.redis.Get(ctx, keyPrefix+name).Bytes()
		if err == redis.Nil {
			continue // No cached data for this provider
		}
		if err != nil {
			return nil, fmt.Errorf("get %s prices: %w", name, err)
		}

		var prices Prices
		if err := json.Unmarshal(data, &prices); err != nil {
			continue // Corrupted data, skip
		}
		result[name] = prices
	}

	return result, nil
}
