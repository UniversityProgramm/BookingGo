package ratelimit

import (
	"BookingGo/internal/auth"
	"BookingGo/pkg/logger"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware(limiter Limiter, config Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := getIdentifier(c)

		endpointConfig := getEndpointConfig(c.Request.URL.Path, config)

		key := fmt.Sprintf("%s:%s", identifier, c.Request.URL.Path)

		// Проверяем лимит
		result, err := limiter.Allow(c.Request.Context(), key, endpointConfig.Limit, endpointConfig.Window)
		if err != nil {
			logger.Log.Warn("[RateLimit] Limiter error, allowing request", "error", err.Error(), "key", key)
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, result.Remaining)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetAt.Unix()))

		if !result.Allowed {
			logger.Log.Warn("[RateLimit] Rate limit exceeded",
				"key", key,
				"limit", result.Limit,
				"ip", c.ClientIP(),
				"path", c.Request.URL.Path,
			)

			c.Header("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": int(result.RetryAfter.Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getIdentifier(c *gin.Context) string {
	currentUser := auth.GetCurrentUser(c)
	if currentUser != nil && currentUser.UserID > 0 {
		return fmt.Sprintf("user:%d", currentUser.UserID)
	}
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

func getEndpointConfig(path string, config Config) EndpointConfig {
	for prefix, cfg := range config.Endpoints {
		if strings.HasPrefix(path, prefix) {
			return cfg
		}
	}
	return config.Default
}
