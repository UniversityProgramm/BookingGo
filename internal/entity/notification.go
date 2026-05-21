package entity

import (
	"time"
)

type Notification struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	RecipientID int       `json:"-" gorm:"not null"`
	Title       string    `json:"title" gorm:"not null"`
	Body        string    `json:"body" gorm:"not null"`
	IsRead      bool      `json:"is_read" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
}

type NotificationSettings struct {
	IsEmailSend bool `json:"is_email" gorm:"not null"`
	IsPhoneSend bool `json:"is_sms" gorm:"not null"`
}

type NotificationParams struct {
	BookingID int    `json:"booking_id,omitempty"`
	IP        string `json:"ip,omitempty"`
}
