package dto

import "time"

type CreateBookingRequest struct {
	SlotStart          time.Time `json:"slot_start" binding:"required"`
	ProblemDescription string    `json:"problem_description"`
}
