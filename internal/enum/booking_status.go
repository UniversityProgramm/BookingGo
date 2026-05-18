package enum

type BookingStatus string

const (
	StatusConfirmed BookingStatus = "confirmed"
	StatusCancelled BookingStatus = "cancelled"
	StatusCompleted BookingStatus = "completed"
)
