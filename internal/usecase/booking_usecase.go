package usecase

import (
	"BookingGo/internal/cache"
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/pkg/logger"
	"context"
	"errors"
	"fmt"
	"time"
)

//go:generate mockgen -source=booking_usecase.go -destination=mocks/booking_mocks.go -package=mocks
type BookingRepository interface {
	Create(ctx context.Context, booking *entity.Booking) error
	IsSlotAvailable(ctx context.Context, start, end time.Time) (bool, error)
	GetAllBookingsByUserID(id int) ([]entity.Booking, error)
	GetBookingByID(id int) (*entity.Booking, error)
	SetBookingComplete(bookingID int) (*entity.Booking, error)
	SetBookingCancel(bookingID, userID int) (*entity.Booking, error)
	GetAll() ([]entity.Booking, error)
}

type OutboxRepository interface {
	CreateEvent(event *entity.OutboxEvent) error
	CreateEventContext(ctx context.Context, event *entity.OutboxEvent) error
}

type BookingUseCase struct {
	bookingRepo BookingRepository
	userRepo    UserRepository
	outboxRepo  OutboxRepository
	cache       cache.Cache
}

func NewBookingUseCase(bookingRepo BookingRepository, userRepo UserRepository, outboxRepo OutboxRepository, cache cache.Cache) *BookingUseCase {
	return &BookingUseCase{bookingRepo: bookingRepo, userRepo: userRepo, outboxRepo: outboxRepo, cache: cache}
}

func (b *BookingUseCase) invalidateSlotCache(ctx context.Context, t time.Time) {
	prefix := fmt.Sprintf("slots:%s", t.Format("2006-01-02"))

	if err := b.cache.DeleteByPrefix(ctx, prefix); err != nil {
		logger.Log.Warn("[BookingUseCase] Failed to invalidate slot cache", "prefix", prefix, "error", err.Error())
	} else {
		logger.Log.Debug("[BookingUseCase] Slot cache invalidated")
	}
}

func (b *BookingUseCase) invalidateUserBookingsCache(ctx context.Context, userID int) {
	prefix := fmt.Sprintf("user:%d:bookings", userID)
	err := b.cache.Delete(ctx, prefix)
	if err != nil {
		logger.Log.Error("[BookingUseCase] Failed to invalidate user bookings cache", "error", err.Error())
	}
}

func (b *BookingUseCase) CreateBooking(ctx context.Context, id int, req *entity.CreateBookingRequest) (*entity.Booking, error) {
	user, err := b.userRepo.GetByIDContext(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if req.SlotStart.Before(now) {
		return nil, domain.ErrInvalidTimeRange
	}

	slotEnd := req.SlotStart.Add(1 * time.Hour)

	available, err := b.IsSlotAvailableCached(ctx, req.SlotStart, slotEnd)
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

	err = b.bookingRepo.Create(ctx, booking)
	if err != nil {
		return nil, err
	}

	b.invalidateSlotCache(ctx, req.SlotStart)
	b.invalidateUserBookingsCache(ctx, user.ID)

	payload := map[string]any{
		"user_id":    user.ID,
		"booking_id": booking.ID,
	}
	outboxEvent, err := entity.NewOutboxEvent(enum.NewBookingType, payload)
	if err != nil {
		logger.Log.Error("[BookingUseCase] Failed to create outboxEvent", "eventType", enum.NewBookingType, "error", err.Error())
	} else {
		err = b.outboxRepo.CreateEventContext(ctx, outboxEvent)
		if err != nil {
			logger.Log.Error("[AuthUseCase] Failed to write outboxEvent", "eventType", enum.NewBookingType, "error", err.Error())
		}
	}

	return booking, nil
}

func (b *BookingUseCase) IsSlotAvailableCached(ctx context.Context, start, end time.Time) (bool, error) {
	key := fmt.Sprintf("slots:%s:%s:%s", start.Format("2006-01-02"), start.Format("15:04"), end.Format("15:04"))

	var available bool
	err := b.cache.Get(ctx, key, &available)
	if err == nil {
		logger.Log.Debug("[BookingUseCase] Got slot cache ", "key", key)
		return available, nil
	} else if errors.Is(err, domain.ErrCacheKeyNotFound) {
		logger.Log.Debug("[BookingUseCase] Cache key not found", "key", key, "error", err.Error())
	}

	available, err = b.bookingRepo.IsSlotAvailable(ctx, start, end)
	if err != nil {
		return false, err
	}

	if err := b.cache.Set(ctx, key, available, 30*time.Second); err != nil {
		logger.Log.Warn("[BookingUseCase] Cache set failed", "error", err.Error())
	}

	return available, nil
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
	} else {
		err = b.outboxRepo.CreateEvent(outboxEvent)
		if err != nil {
			logger.Log.Error("[AuthUseCase] Failed to write outboxEvent", "eventType", enum.CompleteBookingType, "error", err.Error())
		}
	}

	return b.bookingRepo.SetBookingComplete(bookingID)
}

func (b *BookingUseCase) GetMyBookings(userID int) ([]entity.Booking, error) {
	key := fmt.Sprintf("user:%d:bookings", userID)

	var myBookings []entity.Booking
	err := b.cache.Get(context.Background(), key, &myBookings)
	if err == nil {
		logger.Log.Debug("[BookingUseCase] Got bookings for user cache ", "key", key)
		return myBookings, nil
	} else if errors.Is(err, domain.ErrCacheKeyNotFound) {
		logger.Log.Debug("[BookingUseCase] Cache key not found", "key", key, "error", err.Error())
	}

	myBookings, err = b.bookingRepo.GetAllBookingsByUserID(userID)
	if err != nil {
		return nil, err
	}

	if err := b.cache.Set(context.Background(), key, myBookings, 60*time.Second); err != nil {
		logger.Log.Warn("[BookingUseCase] Cache set failed", "error", err.Error())
	}

	return myBookings, nil
}

func (b *BookingUseCase) CancelMyBooking(bookingID, userID int) error {
	booking, err := b.bookingRepo.GetBookingByID(bookingID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if booking.SlotStart.Before(now) {
		return domain.ErrBookingIsActive
	}

	_, err = b.bookingRepo.SetBookingCancel(bookingID, userID)
	if err != nil {
		return err
	}

	b.invalidateSlotCache(context.Background(), booking.SlotStart)
	b.invalidateUserBookingsCache(context.Background(), userID)

	payload := map[string]any{
		"user_id":    userID,
		"booking_id": bookingID,
	}
	outboxEvent, err := entity.NewOutboxEvent(enum.CancelBookingType, payload)
	if err != nil {
		logger.Log.Error("[BookingUseCase] Failed to create outboxEvent", "eventType", enum.CancelBookingType, "error", err.Error())
	} else {
		err = b.outboxRepo.CreateEvent(outboxEvent)
		if err != nil {
			logger.Log.Error("[AuthUseCase] Failed to write outboxEvent", "eventType", enum.CancelBookingType, "error", err.Error())
		}
	}

	return err
}
