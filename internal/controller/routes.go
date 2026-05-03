package controller

import (
	"BookingGo/internal/usecase"

	"github.com/gin-gonic/gin"
)

// Обрабатывает пути эндпоинтов
func SetupRoutes(r *gin.Engine, userUsecase *usecase.UserUsecase) {
	api := r.Group("/api")
	userController := NewUserController(userUsecase)
	{
		usersGroup := api.Group("/users")
		{
			usersGroup.GET("", userController.GetAllUsers)
			usersGroup.GET("/:id", userController.GetUserByID)
			usersGroup.GET("/email/:email", userController.GetUserByEmail)
			usersGroup.POST("", userController.CreateUser)
			usersGroup.PUT("/:id", userController.UpdateUser)
			usersGroup.DELETE("/:id", userController.DeleteUser)
		}
	}
}
