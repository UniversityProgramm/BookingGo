package postgres

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/enum"
	"context"
	"time"

	"gorm.io/gorm"
)

type OutboxPostgres struct {
	db *gorm.DB
}

func NewOutboxPostgres(DB *gorm.DB) *OutboxPostgres {
	return &OutboxPostgres{db: DB}
}

func (r *OutboxPostgres) CreateEvent(event *domain.OutboxEvent) error {
	return r.CreateEventContext(context.Background(), event)
}

func (r *OutboxPostgres) CreateEventContext(ctx context.Context, event *domain.OutboxEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *OutboxPostgres) GetPendingEvents(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	var events []domain.OutboxEvent
	err := r.db.WithContext(ctx).Where("status = ? AND retry_count < ?", enum.StatusPending, 3).
		Order("created_at ASC").
		Limit(limit).Find(&events).Error
	return events, err
}

func (r *OutboxPostgres) MarkEventAsSent(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Model(&domain.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     enum.StatusSent,
			"updated_at": time.Now().UTC(),
		}).Error
}

func (r *OutboxPostgres) MarkEventAsFailed(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Model(&domain.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":      enum.StatusFailed,
			"retry_count": gorm.Expr("retry_count + 1"),
			"updated_at":  time.Now().UTC(),
		}).Error
}
