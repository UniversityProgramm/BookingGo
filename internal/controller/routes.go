package controller

import (
	"BookingGo/internal/middleware"
	"BookingGo/internal/usecase"

	"github.com/gin-gonic/gin"
)

// Обрабатывает пути эндпоинтов
func SetupRoutes(r *gin.Engine, userUseCase *usecase.UserUseCase, authUseCase *usecase.AuthUseCase, bookingUseCase *usecase.BookingUseCase, notificationUseCase *usecase.NotificationUseCase) {
	api := r.Group("/api")

	authController := NewAuthController(authUseCase)
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

	bookingController := NewBookingController(bookingUseCase)
	bookingGroup := protectedGroup.Group("/bookings")
	{
		bookingGroup.GET("", bookingController.GetMyBookings)
		bookingGroup.POST("", bookingController.CreateBooking)
		bookingGroup.DELETE("/:id", bookingController.DeleteBooking)
		bookingGroup.POST("/completeBooking/:id", bookingController.ChangeBookingStatus)
	}

	notificationController := NewNotificationController(notificationUseCase)
	notificationGroup := protectedGroup.Group("/notifications")
	{
		notificationGroup.GET("", notificationController.)
		notificationGroup.PATCH("/settings", notificationController.)
	}

	adminGroup := api.Group("")
	adminGroup.Use(middleware.AuthMiddleware())
	adminGroup.Use(middleware.AdminOnly())
	{
		userController := NewUserController(userUseCase)
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
