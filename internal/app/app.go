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

	userRepo := repository.NewUserRepository()
	bookingRepo := repository.NewBookingRepository()

	userUseCase := usecase.NewUserUseCase(userRepo)
	authUseCase := usecase.NewAuthUseCase(userUseCase)
	bookingUseCase := usecase.NewBookingUseCase(bookingRepo, userRepo)

	controller.SetupRoutes(router, userUseCase, authUseCase, bookingUseCase)
}
