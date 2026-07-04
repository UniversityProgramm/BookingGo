package ratelimit

import (
	"BookingGo/internal/cache"
	"context"
	"fmt"
	"time"
)

type RedisLimiter struct {
	cache cache.Cache
}

func NewRedisLimiter(cache cache.Cache) (*RedisLimiter, error) {
	if cache == nil {
		return nil, fmt.Errorf("cache is nil")
	}

	if cache.GetClient() == nil {
		return nil, fmt.Errorf("cache is not Redis-based")
	}

	return &RedisLimiter{cache: cache}, nil
}

func (l *RedisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (RateResult, error) {
	windowNum := time.Now().Unix() / int64(window.Seconds())
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, windowNum)

	count, err := l.cache.IncrementWithTTL(ctx, windowKey, window)
	if err != nil {
		return RateResult{}, fmt.Errorf("failed to increment counter: %w", err)
	}

	resetAt := time.Now().Truncate(window).Add(window)
	remaining := limit - int(count)

	result := RateResult{
		Limit:     limit,
		Remaining: remaining,
		ResetAt:   resetAt,
	}

	if int(count) > limit {
		result.Allowed = false
		result.RetryAfter = time.Until(resetAt)
	} else {
		result.Allowed = true
	}

	return result, nil
}
