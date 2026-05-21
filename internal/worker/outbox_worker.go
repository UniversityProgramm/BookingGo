package worker

import (
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/internal/repository"
	"BookingGo/internal/usecase"
	"BookingGo/pkg/logger"
	"encoding/json"
	"time"
)

type OutboxWorker struct {
	outboxRepo   *repository.OutboxRepository
	notifUseCase *usecase.NotificationUseCase
	interval     time.Duration
}

func NewOutboxWorker(outboxRepo *repository.OutboxRepository, notifUseCase *usecase.NotificationUseCase) *OutboxWorker {
	return &OutboxWorker{
		outboxRepo:   outboxRepo,
		notifUseCase: notifUseCase,
		interval:     5 * time.Second,
	}
}

func (w *OutboxWorker) Start() {
	logger.Log.Info("[OutboxWorker] Starting outbox worker")
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for range ticker.C {
		w.processBatch()
	}
}

func (w *OutboxWorker) processBatch() {
	events, err := w.outboxRepo.GetPendingEvents(50)
	if err != nil {
		logger.Log.Error("[OutboxWorker] Failed to get pending events", "error", err.Error())
		return
	}

	for _, event := range events {
		if err := w.processEvent(event); err != nil {
			logger.Log.Error("[OutboxWorker] Failed to process event", "eventID", event.ID, "error", err.Error())
			if err := w.outboxRepo.MarkEventAsFailed(event.ID); err != nil {
				logger.Log.Error("[OutboxWorker] Failed to mark event as failed", "eventID", event.ID, "error", err.Error())
			}
		} else {
			if err := w.outboxRepo.MarkEventAsSent(event.ID); err != nil {
				logger.Log.Error("[OutboxWorker] Failed to mark event as sent", "eventID", event.ID, "error", err.Error())
			}
		}
	}
}

func (w *OutboxWorker) processEvent(event entity.OutboxEvent) error {
	var payload map[string]any
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	params := entity.NotificationParams{}
	userID := int(payload["user_id"].(float64))
	eventType := enum.TypeOfNotification(event.EventType)

	if bookingIDRaw, ok := payload["booking_id"]; ok {
		id := int(bookingIDRaw.(float64))
		params.BookingID = id
	}

	if ipRaw, ok := payload["ip"]; ok {
		ip := ipRaw.(string)
		params.IP = ip
	}

	return w.notifUseCase.CreateNotification(userID, eventType, params)
}
