package cache

import (
	"BookingGo/pkg/logger"
	"context"
	"os"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPrefix(ctx context.Context, prefix string) error
	Close() error
}

func Init() Cache {
	var cacheService Cache
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl != "" {
		redisCache, err := NewRedisCache(redisUrl)
		if err != nil {
			logger.Log.Warn("[RedisCache] Redis not available, using noop cache", "error", err.Error())
			cacheService = NewNoopCache()
		} else {
			cacheService = redisCache
			logger.Log.Info("[RedisCache] Redis connected")
		}
	} else {
		logger.Log.Warn("[RedisCache] REDIS_URL is not set, using noop cache")
		cacheService = NewNoopCache()
	}

	return cacheService
}
