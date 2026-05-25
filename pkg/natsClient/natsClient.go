package natsClient

import (
	"BookingGo/internal/domain"
	"BookingGo/pkg/logger"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

var (
	Connection *nats.Conn
	JS         nats.JetStreamContext
)

func Init() error {
	url := os.Getenv("NATS_URL")
	if url == "" {
		logger.Log.Info("[NATS] Custom NATS URL is not set, get default URL", "url", url)
		url = nats.DefaultURL
	}

	logger.Log.Info("[NATS] Connecting to NATS", "url", url)

	var err error
	Connection, err = nats.Connect(url, nats.RetryOnFailedConnect(true), nats.Timeout(10*time.Second))
	if err != nil {
		logger.Log.Error("[NATS] Failed to connect to NATS",
			"url", url,
			"error", err.Error(),
		)
		return domain.ErrNatsConnection
	}

	logger.Log.Info("[NATS] NATS connection established")

	JS, err = Connection.JetStream()
	if err != nil {
		logger.Log.Error("[NATS] Failed to initialize JetStream",
			"error", err.Error(),
		)
		return domain.ErrJetStreamConnection
	}

	logger.Log.Info("[NATS] JetStream context created")

	streamConfig := &nats.StreamConfig{
		Name:     "BOOKING",
		Subjects: []string{"booking.external.create"}, // Канал для прихода
		Storage:  nats.FileStorage,
	}
	_, err = JS.AddStream(streamConfig)
	if err != nil && !isAlreadyExists(err) {
		logger.Log.Error("[NATS] Failed to create NATS stream",
			"stream_name", "BOOKING",
			"error", err.Error(),
		)
		return domain.ErrStreamAdd
	}

	if isAlreadyExists(err) {
		logger.Log.Debug("[NATS] Stream already exists, skipping creation",
			"stream_name", "BOOKING",
		)
	} else {
		logger.Log.Info("[NATS] NATS stream created successfully",
			"stream_name", "BOOKINGS",
			"subjects", []string{"booking.external.create"},
			"storage", "file",
		)
	}

	return nil
}

func isAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "already in use")
}

func Close() {
	if Connection != nil {
		Connection.Close()
		logger.Log.Info("[NATS] NATS connection closed")
	}
}
