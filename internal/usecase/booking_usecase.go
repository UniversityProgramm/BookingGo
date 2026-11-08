package usecase

import (
	"BookingGo/internal/adapter/redis"
	"BookingGo/internal/domain"
	"BookingGo/internal/dto"
	"BookingGo/internal/enum"
	"BookingGo/pkg/logger"
	"context"
	"errors"
	"fmt"
	"time"
)

//go:generate mockgen -source=booking_usecase.go -destination=mocks/booking_mocks.go -package=mocks
type BookingPostgres interface {
	Create(ctx context.Context, booking *domain.Booking) error
	IsSlotAvailable(ctx context.Context, start, end time.Time) (bool, error)
	GetAllBookingsByUserID(id int) ([]domain.Booking, error)
	GetBookingByID(id int) (*domain.Booking, error)
	SetBookingComplete(bookingID int) (*domain.Booking, error)
	SetBookingCancel(bookingID, userID int) (*domain.Booking, error)
	GetAll() ([]domain.Booking, error)
}

type OutboxPostgres interface {
	CreateEvent(event *domain.OutboxEvent) error
	CreateEventContext(ctx context.Context, event *domain.OutboxEvent) error
}

type BookingUseCase struct {
	bookingPostgres BookingPostgres
	userPostgres    UserPostgres
	outboxPostgres  OutboxPostgres
	cache           redis.Cache
}

func NewBookingUseCase(bookingPostgres BookingPostgres, userPostgres UserPostgres, outboxPostgres OutboxPostgres, cache redis.Cache) *BookingUseCase {
	return &BookingUseCase{bookingPostgres: bookingPostgres, userPostgres: userPostgres, outboxPostgres: outboxPostgres, cache: cache}
}

func (b *BookingUseCase) invalidateSlotCache(ctx context.Context, t time.Time) {
	prefix := fmt.Sprintf("slots:%s", t.Format("2006-01-02"))

	if err := b.cache.DeleteByPrefix(ctx, prefix); err != nil {
		logger.Log.Warn("[BookingUseCase] Failed to invalidate slot redis", "prefix", prefix, "error", err.Error())
	} else {
		logger.Log.Debug("[BookingUseCase] Slot redis invalidated")
	}
}

func (b *BookingUseCase) invalidateUserBookingsCache(ctx context.Context, userID int) {
	key := fmt.Sprintf("user:%d:bookings", userID)
	err := b.cache.Delete(ctx, key)
	if err != nil {
		logger.Log.Error("[BookingUseCase] Failed to invalidate user bookings redis", "error", err.Error())
	}
}

func (b *BookingUseCase) CreateBooking(ctx context.Context, id int, req *dto.CreateBookingRequest) (*domain.Booking, error) {
	user, err := b.userPostgres.GetByIDContext(ctx, id)
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

	booking := &domain.Booking{
		UserID:             user.ID,
		SlotStart:          req.SlotStart,
		SlotEnd:            slotEnd,
		Status:             enum.StatusConfirmed,
		ProblemDescription: req.ProblemDescription,
	}

	err = b.bookingPostgres.Create(ctx, booking)
	if err != nil {
		return nil, err
	}

	b.invalidateSlotCache(ctx, req.SlotStart)
	b.invalidateUserBookingsCache(ctx, user.ID)

	payload := map[string]any{
		"user_id":    user.ID,
		"booking_id": booking.ID,
	}
	outboxEvent, err := domain.NewOutboxEvent(enum.NewBookingType, payload)
	if err != nil {
		logger.Log.Error("[BookingUseCase] Failed to create outboxEvent", "eventType", enum.NewBookingType, "error", err.Error())
	} else {
		err = b.outboxPostgres.CreateEventContext(ctx, outboxEvent)
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
		logger.Log.Debug("[BookingUseCase] Got slot redis ", "key", key)
		return available, nil
	} else if errors.Is(err, domain.ErrCacheKeyNotFound) {
		logger.Log.Debug("[BookingUseCase] Cache key not found", "key", key, "error", err.Error())
	}

	available, err = b.bookingPostgres.IsSlotAvailable(ctx, start, end)
	if err != nil {
		return false, err
	}

	if err := b.cache.Set(ctx, key, available, 30*time.Second); err != nil {
		logger.Log.Warn("[BookingUseCase] Cache set failed", "error", err.Error())
	}

	return available, nil
}

func (b *BookingUseCase) GetAllBookings(userID int) ([]domain.Booking, error) {
	user, err := b.userPostgres.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user.Role == enum.RoleClient {
		return nil, domain.ErrOnlyForStaffOrAdmin
	}
	return b.bookingPostgres.GetAll()
}

func (b *BookingUseCase) CompleteBookingByID(bookingID int) (*domain.Booking, error) {
	booking, err := b.bookingPostgres.GetBookingByID(bookingID)
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
	outboxEvent, err := domain.NewOutboxEvent(enum.CompleteBookingType, payload)
	if err != nil {
		logger.Log.Error("[BookingUseCase] Failed to create outboxEvent", "eventType", enum.CompleteBookingType, "error", err.Error())
	} else {
		err = b.outboxPostgres.CreateEvent(outboxEvent)
		if err != nil {
			logger.Log.Error("[AuthUseCase] Failed to write outboxEvent", "eventType", enum.CompleteBookingType, "error", err.Error())
		}
	}

	return b.bookingPostgres.SetBookingComplete(bookingID)
}

func (b *BookingUseCase) GetMyBookings(userID int) ([]domain.Booking, error) {
	key := fmt.Sprintf("user:%d:bookings", userID)

	var myBookings []domain.Booking
	err := b.cache.Get(context.Background(), key, &myBookings)
	if err == nil {
		logger.Log.Debug("[BookingUseCase] Got bookings redis for user", "key", key)
		return myBookings, nil
	} else if errors.Is(err, domain.ErrCacheKeyNotFound) {
		logger.Log.Debug("[BookingUseCase] Cache key not found", "key", key, "error", err.Error())
	}

	myBookings, err = b.bookingPostgres.GetAllBookingsByUserID(userID)
	if err != nil {
		return nil, err
	}

	if err := b.cache.Set(context.Background(), key, myBookings, 60*time.Second); err != nil {
		logger.Log.Warn("[BookingUseCase] Cache set failed", "error", err.Error())
	}

	return myBookings, nil
}

func (b *BookingUseCase) CancelMyBooking(bookingID, userID int) error {
	booking, err := b.bookingPostgres.GetBookingByID(bookingID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if booking.SlotStart.Before(now) {
		return domain.ErrBookingIsActive
	}

	_, err = b.bookingPostgres.SetBookingCancel(bookingID, userID)
	if err != nil {
		return err
	}

	b.invalidateSlotCache(context.Background(), booking.SlotStart)
	b.invalidateUserBookingsCache(context.Background(), userID)

	payload := map[string]any{
		"user_id":    userID,
		"booking_id": bookingID,
	}
	outboxEvent, err := domain.NewOutboxEvent(enum.CancelBookingType, payload)
	if err != nil {
		logger.Log.Error("[BookingUseCase] Failed to create outboxEvent", "eventType", enum.CancelBookingType, "error", err.Error())
	} else {
		err = b.outboxPostgres.CreateEvent(outboxEvent)
		if err != nil {
			logger.Log.Error("[AuthUseCase] Failed to write outboxEvent", "eventType", enum.CancelBookingType, "error", err.Error())
		}
	}

	return err
}
