package main

import (
	"BookingGo/internal/app"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if loadErr := godotenv.Load(); loadErr != nil {
		log.Fatal("Error loading .env file")
	}

	router := gin.Default()
	app.Run(router)

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
