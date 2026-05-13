package middleware

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/enum"
	"net/http"

	"github.com/gin-gonic/gin"
)

const ContextKeyUser = "user"

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Требуется авторизация",
			})
			return
		}

		var tokenString string
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Неверный формат токена",
			})
			return
		}

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Set(ContextKeyUser, claims)
		c.Next()
	}
}

func GetCurrentUser(c *gin.Context) *auth.Claims {
	if user, exists := c.Get(ContextKeyUser); exists {
		if claims, ok := user.(*auth.Claims); ok {
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
