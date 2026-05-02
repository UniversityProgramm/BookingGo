package main

import (
	"BookingGo/internal/controller"
	"BookingGo/internal/repository"
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

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URl в .env не задан")
	}

	err := db.InitDB(dbUrl)
	if err != nil {
		log.Fatal("Ошибка при подключении к БД:", err)
	}
	log.SetOutput(os.Stdout)
	log.Println("Подключились к базе PostgreSQL")

	// Роутеризация запросов
	router := gin.Default()
	userRepo := repository.NewUserRepository()
	controller.SetupRoutes(router, userRepo)

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
