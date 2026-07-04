package ratelimit

import (
	"BookingGo/internal/cache"
	"BookingGo/pkg/logger"
	"context"
	"time"
)

type RateResult struct {
	Allowed    bool
	Limit      int
	Remaining  int
	ResetAt    time.Time
	RetryAfter time.Duration
}

type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (RateResult, error)
}

// Создаёт rate limiter, используя существующий Cache
func Init(cacheService cache.Cache) Limiter {
	limiter, err := NewRedisLimiter(cacheService)
	if err != nil {
		logger.Log.Warn("[RateLimit] Failed to create Redis limiter, using noop",
			"error", err.Error(),
		)
		return NewNoopLimiter()
	}

	logger.Log.Info("[RateLimit] Redis rate limiter initialized (shared cache)")
	return limiter
}
