package main

import (
	"BookingGo/internal/app"
	"BookingGo/pkg/logger"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if loadErr := godotenv.Load(); loadErr != nil {
		logger.Init()
		logger.Fatal("[main] Error loading .env file", "error", loadErr.Error())
	}
	logger.Init()
	logger.Log.Info("[main] Starting app...", "version", "1.0.0")

	logger.Log.Info("[main] Starting router...")
	router := gin.Default()
	app.Run(router)
	logger.Log.Info("[main] Router is initialized")

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		logger.Fatal("[main] Failed to run server", "error", err.Error())
	}
	logger.Log.Info("[main] Server is running", "port", port)
}
