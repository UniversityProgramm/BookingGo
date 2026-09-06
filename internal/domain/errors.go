package domain

import "errors"

// User
var (
	ErrUserNotFound            = errors.New("user not found")
	ErrEmailTaken              = errors.New("email is taken")
	ErrInvalidEmail            = errors.New("invalid email")
	ErrInvalidPassword         = errors.New("invalid password")
	ErrCurPasswordIsNotCorrect = errors.New("current password is not correct")
	ErrSamePassword            = errors.New("entered password is the same as current")
)

// Booking
var (
	ErrBookingNotFound    = errors.New("booking not found")
	ErrInvalidTimeRange   = errors.New("invalid time range")
	ErrSlotNotAvailable   = errors.New("slot not available")
	ErrBookingNotFinished = errors.New("booking not finished")
	ErrBookingIsActive    = errors.New("booking is active right now")
)

// Auth
var (
	ErrInvalidToken              = errors.New("invalid token")
	ErrExpiredToken              = errors.New("token is expired")
	ErrSigningMethodNotSupported = errors.New("signing method not supported")
)

// TOTP / 2FA
var (
	ErrTotpAlreadyEnabled  = errors.New("totp already enabled")
	ErrTotpSecretNotSet    = errors.New("totp secret not set")
	ErrInvalidTotpCode     = errors.New("invalid otp secret")
	ErrTotpAlreadyDisabled = errors.New("totp already disabled")
)

// Notifications
var (
	ErrNotificationsAlreadyRead = errors.New("all notifications is already read")
)

// Middleware
var (
	ErrOnlyForStaffOrAdmin = errors.New("only for staff or admin")
)

// NATS / JetStream
var (
	ErrNatsConnection      = errors.New("natsClient connection error")
	ErrJetStreamConnection = errors.New("jetstream connection error")
	ErrStreamAdd           = errors.New("stream add error")
	ErrStreamAlreadyExists = errors.New("stream already exists")
)

// Cache
var (
	ErrCacheKeyNotFound = errors.New("cache key not found")
)

// Stubs
var (
	ErrNotImplemented = errors.New("not implemented")
)
