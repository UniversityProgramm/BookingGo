package main

import (
	"BookingGo/internal/app"
	"BookingGo/pkg/logger"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	logger.Init()
	logger.Log.Info("[main] Starting app...", "version", "1.0.0")

	router := gin.Default()
	workers, err := app.Init(router)
	if err != nil {
		logger.Fatal("[main] Failed to initialize app", "err", err.Error())
	}
	defer workers.Close()
	logger.Log.Info("[main] Router is initialized")

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		logger.Log.Info("[main] Server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("[main] Server failed", "error", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Log.Info("[main] Shutdown signal received", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	logger.Log.Info("[main] Shutting down HTTP server")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("[main] HTTP server shutdown error", "error", err.Error())
	}

	logger.Log.Info("[main] Server exited")
}
