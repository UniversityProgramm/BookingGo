package usecase

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"log"
	"time"
)

type BookingRepository interface {
	Create(booking *entity.Booking) error
	IsSlotAvailable(start, end time.Time) (bool, error)
	GetAllBookingsByUserID(id int) ([]entity.Booking, error)
	GetBookingByID(id int) (*entity.Booking, error)
	SetBookingComplete(bookingID int) (*entity.Booking, error)
	SetBookingCancel(bookingID, userID int) (*entity.Booking, error)
}

type BookingNotifier interface {
	CreateNotification(userID int, notificationType enum.TypeOfNotification, params entity.NotificationParams) error
}

type BookingUseCase struct {
	bookingRepo BookingRepository
	userRepo    UserRepository
	notifier    BookingNotifier
}

func NewBookingUseCase(bookingRepo BookingRepository, userRepo UserRepository, notifier BookingNotifier) *BookingUseCase {
	return &BookingUseCase{bookingRepo: bookingRepo, userRepo: userRepo, notifier: notifier}
}

func (b *BookingUseCase) SendBookingNotification(userID int, bookingType enum.TypeOfNotification, params entity.NotificationParams) {
	err := b.notifier.CreateNotification(userID, bookingType, params)
	if err != nil {
		log.Printf("[Notifications] Ошибка при создании уведомления: %v", err.Error())
	}
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

	params := entity.NotificationParams{
		BookingID: booking.ID,
	}
	b.SendBookingNotification(user.ID, enum.NewBookingType, params)

	return booking, nil
}

func (b *BookingUseCase) CompleteBookingByID(bookingID int, userID int) (*entity.Booking, error) {
	user, err := b.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user.Role != enum.RoleStaff {
		return nil, domain.ErrOnlyForStaff
	}

	booking, err := b.bookingRepo.GetBookingByID(bookingID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if booking.SlotEnd.After(now) {
		return nil, domain.ErrBookingNotFinished
	}

	params := entity.NotificationParams{
		BookingID: booking.ID,
	}
	b.SendBookingNotification(booking.UserID, enum.CompleteBookingType, params)

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

	params := entity.NotificationParams{
		BookingID: booking.ID,
	}
	b.SendBookingNotification(userID, enum.CancelBookingType, params)

	return err
}
