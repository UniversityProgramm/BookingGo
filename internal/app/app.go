package app

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/cache"
	"BookingGo/internal/domain"
	"BookingGo/internal/repository"
	"BookingGo/internal/totp"
	"BookingGo/internal/usecase"
	"BookingGo/internal/worker"
	"BookingGo/pkg/db"
	"BookingGo/pkg/logger"
	"BookingGo/pkg/natsClient"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	UserUseCase         *usecase.UserUseCase
	AuthUseCase         *usecase.AuthUseCase
	BookingUseCase      *usecase.BookingUseCase
	NotificationUseCase *usecase.NotificationUseCase
}

type Workers struct {
	OutboxWorker          *worker.OutboxWorker
	ExternalBookingWorker *worker.ExternalBookingWorker
	cacheService          cache.Cache
}

func (r *Workers) Close() {
	logger.Log.Info("[app] Closing resources...")

	if r.OutboxWorker != nil {
		r.OutboxWorker.Stop()
	}
	logger.Log.Info("[app] Outbox worker closed")

	if r.ExternalBookingWorker != nil {
		r.ExternalBookingWorker.Stop()
	}
	logger.Log.Info("[app] External booking worker closed")

	if r.cacheService != nil {
		if err := r.cacheService.Close(); err != nil {
			logger.Log.Warn("[app] Failed to close cache", "error", err.Error())
		} else {
			logger.Log.Info("[app] Cache connection closed")
		}
	}

	natsClient.Close()
	logger.Log.Info("[app] NATS connection closed")

	if db.DB != nil {
		if sqlDB, err := db.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				logger.Log.Error("[app] Failed to close database connection", "error", err.Error())
			} else {
				logger.Log.Info("[app] Database connection pool closed")
			}
		} else {
			logger.Log.Warn("[app] Could not retrieve sql.DB for closing", "error", err.Error())
		}
	}

	logger.Log.Info("[app] All resources closed")
}

func Init(router *gin.Engine) (*Workers, error) {
	err := db.InitDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Log.Info("[app] Connected to PostgreSQL base")

	auth.InitAuth()
	logger.Log.Info("[app] Auth initialized")

	userRepo := repository.NewUserRepository(db.DB)
	bookingRepo := repository.NewBookingRepository(db.DB)
	notificationRepo := repository.NewNotificationRepository(db.DB)
	outboxRepo := repository.NewOutboxRepository(db.DB)
	totpService := totp.NewTotpService("Booking OTP")
	logger.Log.Info("[app] All repositories and services initialized")

	cacheService := cache.Init()

	userUseCase := usecase.NewUserUseCase(userRepo)
	notificationUseCase := usecase.NewNotificationUseCase(notificationRepo, bookingRepo, userRepo, cacheService)
	authUseCase := usecase.NewAuthUseCase(userUseCase, outboxRepo, totpService, cacheService)
	bookingUseCase := usecase.NewBookingUseCase(bookingRepo, userRepo, outboxRepo, cacheService)
	logger.Log.Info("[app] All usecases initialized")

	outboxWorker := worker.NewOutboxWorker(outboxRepo, notificationUseCase)
	if err := outboxWorker.Start(); err != nil {
		return nil, fmt.Errorf("failed to start outbox worker: %w", err)
	}
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
			return nil, fmt.Errorf("failed to init NATS: %w", err)
		}
	}
	logger.Log.Info("[app] NATS stream created successfully",
		"stream_name", "BOOKINGS",
		"subjects", []string{"booking.external.create"},
		"storage", "file",
	)

	bookingWorker := worker.NewExternalBookingWorker(bookingUseCase, userUseCase)
	if err := bookingWorker.Start(); err != nil {
		return nil, fmt.Errorf("failed to start external booking worker: %w", err)
	}
	logger.Log.Info("[app] External booking worker is running")

	SetupRoutes(router, &Dependencies{
		UserUseCase:         userUseCase,
		AuthUseCase:         authUseCase,
		BookingUseCase:      bookingUseCase,
		NotificationUseCase: notificationUseCase,
	})
	logger.Log.Info("[app] Router is running")

	return &Workers{
		OutboxWorker:          outboxWorker,
		ExternalBookingWorker: bookingWorker,
		cacheService:          cacheService,
	}, nil
}
