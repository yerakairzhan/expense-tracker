package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// AnalyticsCache stores JSON-serialised analytics query results.
// Keys: analytics:{userID}:{type}[:{param}...], e.g. analytics:42:summary:2024-01-01:2024-01-31
// Invalidation: SCAN + DEL on pattern analytics:{userID}:* after any transaction mutation.
type AnalyticsCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewAnalyticsCache(client *redis.Client, ttl time.Duration) *AnalyticsCache {
	if client == nil {
		return nil
	}
	return &AnalyticsCache{client: client, ttl: ttl}
}

// Get decodes a cached entry into dest. Returns (true, nil) on hit, (false, nil) on miss.
func (c *AnalyticsCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	if c == nil || c.client == nil {
		return false, nil
	}
	raw, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return true, json.Unmarshal(raw, dest)
}

// Set stores value as JSON with the configured TTL.
func (c *AnalyticsCache) Set(ctx context.Context, key string, value interface{}) error {
	if c == nil || c.client == nil {
		return nil
	}
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, b, c.ttl).Err()
}

// InvalidateUser removes all analytics cache entries for userID via SCAN + DEL.
// Called after every transaction Create / Update / Delete.
func (c *AnalyticsCache) InvalidateUser(ctx context.Context, userID int64) error {
	if c == nil || c.client == nil {
		return nil
	}
	pattern := fmt.Sprintf("analytics:%d:*", userID)
	var cursor uint64
	for {
		keys, next, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}

// Key builds a canonical cache key.
func AnalyticsCacheKey(userID int64, queryType string, params ...string) string {
	key := fmt.Sprintf("analytics:%d:%s", userID, queryType)
	for _, p := range params {
		key += ":" + p
	}
	return key
}
