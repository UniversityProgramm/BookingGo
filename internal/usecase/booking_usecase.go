package usecase

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"time"
)

type BookingRepository interface {
	Create(booking *entity.Booking) error
	IsSlotAvailable(start, end time.Time) (bool, error)
	GetBookingsByUserID(id int) ([]entity.Booking, error)
	GetBookingByID(id int) (*entity.Booking, error)
	DeleteBookingByID(bookingID, userID int) error
}

type BookingUseCase struct {
	bookingRepo BookingRepository
	userRepo    UserRepository
}

func NewBookingUseCase(bookingRepo BookingRepository, userRepo UserRepository) *BookingUseCase {
	return &BookingUseCase{bookingRepo: bookingRepo, userRepo: userRepo}
}

func (b *BookingUseCase) CreateBooking(id int, req *entity.CreateBookingRequest) (*entity.Booking, error) {
	user, err := b.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user.Role != enum.RoleClient {
		return nil, domain.ErrOnlyForClient
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

	return booking, nil
}

func (b *BookingUseCase) GetMyBookings(id int) ([]entity.Booking, error) {
	return b.bookingRepo.GetBookingsByUserID(id)
}

func (b *BookingUseCase) DeleteMyBooking(bookingID, userID int) error {
	return b.bookingRepo.DeleteBookingByID(bookingID, userID)
}
