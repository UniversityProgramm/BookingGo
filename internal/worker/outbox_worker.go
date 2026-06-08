package worker

import (
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/internal/repository"
	"BookingGo/internal/usecase"
	"BookingGo/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type OutboxWorker struct {
	outboxRepo   *repository.OutboxRepository
	notifUseCase *usecase.NotificationUseCase
	interval     time.Duration

	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running atomic.Bool
}

func NewOutboxWorker(outboxRepo *repository.OutboxRepository, notifUseCase *usecase.NotificationUseCase) *OutboxWorker {
	return &OutboxWorker{
		outboxRepo:   outboxRepo,
		notifUseCase: notifUseCase,
		interval:     5 * time.Second,
	}
}

func (w *OutboxWorker) Start() error {
	if !w.running.CompareAndSwap(false, true) {
		return fmt.Errorf("OutboxWorker is already running")
	}

	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.wg.Add(1)

	go w.Run()

	logger.Log.Info("[OutboxWorker] Started")
	return nil
}

func (w *OutboxWorker) Run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			logger.Log.Info("[OutboxWorker] Stopping...")
			return
		case <-ticker.C:
			w.processBatch()
		}
	}
}

func (w *OutboxWorker) processBatch() {
	events, err := w.outboxRepo.GetPendingEvents(w.ctx, 50)
	if err != nil {
		if w.ctx.Err() != nil {
			return
		}
		logger.Log.Error("[OutboxWorker] Failed to get pending events", "error", err.Error())
		return
	}

	for _, event := range events {
		if w.ctx.Err() != nil {
			return
		}

		if err := w.processEvent(event); err != nil {
			logger.Log.Warn("[OutboxWorker] Failed to process event", "eventID", event.ID, "error", err.Error())
			if err := w.outboxRepo.MarkEventAsFailed(w.ctx, event.ID); err != nil {
				logger.Log.Error("[OutboxWorker] Failed to mark event as failed", "eventID", event.ID, "error", err.Error())
			}
		} else {
			if err := w.outboxRepo.MarkEventAsSent(w.ctx, event.ID); err != nil {
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

func (w *OutboxWorker) Stop() {
	if !w.running.CompareAndSwap(true, false) {
		return
	}

	w.cancel()
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Log.Info("[OutboxWorker] Stopped")
	case <-time.After(10 * time.Second):
		logger.Log.Warn("[OutboxWorker] Stop timeout, forcing shutdown")
	}
}
