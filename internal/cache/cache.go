package cache

import (
	"BookingGo/pkg/logger"
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:generate mockgen -source=cache.go -destination=cache_mocks.go -package=mocks
type Cache interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPrefix(ctx context.Context, prefix string) error
	IncrementWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error)
	GetClient() *redis.Client
	Close() error
}

func Init() Cache {
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		logger.Log.Warn("[RedisCache] REDIS_URL is not set, using noop cache")
		return NewNoopCache()
	}

	redisCache, err := NewRedisCache(redisUrl)
	if err != nil {
		logger.Log.Warn("[RedisCache] Redis not available, using noop cache", "error", err.Error())
		return NewNoopCache()
	}
	logger.Log.Info("[RedisCache] Redis connected")

	return redisCache
}
