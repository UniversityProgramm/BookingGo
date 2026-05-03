package main

import (
	"BookingGo/internal/controller"
	"BookingGo/internal/repository"
	"BookingGo/internal/usecase"
	"BookingGo/pkg/db"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	loadErr := godotenv.Load()
	if loadErr != nil {
		log.Fatal("Error loading .env file")
	}

	err := db.InitDB()
	if err != nil {
		log.Fatal("Ошибка при подключении к БД:", err.Error())
	}
	log.SetOutput(os.Stdout)
	log.Println("Подключились к базе PostgreSQL")

	// Роутеризация запросов
	router := gin.Default()
	userRepo := repository.NewUserRepository()
	userUsecase := usecase.NewUserUsecase(userRepo)
	controller.SetupRoutes(router, userUsecase)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	log.SetOutput(os.Stdout)
	log.Printf("\n\nСервер запущен на порте %s\n\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
