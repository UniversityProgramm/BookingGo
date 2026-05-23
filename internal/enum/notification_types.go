package enum

type TypeOfNotification string

const (
	NewBookingType      TypeOfNotification = "create new booking"
	CancelBookingType   TypeOfNotification = "cancel booking"
	AuthType            TypeOfNotification = "login or register"
	CompleteBookingType TypeOfNotification = "complete booking"
	ChangeEmailType     TypeOfNotification = "change email"
	ChangePasswordType  TypeOfNotification = "change password"
)
