package worker

import (
	"BookingGo/internal/entity"
	"BookingGo/internal/usecase"
	"BookingGo/pkg/logger"
	"BookingGo/pkg/natsClient"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

type ExternalBookingWorker struct {
	bookingUseCase *usecase.BookingUseCase
	userUseCase    *usecase.UserUseCase
}

func NewExternalBookingWorker(bookingUseCase *usecase.BookingUseCase, userUseCase *usecase.UserUseCase) *ExternalBookingWorker {
	return &ExternalBookingWorker{
		bookingUseCase: bookingUseCase,
		userUseCase:    userUseCase,
	}
}
func (w *ExternalBookingWorker) Start() error {
	logger.Log.Info("[ExternalBookingWorker] Starting external booking worker")

	sub, err := natsClient.JS.PullSubscribe(
		"booking.external.create",    // куда приходят сообщения
		"booking-consumer",           // имя потребителя
		nats.AckWait(30*time.Second), // время на обработку перед авто-ретраем
		nats.MaxAckPending(10),       // максимальное кол-во сообщений в работе параллельно
	)
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	go func() {
		defer logger.Log.Info("[ExternalBookingWorker] Worker stopped")

		for {
			msgs, err := sub.Fetch(1, nats.MaxWait(2*time.Second))
			if err != nil {
				if errors.Is(err, nats.ErrTimeout) {
					continue
				}

				if errors.Is(err, nats.ErrConnectionClosed) ||
					errors.Is(err, nats.ErrBadSubscription) {
					logger.Log.Warn("[ExternalBookingWorker] Subscription/connection closed, stopping worker")
					return // ← КРИТИЧНО: прекращаем цикл
				}

				logger.Log.Error("[ExternalBookingWorker] Fetch error, retrying in 5s", "error", err.Error())
				time.Sleep(5 * time.Second)
				continue
			}

			for _, msg := range msgs {
				w.processMessage(msg)
			}
		}
	}()

	return nil
}

func (w *ExternalBookingWorker) processMessage(msg *nats.Msg) {
	var req entity.ExternalBookingRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		logger.Log.Warn("[ExternalBookingWorker] Invalid message payload", "error", err.Error())
		w.safeAck(msg, req.ExternalID) // невалидное сообщение - удаляем из очереди

		return
	}

	logger.Log.Info("[ExternalBookingWorker] Processing external booking",
		"external_id", req.ExternalID,
		"user_email", req.UserEmail,
		"slot_start", req.SlotStart,
	)

	user, err := w.userUseCase.GetUserByEmail(req.UserEmail)
	if err != nil {
		logger.Log.Error("[ExternalBookingWorker] User not found for external booking",
			"email", req.UserEmail,
			"external_id", req.ExternalID,
			"error", err.Error(),
		)
		w.safeAck(msg, req.ExternalID)

		return
	}

	createReq := &entity.CreateBookingRequest{
		SlotStart:          req.SlotStart,
		ProblemDescription: req.ProblemDescription,
	}

	_, err = w.bookingUseCase.CreateBooking(user.ID, createReq)
	if err != nil {
		// Если ошибка временная
		if isRetryable(err) {
			logger.Log.Warn("[ExternalBookingWorker] Temporary error, will retry",
				"error", err.Error(),
				"external_id", req.ExternalID,
			)
			w.safeNak(msg, req.ExternalID)

			return
		}

		logger.Log.Error("[ExternalBookingWorker] Business logic error, discarding",
			"error", err.Error(),
			"external_id", req.ExternalID,
		)
		w.safeAck(msg, req.ExternalID)

		return
	}

	logger.Log.Info("[ExternalBookingWorker] External booking created successfully",
		"external_id", req.ExternalID,
		"user_id", user.ID,
	)

	w.safeAck(msg, req.ExternalID)
}

func isRetryable(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, sql.ErrConnDone) ||
		strings.Contains(err.Error(), "connection refused")
}

func (w *ExternalBookingWorker) safeAck(msg *nats.Msg, externalID string) {
	if err := msg.Ack(); err != nil {
		logger.Log.Warn("[ExternalBookingWorker] Failed to Ack message",
			"external_id", externalID,
			"error", err.Error(),
		)
	}
}

func (w *ExternalBookingWorker) safeNak(msg *nats.Msg, externalID string) {
	if err := msg.Nak(); err != nil {
		logger.Log.Warn("[ExternalBookingWorker] Failed to Nak message",
			"external_id", externalID,
			"error", err.Error(),
		)
	}
}
