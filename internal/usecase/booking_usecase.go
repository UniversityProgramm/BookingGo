package usecase

import (
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/internal/repository"
	"errors"
	"time"
)

var (
	ErrOnlyForClient    = errors.New("available only for client")
	ErrInvalidTimeRange = errors.New("invalid time range")
	ErrSlotNotAvailable = errors.New("slot not available")
	ErrBookingNotFound  = errors.New("booking not found")
)

type BookingRepository interface {
	Create(booking *entity.Booking) error
	IsSlotAvailable(start, end time.Time) (bool, error)
	GetBookingsByUserID(id int) ([]entity.Booking, error)
	GetBookingByID(id int) (*entity.Booking, error)
	DeleteBookingByID(bookingID, userID int) error
}

type BookingUsecase struct {
	bookingRepo BookingRepository
	userRepo    UserRepository
}

func NewBookingUsecase(bookingRepo BookingRepository, userRepo UserRepository) *BookingUsecase {
	return &BookingUsecase{bookingRepo: bookingRepo, userRepo: userRepo}
}

func (b *BookingUsecase) CreateBooking(id int, req *entity.CreateBookingRequest) (*entity.Booking, error) {
	user, err := b.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user.Role != enum.RoleClient {
		return nil, ErrOnlyForClient
	}

	now := time.Now()
	if req.SlotStart.Before(now) {
		return nil, ErrInvalidTimeRange
	}

	slotEnd := req.SlotStart.Add(1 * time.Hour)

	available, err := b.bookingRepo.IsSlotAvailable(req.SlotStart, slotEnd)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, ErrSlotNotAvailable
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

	return booking, nil
}

func (b *BookingUsecase) GetMyBookings(id int) ([]entity.Booking, error) {
	return b.bookingRepo.GetBookingsByUserID(id)
}

func (b *BookingUsecase) DeleteMyBooking(bookingID, userID int) error {
	err := b.bookingRepo.DeleteBookingByID(bookingID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return ErrBookingNotFound
		}
		return err
	}

	return nil
}
