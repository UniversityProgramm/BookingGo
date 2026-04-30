package routes

import (
	"BookingGo/internal/handlers"

	"github.com/gin-gonic/gin"
)

// Обрабатывает пути эндпоинтов
func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/users", handlers.GetAllUsers)
		api.GET("/users/:id", handlers.GetUserByID)
		api.POST("/users", handlers.CreateUser)
	}
}
