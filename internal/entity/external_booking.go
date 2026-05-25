package entity

import (
	"time"
)

type ExternalBookingRequest struct {
	UserEmail          string    `json:"user_email"`
	SlotStart          time.Time `json:"slot_start" binding:"required"`
	ProblemDescription string    `json:"problem_description"`
	ExternalID         string    `json:"external_id"`
}
