package app

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/controller"
	"BookingGo/internal/repository"
	"BookingGo/internal/usecase"
	"BookingGo/internal/worker"
	"BookingGo/pkg/db"
	"BookingGo/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Run(router *gin.Engine) {
	err := db.InitDB()
	if err != nil {
		logger.Fatal("[app] Failed to connect to database", "error", err.Error())
	}
	logger.Log.Info("[app] Connected to PostgreSQL base")

	auth.InitAuth()
	logger.Log.Info("[app] Auth is initialized")

	userRepo := repository.NewUserRepository(db.DB)
	bookingRepo := repository.NewBookingRepository(db.DB)
	notificationRepo := repository.NewNotificationRepository(db.DB)
	outboxRepo := repository.NewOutboxRepository(db.DB)
	logger.Log.Info("[app] All repositories initialized")

	userUseCase := usecase.NewUserUseCase(userRepo)
	notificationUseCase := usecase.NewNotificationUseCase(notificationRepo, bookingRepo, userRepo)
	authUseCase := usecase.NewAuthUseCase(userUseCase, outboxRepo)
	bookingUseCase := usecase.NewBookingUseCase(bookingRepo, userRepo, outboxRepo)
	logger.Log.Info("[app] All usecases initialized")

	outboxWorker := worker.NewOutboxWorker(outboxRepo, notificationUseCase)
	go outboxWorker.Start()
	logger.Log.Info("[app] Outbox worker is running")

	controller.SetupRoutes(router, userUseCase, authUseCase, bookingUseCase, notificationUseCase)
	logger.Log.Info("[app] Router is running")
}
