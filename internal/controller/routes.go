package controller

import (
	"BookingGo/internal/repository"

	"github.com/gin-gonic/gin"
)

// Обрабатывает пути эндпоинтов
func SetupRoutes(r *gin.Engine, userRep *repository.UserRepository) {
	api := r.Group("/api")
	userController := UserController{userRepository: *userRep}
	{
		usersGroup := api.Group("/users")
		{
			usersGroup.GET("", userController.GetAllUsers)
			usersGroup.GET("/:id", userController.GetUserByID)
			usersGroup.POST("", userController.CreateUser)
			usersGroup.PUT("/:id", userController.UpdateUser)
			usersGroup.DELETE("/:id", userController.DeleteUser)
		}
	}
}
