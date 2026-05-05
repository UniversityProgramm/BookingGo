package entity

import (
	"BookingGo/internal/enum"
	"time"
)

type User struct {
	ID           int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	FIO          string    `json:"fio"`
	Phone        string    `gorm:"not null" json:"phone"`
	Role         enum.Role `gorm:"not null" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsActive     bool      `json:"is_active"`
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=20"`
	FIO      string `json:"fio" binding:"required"`
	Phone    string `json:"phone" binding:"required,min=10,max=15"`
}

type UpdateUserRequest struct {
	FIO   *string `json:"fio"`
	Phone *string `json:"phone"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"cur_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=20"`
}
