package handlers

import (
	"BookingGo/internal/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Методы для эндпоинтов
func GetAllUsers(c *gin.Context) {
	rep := repositories.NewUserRepository()
	users, err := rep.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось получить пользователей",
		})
		return
	}
	c.JSON(http.StatusOK, users)
}
