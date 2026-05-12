package app

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/controller"
	"BookingGo/internal/repository"
	"BookingGo/internal/usecase"
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

	userUseCase := usecase.NewUserUseCase(userRepo)
	authUseCase := usecase.NewAuthUseCase(userUseCase)
	bookingUseCase := usecase.NewBookingUseCase(bookingRepo, userRepo)
	notificationUseCase := usecase.NewNotificationUseCase(notificationRepo)
	controller.SetupRoutes(router, userUseCase, authUseCase, bookingUseCase, notificationUseCase)
}
