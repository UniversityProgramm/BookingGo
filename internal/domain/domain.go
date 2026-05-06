package domain

import "errors"

var (
	ErrUserNotFound            = errors.New("user not found")
	ErrBookingNotFound         = errors.New("booking not found")
	ErrEmailTaken              = errors.New("email is taken")
	ErrInvalidEmail            = errors.New("invalid email")
	ErrInvalidPassword         = errors.New("invalid password")
	ErrCurPasswordIsNotCorrect = errors.New("current password is not correct")
	ErrSamePassword            = errors.New("entered password is the same as current")
	ErrOnlyForClient           = errors.New("available only for client")
	ErrInvalidTimeRange        = errors.New("invalid time range")
	ErrSlotNotAvailable        = errors.New("slot not available")
)
