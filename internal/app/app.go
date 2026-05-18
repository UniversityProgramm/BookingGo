package app

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/controller"
	"BookingGo/internal/repository"
	"BookingGo/internal/usecase"
	"BookingGo/internal/worker"
	"BookingGo/pkg/db"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func Run(router *gin.Engine) {
	err := db.InitDB()
	if err != nil {
		log.Fatal("Ошибка при подключении к БД:", err.Error())
	}
	log.SetOutput(os.Stdout)
	log.Println("Подключились к базе PostgreSQL")

	auth.InitAuth()

	userRepo := repository.NewUserRepository(db.DB)
	bookingRepo := repository.NewBookingRepository(db.DB)
	notificationRepo := repository.NewNotificationRepository(db.DB)
	outboxRepo := repository.NewOutboxRepository(db.DB)

	userUseCase := usecase.NewUserUseCase(userRepo)
	notificationUseCase := usecase.NewNotificationUseCase(notificationRepo, bookingRepo, userRepo)
	authUseCase := usecase.NewAuthUseCase(userUseCase, outboxRepo)
	bookingUseCase := usecase.NewBookingUseCase(bookingRepo, userRepo, outboxRepo)

	outboxWorker := worker.NewOutboxWorker(outboxRepo, notificationUseCase)
	go outboxWorker.Start()

	controller.SetupRoutes(router, userUseCase, authUseCase, bookingUseCase, notificationUseCase)
}
