package availability

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const offeringsKey = "gpu:offerings:all"

// Cache provides Redis-backed storage for GPU offerings.
// Uses a single key with JSON array for atomic reads (no SCAN needed).
type Cache struct {
	redis *redis.Client
	ttl   time.Duration
}

// NewCache creates a new Cache with the specified TTL.
func NewCache(redisClient *redis.Client, ttl time.Duration) *Cache {
	return &Cache{
		redis: redisClient,
		ttl:   ttl,
	}
}

// SetOfferings writes all offerings as a single JSON array to Redis with TTL.
// Atomic: readers always see a complete snapshot, never partial data.
func (c *Cache) SetOfferings(ctx context.Context, offerings []AvailableOffering) error {
	data, err := json.Marshal(offerings)
	if err != nil {
		return fmt.Errorf("marshal offerings: %w", err)
	}
	return c.redis.Set(ctx, offeringsKey, data, c.ttl).Err()
}

// GetOfferings reads all cached offerings from Redis.
// Returns nil slice (not error) if no data is cached yet.
func (c *Cache) GetOfferings(ctx context.Context) ([]AvailableOffering, error) {
	data, err := c.redis.Get(ctx, offeringsKey).Bytes()
	if err == redis.Nil {
		return nil, nil // No cached data yet
	}
	if err != nil {
		return nil, err
	}
	var offerings []AvailableOffering
	if err := json.Unmarshal(data, &offerings); err != nil {
		return nil, fmt.Errorf("unmarshal offerings: %w", err)
	}
	return offerings, nil
}
