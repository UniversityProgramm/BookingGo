package controllers

import (
	"github.com/gin-gonic/gin"
)

// Обрабатывает пути эндпоинтов
func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/users", GetAllUsers)
		api.GET("/users/:id", GetUserByID)
		api.POST("/users", CreateUser)
		api.PUT("/users/:id", UpdateUser)
		api.DELETE("/users/:id", DeleteUser)
	}
}
