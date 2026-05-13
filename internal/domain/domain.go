package domain

import "errors"

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrBookingNotFound           = errors.New("booking not found")
	ErrEmailTaken                = errors.New("email is taken")
	ErrInvalidEmail              = errors.New("invalid email")
	ErrInvalidPassword           = errors.New("invalid password")
	ErrCurPasswordIsNotCorrect   = errors.New("current password is not correct")
	ErrSamePassword              = errors.New("entered password is the same as current")
	ErrInvalidTimeRange          = errors.New("invalid time range")
	ErrSlotNotAvailable          = errors.New("slot not available")
	ErrInvalidToken              = errors.New("invalid token")
	ErrExpiredToken              = errors.New("token is expired")
	ErrSigningMethodNotSupported = errors.New("signing method not supported")
	ErrNotImplemented            = errors.New("not implemented")
	ErrOnlyForStaff              = errors.New("available only for stuff")
	ErrBookingNotFinished        = errors.New("booking not finished")
	ErrNotificationsAlreadyRead  = errors.New("all notifications is already read")
	ErrBookingAlreadyActive      = errors.New("booking is active right now")
)
