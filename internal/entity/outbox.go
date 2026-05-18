package entity

import (
	"BookingGo/internal/enum"
	"encoding/json"
	"time"
)

type OutboxEvent struct {
	ID         int               `json:"id" gorm:"primaryKey"`
	EventType  string            `json:"event_type" gorm:"not null"`
	Payload    []byte            `json:"payload" gorm:"type:jsonb;not null"`
	Status     enum.OutboxStatus `json:"status" gorm:"default:'pending'"`
	RetryCount int               `json:"retry_count"`
	CreatedAt  time.Time         `json:"created_at"`
}

func NewOutboxEvent(eventType enum.TypeOfNotification, payload any) (*OutboxEvent, error) {
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	event := &OutboxEvent{
		EventType: string(eventType),
		Payload:   payloadJson,
	}
	return event, nil
}
