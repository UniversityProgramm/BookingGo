package controller

import (
	"BookingGo/internal/middleware"
	"BookingGo/internal/repository"
	"BookingGo/internal/usecase"

	"github.com/gin-gonic/gin"
)

// Обрабатывает пути эндпоинтов
func SetupRoutes(r *gin.Engine, userUsecase *usecase.UserUsecase, authUsecase *usecase.AuthUsecase, userRepo *repository.UserRepository) {
	api := r.Group("/api")

	authController := NewAuthController(authUsecase)
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/register", authController.Register)
	}

	protectedGroup := api.Group("/me")
	protectedGroup.Use(middleware.AuthMiddleware())
	{
		protectedGroup.GET("", authController.GetMe)
		protectedGroup.PUT("", authController.UpdateMe)
		protectedGroup.PUT("/password", authController.ChangePassword)
	}

	bookingRepo := repository.NewBookingRepository()
	bookingUsecase := usecase.NewBookingUsecase(bookingRepo, userRepo)
	bookingController := NewBookingController(bookingUsecase)
	bookingGroup := protectedGroup.Group("/bookings")
	{
		bookingGroup.GET("", bookingController.GetMyBookings)
		bookingGroup.POST("", bookingController.CreateBooking)
		bookingGroup.DELETE("/:id", bookingController.DeleteBooking)
	}

	adminGroup := api.Group("")
	adminGroup.Use(middleware.AuthMiddleware())
	adminGroup.Use(middleware.AdminOnly())
	{
		userController := NewUserController(userUsecase)
		usersGroup := adminGroup.Group("/users")
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
