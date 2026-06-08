package repository

import (
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"context"
	"time"

	"gorm.io/gorm"
)

type OutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(DB *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: DB}
}

func (r *OutboxRepository) CreateEvent(event *entity.OutboxEvent) error {
	return r.CreateEventContext(context.Background(), event)
}

func (r *OutboxRepository) CreateEventContext(ctx context.Context, event *entity.OutboxEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *OutboxRepository) GetPendingEvents(ctx context.Context, limit int) ([]entity.OutboxEvent, error) {
	var events []entity.OutboxEvent
	err := r.db.WithContext(ctx).Where("status = ? AND retry_count < ?", enum.StatusPending, 3).
		Order("created_at ASC").
		Limit(limit).Find(&events).Error
	return events, err
}

func (r *OutboxRepository) MarkEventAsSent(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Model(&entity.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     enum.StatusSent,
			"updated_at": time.Now().UTC(),
		}).Error
}

func (r *OutboxRepository) MarkEventAsFailed(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Model(&entity.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":      enum.StatusFailed,
			"retry_count": gorm.Expr("retry_count + 1"),
			"updated_at":  time.Now().UTC(),
		}).Error
}
