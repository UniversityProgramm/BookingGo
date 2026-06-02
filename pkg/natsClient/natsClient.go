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

	var err error
	Connection, err = nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.Timeout(10*time.Second),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(100),

		nats.ReconnectBufSize(1024*1024), // 1 MB

		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.Log.Error("[NATS] Disconnected", "error", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Log.Info("[NATS] Reconnected", "url", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logger.Log.Info("[NATS] Connection closed")
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			logger.Log.Error("[NATS] Async error", "error", err, "subject", sub.Subject)
		}),
	)
	if err != nil {
		return domain.ErrNatsConnection
	}

	JS, err = Connection.JetStream(nats.MaxWait(10 * time.Second))
	if err != nil {
		return domain.ErrJetStreamConnection
	}

	streamConfig := &nats.StreamConfig{
		Name:     "BOOKING",
		Subjects: []string{"booking.external.create"}, // Канал для прихода
		Storage:  nats.FileStorage,
	}
	_, err = JS.AddStream(streamConfig)
	if err != nil && !isAlreadyExists(err) {
		return domain.ErrStreamAdd
	}

	if isAlreadyExists(err) {
		return domain.ErrStreamAlreadyExists
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
	}
}
