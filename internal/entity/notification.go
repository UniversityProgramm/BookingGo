package entity

import "time"

type Notification struct {
	ID          int       `json:"id" db:"id"`
	BookingID   int       `json:"bookingId" db:"bookingId"`
	RecipientID int       `json:"recipientId" db:"recipientId"`
	IsRead      bool      `json:"isRead" db:"isRead"`
	CreatedAt   time.Time `json:"createdAt" db:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updatedAt"`
}
