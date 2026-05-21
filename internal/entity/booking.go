package entity

import (
	"BookingGo/internal/enum"
	"time"
)

type Booking struct {
	ID                 int                `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID             int                `json:"user_id" gorm:"not null"`
	SlotStart          time.Time          `json:"slot_start" gorm:"not null"`
	SlotEnd            time.Time          `json:"slot_end" gorm:"not null"`
	Status             enum.BookingStatus `json:"status" gorm:"not null"`
	ProblemDescription string             `json:"problem_description" gorm:"not null"`
	CreatedAt          time.Time          `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt          time.Time          `json:"updated_at" gorm:"not null;autoUpdateTime"`
}

type CreateBookingRequest struct {
	SlotStart          time.Time `json:"slot_start" binding:"required"`
	ProblemDescription string    `json:"problem_description"`
}
