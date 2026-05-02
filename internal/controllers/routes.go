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
			usersGroup.GET("/users", GetAllUsers)
			usersGroup.GET("/users/:id", GetUserByID)
			usersGroup.POST("/users", CreateUser)
			usersGroup.PUT("/users/:id", UpdateUser)
			usersGroup.DELETE("/users/:id", DeleteUser)
		}
	}
}
