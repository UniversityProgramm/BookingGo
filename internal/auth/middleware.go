package auth

import (
	"BookingGo/internal/enum"
	"BookingGo/pkg/logger"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const ContextKeyUser = "user"

func AuthMiddleware(blacklist Blacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Требуется авторизация",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверный формат токена"})
			return
		}

		tokenString := parts[1]

		claims, err := ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Невалидный токен"})
			return
		}

		if blacklist.IsInBlacklist(c.Request.Context(), strconv.Itoa(claims.UserID)) {
			logger.Log.Warn("[AuthMiddleware] Request with blacklisted token", "user_id", claims.UserID, "jti", claims.ID, "ip", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Токен отозван"})
			return
		}

		if !blacklist.IsSessionValid(c.Request.Context(), claims.UserID, claims.IssuedAt.Time) {
			logger.Log.Warn("[AuthMiddleware] Session invalidated by password/email change", "user_id", claims.UserID, "jti", claims.ID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Сессия устарела. Войдите заново"})
			return
		}

		c.Set(ContextKeyUser, claims)
		c.Next()
	}
}

func GetCurrentUser(c *gin.Context) *Claims {
	if user, exists := c.Get(ContextKeyUser); exists {
		if claims, ok := user.(*Claims); ok {
			return claims
		}
	}
	return nil
}

func StaffOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil || user.Role != enum.RoleStaff {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Доступно только для Staff"})
			return
		}
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil || user.Role != enum.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Доступно только для Admin"})
			return
		}
		c.Next()
	}
}
