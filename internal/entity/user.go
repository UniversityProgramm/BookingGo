package entity

import (
	"BookingGo/internal/enum"
	"time"
)

type User struct {
	ID        int       `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"password" db:"password"`
	FIO       string    `json:"fio" db:"fio"`
	Phone     string    `json:"phone" db:"phone"`
	Role      enum.Role `json:"role" db:"role"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	FIO      string `json:"fio" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
}

type UpdateUserRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
	FIO      *string `json:"fio"`
	Phone    *string `json:"phone"`
}
