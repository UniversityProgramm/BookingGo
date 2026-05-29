package app

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/domain"
	"BookingGo/internal/repository"
	"BookingGo/internal/totp"
	"BookingGo/internal/usecase"
	"BookingGo/internal/worker"
	"BookingGo/pkg/db"
	"BookingGo/pkg/logger"
	"BookingGo/pkg/natsClient"
	"errors"

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

	err = natsClient.Init()
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNatsConnection):
			logger.Log.Error("[app] Failed to connect to NATS", "error", err.Error())
		case errors.Is(err, domain.ErrJetStreamConnection):
			logger.Log.Error("[app] Failed to initialize JetStream", "error", err.Error())
		case errors.Is(err, domain.ErrStreamAdd):
			logger.Log.Error("[app] Failed to create NATS stream", "stream_name", "BOOKING", "error", err.Error())
		case errors.Is(err, domain.ErrStreamAlreadyExists):
			logger.Log.Debug("[app] Stream already exists, skipping creation", "stream_name", "BOOKING")
		default:
			logger.Fatal("[app] Failed to init NATS", "error", err.Error())
		}
	}
	logger.Log.Info("[app] NATS stream created successfully",
		"stream_name", "BOOKINGS",
		"subjects", []string{"booking.external.create"},
		"storage", "file",
	)

	bookingWorker := worker.NewExternalBookingWorker(bookingUseCase, userUseCase)
	if err := bookingWorker.Start(); err != nil {
		logger.Fatal("[app] Failed to start external booking worker", "error", err.Error())
	}
	logger.Log.Info("[app] External booking worker is running")

	SetupRoutes(router, userUseCase, authUseCase, bookingUseCase, notificationUseCase)
	logger.Log.Info("[app] Router is running")
}
