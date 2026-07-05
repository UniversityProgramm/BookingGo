package auth

import (
	"BookingGo/internal/cache"
	"context"
	"strconv"
	"time"
)

const (
	blacklistPrefix      = "blacklist:"
	sessionVersionPrefix = "session_ver:"
)

type Blacklist interface {
	AddToBlacklist(ctx context.Context, jti string, expiresAt time.Time) error
	IsInBlacklist(ctx context.Context, jti string) bool
	InvalidateAllSessions(ctx context.Context, userID int) error
	IsSessionValid(ctx context.Context, userID int, issuedAt time.Time) bool
}

type BlacklistService struct {
	cache cache.Cache
}

func NewBlacklistService(cache cache.Cache) Blacklist {
	return &BlacklistService{cache: cache}
}

func (b *BlacklistService) AddToBlacklist(ctx context.Context, jti string, expiredAt time.Time) error {
	key := blacklistPrefix + jti

	ttl := time.Until(expiredAt)
	if ttl < 0 {
		return nil
	}

	return b.cache.Set(ctx, key, "canceled", ttl)
}

func (b *BlacklistService) IsInBlacklist(ctx context.Context, jti string) bool {
	key := blacklistPrefix + jti

	var value string
	err := b.cache.Get(ctx, key, &value)

	return err == nil
}

func (b *BlacklistService) InvalidateAllSessions(ctx context.Context, userID int) error {
	key := sessionVersionPrefix + strconv.Itoa(userID)
	return b.cache.Set(ctx, key, time.Now().Unix(), 30*24*time.Hour)
}

func (b *BlacklistService) IsSessionValid(ctx context.Context, userID int, issuedAt time.Time) bool {
	key := sessionVersionPrefix + strconv.Itoa(userID)

	var version int64
	err := b.cache.Get(ctx, key, &version)

	// Если версии нет в кэше, значит токен валиден
	if err != nil {
		return true
	}

	// Если токен создан до последней смены пароля или почты, то он невалиден
	return issuedAt.Unix() >= version
}
