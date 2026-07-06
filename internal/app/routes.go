package app

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/controller"
	"BookingGo/internal/ratelimit"

	_ "BookingGo/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes настраивает все маршруты API
//
// Структура API:
//
//	/api/auth/*              — публичные эндпоинты авторизации (тег: auth)
//	/api/me/profile/*        — профиль пользователя (тег: profile)
//	/api/me/otp/*            — двухфакторная аутентификация (тег: totp)
//	/api/me/bookings/*       — бронирования пользователя (тег: bookings)
//	/api/me/notifications/*  — уведомления пользователя (тег: notifications)
//	/api/staffPanel/*        — панель персонала (тег: management)
//	/api/adminPanel/*        — панель администратора (тег: management)
func SetupRoutes(r *gin.Engine, deps *Dependencies) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	rlConfig := ratelimit.DefaultConfig()

	r.Use(ratelimit.RateLimitMiddleware(deps.RateLimiter, rlConfig))

	api := r.Group("/api")

	authController := controller.NewAuthController(deps.AuthUseCase)
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/register", authController.Register)
		authGroup.POST("/logout", authController.Logout)
	}

	protectedGroup := api.Group("/me")
	protectedGroup.Use(auth.AuthMiddleware(deps.BlacklistService))
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

	bookingController := controller.NewBookingController(deps.BookingUseCase)
	bookingGroup := protectedGroup.Group("/bookings")
	{
		bookingGroup.GET("", bookingController.GetMyBookings)
		bookingGroup.POST("", bookingController.CreateBooking)
		bookingGroup.POST("/cancelBooking/:id", bookingController.CancelMyBooking)
	}

	notificationController := controller.NewNotificationController(deps.NotificationUseCase)
	notificationGroup := protectedGroup.Group("/notifications")
	{
		notificationGroup.GET("", notificationController.GetMyNotifications)
		notificationGroup.PATCH("/settings", notificationController.UpdateNotificationSettings)
	}

	userController := controller.NewUserController(deps.UserUseCase)
	staffGroup := api.Group("/staffPanel")
	staffGroup.Use(auth.AuthMiddleware(deps.BlacklistService))
	staffGroup.Use(auth.StaffOnly())
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
	adminGroup.Use(auth.AuthMiddleware(deps.BlacklistService))
	adminGroup.Use(auth.AdminOnly())
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
