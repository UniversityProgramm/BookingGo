package app

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/repository"
	"BookingGo/internal/totp"
	"BookingGo/internal/usecase"
	"BookingGo/internal/worker"
	"BookingGo/pkg/db"
	"BookingGo/pkg/logger"
	"BookingGo/pkg/natsClient"

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
	totpService := totp.NewTotpService("Booking OTP")
	logger.Log.Info("[app] All repositories and services initialized")

	userUseCase := usecase.NewUserUseCase(userRepo)
	notificationUseCase := usecase.NewNotificationUseCase(notificationRepo, bookingRepo, userRepo)
	authUseCase := usecase.NewAuthUseCase(userUseCase, outboxRepo, totpService)
	bookingUseCase := usecase.NewBookingUseCase(bookingRepo, userRepo, outboxRepo)
	logger.Log.Info("[app] All usecases initialized")

	outboxWorker := worker.NewOutboxWorker(outboxRepo, notificationUseCase)
	go outboxWorker.Start()
	logger.Log.Info("[app] Outbox worker is running")

	if err := natsClient.Init(); err != nil {
		logger.Fatal("[app] Failed to init NATS", "error", err.Error())
	}

	bookingWorker := worker.NewExternalBookingWorker(bookingUseCase, userUseCase)
	if err := bookingWorker.Start(); err != nil {
		logger.Fatal("[app] Failed to start booking worker", "error", err.Error())
	}

	SetupRoutes(router, userUseCase, authUseCase, bookingUseCase, notificationUseCase)
	logger.Log.Info("[app] Router is running")
}
