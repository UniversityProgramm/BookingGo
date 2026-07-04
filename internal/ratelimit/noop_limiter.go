package ratelimit

import (
	"context"
	"time"
)

type NoopLimiter struct{}

func NewNoopLimiter() *NoopLimiter {
	return &NoopLimiter{}
}

func (l *NoopLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (RateResult, error) {
	return RateResult{
		Allowed:   true,
		Limit:     limit,
		Remaining: limit,
		ResetAt:   time.Now().Add(window),
	}, nil
}
