package entity

import "time"

type Booking struct {
	ID                 int       `json:"id" db:"id"`
	UserID             int       `json:"userId" db:"user_id"`
	SlotStart          time.Time `json:"slotStart" db:"slot_start"`
	Status             string    `json:"status" db:"status"`
	ProblemDescription string    `json:"problem" db:"problem"`
	CreatedAt          time.Time `json:"createdAt" db:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt" db:"updatedAt"`
}
