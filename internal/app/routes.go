package app

import (
	"BookingGo/internal/controller"
	"BookingGo/internal/middleware"
	"BookingGo/internal/usecase"

	"github.com/gin-gonic/gin"
)

// Обрабатывает пути эндпоинтов
func SetupRoutes(r *gin.Engine, userUseCase *usecase.UserUseCase, authUseCase *usecase.AuthUseCase, bookingUseCase *usecase.BookingUseCase, notificationUseCase *usecase.NotificationUseCase) {
	api := r.Group("/api")

	authController := controller.NewAuthController(authUseCase)
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
		protectedGroup.PUT("/email", authController.ChangeEmail)
	}

	bookingController := controller.NewBookingController(bookingUseCase)
	bookingGroup := protectedGroup.Group("/bookings")
	{
		bookingGroup.GET("", bookingController.GetMyBookings)
		bookingGroup.POST("", bookingController.CreateBooking)
		bookingGroup.POST("/cancelBooking/:id", bookingController.CancelMyBooking)
	}

	notificationController := controller.NewNotificationController(notificationUseCase)
	notificationGroup := protectedGroup.Group("/notifications")
	{
		notificationGroup.GET("", notificationController.GetMyNotifications)
		notificationGroup.PATCH("/settings", notificationController.UpdateNotificationSettings)
	}

	staffGroup := api.Group("/staffPanel")
	staffGroup.Use(middleware.AuthMiddleware())
	staffGroup.Use(middleware.StaffOnly())
	{
		userController := controller.NewUserController(userUseCase)
		usersGroup := staffGroup.Group("/users")
		{
			usersGroup.GET("", userController.GetAllUsers)
			usersGroup.GET("/:id", userController.GetUserByID)
			usersGroup.GET("/email/:email", userController.GetUserByEmail)
		}
		bookingStaffController := controller.NewBookingController(bookingUseCase)
		bookingStaffGroup := staffGroup.Group("/bookings")
		{
			bookingStaffGroup.GET("", bookingStaffController.GetAllBookings)
			bookingStaffGroup.POST("/completeBooking/:id", bookingController.CompleteBookingByID)
		}
	}

	adminGroup := api.Group("/adminPanel")
	adminGroup.Use(middleware.AuthMiddleware())
	adminGroup.Use(middleware.AdminOnly())
	{
		userController := controller.NewUserController(userUseCase)
		usersGroup := adminGroup.Group("/users")
		{
			usersGroup.GET("", userController.GetAllUsers)
			usersGroup.GET("/:id", userController.GetUserByID)
			usersGroup.GET("/email/:email", userController.GetUserByEmail)
			usersGroup.POST("", userController.CreateUser)
			usersGroup.PUT("/:id", userController.UpdateUser)
			usersGroup.DELETE("/:id", userController.DeleteUser)
		}
		bookingAdminController := controller.NewBookingController(bookingUseCase)
		bookingAdminGroup := adminGroup.Group("/bookings")
		{
			bookingAdminGroup.GET("", bookingAdminController.GetAllBookings)
		}
	}
}
