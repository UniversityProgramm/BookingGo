package entity

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

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=20"`
	FIO      string `json:"fio" binding:"required"`
	Phone    string `json:"phone" binding:"required,min=10,max=15"`
}

type UpdateUserRequest struct {
	FIO   *string `json:"fio"`
	Phone *string `json:"phone,min=10,max=15"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"cur_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=20"`
	OtpCode         string `json:"otp_code" binding:"required,len=6"`
}

type ChangeEmailRequest struct {
	NewEmail        string `json:"new_email" binding:"required,email"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	OtpCode         string `json:"otp_code" binding:"required,len=6"`
}

type VerifyTotpRequest struct {
	OtpCode string `json:"otp_code" binding:"required,len=6"`
}

type DisableTotpRequest struct {
	OtpCode         string `json:"otp_code" binding:"required,len=6"`
	CurrentPassword string `json:"confirm_password" binding:"required"`
}
