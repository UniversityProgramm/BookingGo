package entity

import (
	"BookingGo/internal/enum"
	"time"
)

type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FIO          string    `json:"fio" db:"fio"`
	Phone        string    `json:"phone" db:"phone"`
	Role         enum.Role `json:"role" db:"role"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
	IsActive     bool      `json:"isActive" db:"is_active"`
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=20"`
	FIO      string `json:"fio" binding:"required"`
	Phone    string `json:"phone" binding:"required,min=10,max=15"`
}

type UpdateUserRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
	FIO      *string `json:"fio"`
	Phone    *string `json:"phone"`
}
