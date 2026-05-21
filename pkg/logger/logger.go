package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func Init() {
	level := slog.LevelInfo
	if os.Getenv("DEBUG") == "true" {
		level = slog.LevelDebug
	}

	var handler slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	Log = slog.New(handler)
}

func Fatal(msg string, args ...any) {
	Log.Error(msg, args...)
	os.Exit(1)
}
