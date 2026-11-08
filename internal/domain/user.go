package domain

import (
	"BookingGo/internal/enum"
	"time"
)

type User struct {
	ID                       int                  `json:"id" gorm:"primaryKey;autoIncrement" `
	Email                    string               `json:"email" gorm:"uniqueIndex;not null" `
	PasswordHash             string               `json:"-" gorm:"not null"`
	FIO                      string               `json:"fio" `
	Phone                    string               `json:"phone" gorm:"not null"`
	Role                     enum.Role            `json:"role" gorm:"not null"`
	CreatedAt                time.Time            `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt                time.Time            `json:"updated_at" gorm:"not null;autoUpdateTime"`
	IsActive                 bool                 `json:"is_active"`
	TotpSecret               string               `json:"-" gorm:"column:totp_secret"`
	IsTotpEnabled            bool                 `json:"is_totp_enabled" gorm:"default:false"`
	UserNotificationSettings NotificationSettings `json:"notification_settings" gorm:"type:json;serializer:json;column:notification_settings;not null"`
}
