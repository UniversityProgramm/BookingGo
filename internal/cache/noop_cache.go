package cache

import (
	"BookingGo/internal/domain"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type NoopCache struct{}

func NewNoopCache() *NoopCache {
	return &NoopCache{}
}

func (c *NoopCache) Get(ctx context.Context, key string, dest any) error {
	return domain.ErrCacheKeyNotFound
}

func (c *NoopCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return nil
}

func (c *NoopCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (c *NoopCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	return nil
}

func (c *NoopCache) IncrementWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	return 0, nil
}

func (c *NoopCache) GetClient() *redis.Client {
	return nil
}

func (c *NoopCache) Close() error {
	return nil
}
