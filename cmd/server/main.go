package main

import (
	"BookingGo/internal/routes"
	"BookingGo/pkg/db"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")

	if dbUrl == "" {
		log.Fatal("DB_URl в .env не задан")
	}
	db.InitDB(dbUrl)

	router := gin.Default()
	routes.SetupRoutes(router)

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
