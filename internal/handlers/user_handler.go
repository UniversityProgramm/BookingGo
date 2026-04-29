package handlers

import (
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Методы для эндпоинтов
func GetAllUsers(c *gin.Context) {
	users := []entity.User{
		{ID: 1, Email: "test@example.com", FIO: "Иванов И.И.", Role: enum.RoleClient},
	}
	c.JSON(http.StatusOK, users)
}
