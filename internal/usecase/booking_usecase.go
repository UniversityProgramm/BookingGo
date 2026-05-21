package usecase

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/pkg/logger"
	"time"
)

type BookingRepository interface {
	Create(booking *entity.Booking) error
	IsSlotAvailable(start, end time.Time) (bool, error)
	GetAllBookingsByUserID(id int) ([]entity.Booking, error)
	GetBookingByID(id int) (*entity.Booking, error)
	SetBookingComplete(bookingID int) (*entity.Booking, error)
	SetBookingCancel(bookingID, userID int) (*entity.Booking, error)
	GetAll() ([]entity.Booking, error)
}

type OutboxBookingRepository interface {
	CreateEvent(event *entity.OutboxEvent) error
}

type BookingUseCase struct {
	bookingRepo BookingRepository
	userRepo    UserRepository
	outboxRepo  OutboxBookingRepository
}

func NewBookingUseCase(bookingRepo BookingRepository, userRepo UserRepository, outboxRepo OutboxBookingRepository) *BookingUseCase {
	return &BookingUseCase{bookingRepo: bookingRepo, userRepo: userRepo, outboxRepo: outboxRepo}
}

func (b *BookingUseCase) CreateBooking(id int, req *entity.CreateBookingRequest) (*entity.Booking, error) {
	user, err := b.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if req.SlotStart.Before(now) {
		return nil, domain.ErrInvalidTimeRange
	}

	slotEnd := req.SlotStart.Add(1 * time.Hour)

	available, err := b.bookingRepo.IsSlotAvailable(req.SlotStart, slotEnd)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, domain.ErrSlotNotAvailable
	}

	booking := &entity.Booking{
		UserID:             user.ID,
		SlotStart:          req.SlotStart,
		SlotEnd:            slotEnd,
		Status:             enum.StatusConfirmed,
		ProblemDescription: req.ProblemDescription,
	}

	err = b.bookingRepo.Create(booking)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"user_id":    user.ID,
		"booking_id": booking.ID,
	}
	outboxEvent, err := entity.NewOutboxEvent(enum.NewBookingType, payload)
	if err != nil {
		logger.Log.Error("[BookingUseCase] Failed to create outboxEvent", "eventType", enum.NewBookingType, "error", err.Error())
	}

	err = b.outboxRepo.CreateEvent(outboxEvent)
	if err != nil {
		logger.Log.Error("[AuthUseCase] Failed to write outboxEvent", "eventType", enum.NewBookingType, "error", err.Error())
	}

	return booking, nil
}

func (b *BookingUseCase) GetAllBookings(userID int) ([]entity.Booking, error) {
	user, err := b.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user.Role == enum.RoleClient {
		return nil, domain.ErrOnlyForStaffOrAdmin
	}
	return b.bookingRepo.GetAll()
}

func (b *BookingUseCase) CompleteBookingByID(bookingID int) (*entity.Booking, error) {
	booking, err := b.bookingRepo.GetBookingByID(bookingID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if booking.SlotEnd.After(now) {
		return nil, domain.ErrBookingNotFinished
	}

	payload := map[string]any{
		"user_id":    booking.UserID,
		"booking_id": booking.ID,
	}
	outboxEvent, err := entity.NewOutboxEvent(enum.CompleteBookingType, payload)
	if err != nil {
		logger.Log.Error("[BookingUseCase] Failed to create outboxEvent", "eventType", enum.CompleteBookingType, "error", err.Error())
	}

	err = b.outboxRepo.CreateEvent(outboxEvent)
	if err != nil {
		logger.Log.Error("[AuthUseCase] Failed to write outboxEvent", "eventType", enum.CompleteBookingType, "error", err.Error())
	}

	return b.bookingRepo.SetBookingComplete(bookingID)
}

func (b *BookingUseCase) GetMyBookings(id int) ([]entity.Booking, error) {
	return b.bookingRepo.GetAllBookingsByUserID(id)
}

func (b *BookingUseCase) CancelMyBooking(bookingID, userID int) error {
	booking, err := b.bookingRepo.GetBookingByID(bookingID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if booking.SlotStart.Before(now) {
		return domain.ErrBookingAlreadyActive
	}

	_, err = b.bookingRepo.SetBookingCancel(bookingID, userID)

	payload := map[string]any{
		"user_id":    userID,
		"booking_id": bookingID,
	}
	outboxEvent, err := entity.NewOutboxEvent(enum.CancelBookingType, payload)
	if err != nil {
		logger.Log.Error("[BookingUseCase] Failed to create outboxEvent", "eventType", enum.CancelBookingType, "error", err.Error())
	}

	err = b.outboxRepo.CreateEvent(outboxEvent)
	if err != nil {
		logger.Log.Error("[AuthUseCase] Failed to write outboxEvent", "eventType", enum.CancelBookingType, "error", err.Error())
	}

	return err
}
