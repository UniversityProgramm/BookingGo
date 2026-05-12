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

type UserNotificationSettings struct {
	UserID       int  `json:"user_id" gorm:"primaryKey"`
	EmailEnabled bool `json:"email_enabled" gorm:"default:false;not null"`
	SMSEnabled   bool `json:"sms_enabled" gorm:"default:false;not null"`
}

type NotificationRequestEmail struct {
}

type NotificationRequestSMS struct{}
