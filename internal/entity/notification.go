package entity

import (
	"BookingGo/internal/enum"
	"time"
)

type Notification struct {
	ID          int          `json:"id" gorm:"primaryKey;autoIncrement"`
	BookingID   int          `json:"bookingId" gorm:"not null"`
	RecipientID int          `json:"recipientId" gorm:"not null"`
	IsRead      bool         `json:"isRead" gorm:"not null"`
	Channel     enum.Channel `json:"channel" gorm:"not null"`
	CreatedAt   time.Time    `json:"createdAt" gorm:"not null;autoCreateTime"`
	UpdatedAt   time.Time    `json:"updatedAt" gorm:"not null;autoUpdateTime"`
}

type NotificationRequestEmail struct {
}

type NotificationRequestSMS struct{}
