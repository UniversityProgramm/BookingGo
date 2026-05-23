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
		profileGroup := protectedGroup.Group("/profile")
		{
			profileGroup.GET("", authController.GetMe)
			profileGroup.PUT("", authController.UpdateMe)
			profileGroup.PUT("/password", authController.ChangePassword)
			profileGroup.PUT("/email", authController.ChangeEmail)
		}
		totpGroup := protectedGroup.Group("/otp")
		{
			totpGroup.POST("/setup", authController.SetupTotp)
			totpGroup.POST("/verify", authController.VerifyTotp)
			totpGroup.POST("/disable", authController.DisableTotp)
		}
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

	userController := controller.NewUserController(userUseCase)
	staffGroup := api.Group("/staffPanel")
	staffGroup.Use(middleware.AuthMiddleware())
	staffGroup.Use(middleware.StaffOnly())
	{
		userStaffGroup := staffGroup.Group("/users")
		{
			userStaffGroup.GET("", userController.GetAllUsers)
			userStaffGroup.GET("/:id", userController.GetUserByID)
			userStaffGroup.GET("/email/:email", userController.GetUserByEmail)
		}

		bookingStaffGroup := staffGroup.Group("/bookings")
		{
			bookingStaffGroup.GET("", bookingController.GetAllBookings)
			bookingStaffGroup.POST("/completeBooking/:id", bookingController.CompleteBookingByID)
		}
	}

	adminGroup := api.Group("/adminPanel")
	adminGroup.Use(middleware.AuthMiddleware())
	adminGroup.Use(middleware.AdminOnly())
	{
		userAdminGroup := adminGroup.Group("/users")
		{
			userAdminGroup.GET("", userController.GetAllUsers)
			userAdminGroup.GET("/:id", userController.GetUserByID)
			userAdminGroup.GET("/email/:email", userController.GetUserByEmail)
			userAdminGroup.POST("", userController.CreateUser)
			userAdminGroup.PUT("/:id", userController.UpdateUser)
			userAdminGroup.DELETE("/:id", userController.DeleteUser)
		}

		bookingAdminGroup := adminGroup.Group("/bookings")
		{
			bookingAdminGroup.GET("", bookingController.GetAllBookings)
		}
	}
}
