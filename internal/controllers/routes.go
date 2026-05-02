package controllers

import (
	"github.com/gin-gonic/gin"
)

// Обрабатывает пути эндпоинтов
func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		usersGroup := api.Group("/users")
		{
			usersGroup.GET("", GetAllUsers)
			usersGroup.GET("/:id", GetUserByID)
			usersGroup.POST("", CreateUser)
			usersGroup.PUT("/:id", UpdateUser)
			usersGroup.DELETE("/:id", DeleteUser)
		}
	}
}
