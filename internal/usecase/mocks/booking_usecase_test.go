package mocks

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/internal/usecase"
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

var (
	testClient = &entity.User{
		ID:       1,
		Email:    "client@example.com",
		Role:     enum.RoleClient,
		IsActive: true,
		UserNotificationSettings: entity.NotificationSettings{
			IsEmailSend: true,
			IsPhoneSend: false,
		},
	}

	testStaff = &entity.User{
		ID:       3,
		Email:    "staff@mail.ru",
		Role:     enum.RoleStaff,
		IsActive: true,
		UserNotificationSettings: entity.NotificationSettings{
			IsEmailSend: true,
			IsPhoneSend: false,
		},
	}

	testBooking = &entity.Booking{
		ID:                 10,
		UserID:             1,
		SlotStart:          time.Now().Add(time.Hour),
		SlotEnd:            time.Now().Add(2 * time.Hour),
		Status:             enum.StatusConfirmed,
		ProblemDescription: "problem",
	}

	testBookingPast = &entity.Booking{
		ID:                 11,
		UserID:             1,
		SlotStart:          time.Now().Add(-2 * time.Hour),
		SlotEnd:            time.Now().Add(-1 * time.Hour),
		Status:             enum.StatusConfirmed,
		ProblemDescription: "problem",
	}
)

func TestBookingUseCase_CreateBooking_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	futureTime := time.Now().Add(24 * time.Hour)
	ctx := context.Background()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockUserRepo.EXPECT().
		GetByIDContext(ctx, testClient.ID).
		Return(testClient, nil)

	mockCache.EXPECT().
		Get(ctx, gomock.Any(), gomock.Any()).
		Return(domain.ErrCacheKeyNotFound)

	mockBookingRepo.EXPECT().
		IsSlotAvailable(ctx, futureTime, gomock.Any()).
		Return(true, nil)

	mockCache.EXPECT().
		Set(ctx, gomock.Any(), true, 30*time.Second).
		Return(nil)

	mockBookingRepo.EXPECT().
		Create(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, b *entity.Booking) error {
			b.ID = 100
			return nil
		})

	mockCache.EXPECT().
		DeleteByPrefix(ctx, gomock.Any()).
		Return(nil)

	mockCache.EXPECT().
		Delete(ctx, gomock.Any()).
		Return(nil)

	mockOutboxRepo.EXPECT().
		CreateEventContext(ctx, gomock.Cond(func(e *entity.OutboxEvent) bool {
			if e.EventType != string(enum.NewBookingType) {
				t.Errorf("Expected event type %s, got %s", enum.NewBookingType, e.EventType)
			}
			return true
		})).
		Return(nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	req := &entity.CreateBookingRequest{
		SlotStart:          futureTime,
		ProblemDescription: "Test problem",
	}

	booking, err := useCase.CreateBooking(ctx, testClient.ID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if booking == nil {
		t.Error("Booking should not be nil")
	}
	if booking.ID != 100 {
		t.Errorf("Expected booking ID 100, got %d", booking.ID)
	}
}

func TestBookingUseCase_CreateBooking_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	futureTime := time.Now().Add(24 * time.Hour)
	ctx := context.Background()
	slotCheckedInDB := false

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockUserRepo.EXPECT().
		GetByIDContext(ctx, testClient.ID).
		Return(testClient, nil)

	mockCache.EXPECT().
		Get(ctx, gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, dest any) error {
			if boolPtr, ok := dest.(*bool); ok {
				*boolPtr = true
			}
			return nil
		})

	mockBookingRepo.EXPECT().
		IsSlotAvailable(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, start, end time.Time) (bool, error) {
			slotCheckedInDB = true
			return true, nil
		}).
		MaxTimes(0)

	mockBookingRepo.EXPECT().
		Create(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, b *entity.Booking) error {
			b.ID = 100
			return nil
		})

	mockCache.EXPECT().DeleteByPrefix(ctx, gomock.Any()).Return(nil)
	mockCache.EXPECT().Delete(ctx, gomock.Any()).Return(nil)

	mockOutboxRepo.EXPECT().CreateEventContext(ctx, gomock.Any()).Return(nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	req := &entity.CreateBookingRequest{
		SlotStart:          futureTime,
		ProblemDescription: "Test cache hit",
	}

	_, err := useCase.CreateBooking(ctx, testClient.ID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if slotCheckedInDB {
		t.Error("IsSlotAvailable should not be called on cache hit")
	}
}

func TestBookingUseCase_CreateBooking_PastTimeRange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pastTime := time.Now().Add(-24 * time.Hour)
	ctx := context.Background()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockUserRepo.EXPECT().
		GetByIDContext(ctx, testClient.ID).
		Return(testClient, nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	req := &entity.CreateBookingRequest{
		SlotStart:          pastTime,
		ProblemDescription: "Test problem",
	}

	_, err := useCase.CreateBooking(ctx, testClient.ID, req)

	if !errors.Is(err, domain.ErrInvalidTimeRange) {
		t.Errorf("Expected ErrInvalidTimeRange, got error: %v", err)
	}

}

func TestBookingUseCase_CreateBooking_SlotNotAvailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	futureTime := time.Now().Add(24 * time.Hour)
	ctx := context.Background()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockUserRepo.EXPECT().
		GetByIDContext(ctx, testClient.ID).
		Return(testClient, nil)

	mockCache.EXPECT().
		Get(ctx, gomock.Any(), gomock.Any()).
		Return(domain.ErrCacheKeyNotFound)

	mockBookingRepo.EXPECT().
		IsSlotAvailable(ctx, futureTime, gomock.Any()).
		Return(false, nil)

	mockCache.EXPECT().
		Set(ctx, gomock.Any(), false, 30*time.Second).
		Return(nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	req := &entity.CreateBookingRequest{
		SlotStart:          futureTime,
		ProblemDescription: "Test problem",
	}

	_, err := useCase.CreateBooking(ctx, testClient.ID, req)

	if !errors.Is(err, domain.ErrSlotNotAvailable) {
		t.Errorf("Expected ErrSlotNotAvailable, got error: %v", err)
	}

}

func TestBookingUseCase_CreateBooking_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockUserRepo.EXPECT().
		GetByIDContext(ctx, testClient.ID).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	req := &entity.CreateBookingRequest{
		SlotStart:          time.Now().Add(24 * time.Hour),
		ProblemDescription: "Test problem",
	}

	_, err := useCase.CreateBooking(ctx, testClient.ID, req)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got error: %v", err)
	}
}

func TestBookingUseCase_GetAllBookings_Staff_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockUserRepo.EXPECT().
		GetByID(testStaff.ID).
		Return(testStaff, nil)

	expectedBookings := []entity.Booking{*testBooking}
	mockBookingRepo.EXPECT().
		GetAll().
		Return(expectedBookings, nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	_, err := useCase.GetAllBookings(testStaff.ID)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestBookingUseCase_GetAllBookings_Client_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockUserRepo.EXPECT().
		GetByID(testClient.ID).
		Return(testClient, nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	_, err := useCase.GetAllBookings(testClient.ID)

	if !errors.Is(err, domain.ErrOnlyForStaffOrAdmin) {
		t.Errorf("Expected ErrOnlyForStaffOrAdmin, got: %v", err)
	}
}

func TestBookingUseCase_GetAllBookings_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockUserRepo.EXPECT().
		GetByID(testStaff.ID).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	_, err := useCase.GetAllBookings(testStaff.ID)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got error: %v", err)
	}
}

func TestBookingUseCase_CompleteBookingByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockBookingRepo.EXPECT().
		GetBookingByID(testBookingPast.ID).
		Return(testBookingPast, nil)

	mockBookingRepo.EXPECT().
		SetBookingComplete(testBookingPast.ID).
		Return(&entity.Booking{
			ID:     testBookingPast.ID,
			Status: enum.StatusCompleted,
		}, nil)

	mockOutboxRepo.EXPECT().
		CreateEvent(gomock.Cond(func(e *entity.OutboxEvent) bool {
			if e.EventType != string(enum.CompleteBookingType) {
				t.Errorf("Expected event type %s, got %s", enum.CompleteBookingType, e.EventType)
			}
			return true
		})).
		Return(nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	result, err := useCase.CompleteBookingByID(testBookingPast.ID)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if result.Status != enum.StatusCompleted {
		t.Errorf("Expected status %s, got %s", enum.StatusCompleted, result.Status)
	}
}

func TestBookingUseCase_CompleteBookingByID_BookingNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockBookingRepo.EXPECT().
		GetBookingByID(testBookingPast.ID).
		Return(nil, domain.ErrBookingNotFound)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	_, err := useCase.CompleteBookingByID(testBookingPast.ID)

	if !errors.Is(err, domain.ErrBookingNotFound) {
		t.Errorf("Expected ErrBookingNotFound, got error: %v", err)
	}
}

func TestBookingUseCase_CompleteBookingByID_NotFinished(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockBookingRepo.EXPECT().
		GetBookingByID(testBooking.ID).
		Return(testBooking, nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	_, err := useCase.CompleteBookingByID(testBooking.ID)

	if !errors.Is(err, domain.ErrBookingNotFinished) {
		t.Errorf("Expected ErrBookingNotFinished, got: %v", err)
	}
}

func TestBookingUseCase_GetMyBookings_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockCache.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.ErrCacheKeyNotFound)

	expectedBookings := []entity.Booking{*testBooking}
	mockBookingRepo.EXPECT().
		GetAllBookingsByUserID(testClient.ID).
		Return(expectedBookings, nil)

	mockCache.EXPECT().
		Set(gomock.Any(), gomock.Any(), gomock.Any(), 60*time.Second).
		Return(nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	_, err := useCase.GetMyBookings(testClient.ID)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestBookingUseCase_GetMyBookings_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	expectedBookings := []entity.Booking{*testBooking}
	dbCalled := false

	mockCache.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, dest any) error {
			if bookingsPtr, ok := dest.(*[]entity.Booking); ok {
				*bookingsPtr = expectedBookings
			}
			return nil
		})

	mockBookingRepo.EXPECT().
		GetAllBookingsByUserID(gomock.Any()).
		DoAndReturn(func(id int) ([]entity.Booking, error) {
			dbCalled = true
			return expectedBookings, nil
		}).
		MaxTimes(0)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	bookings, err := useCase.GetMyBookings(testClient.ID)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if len(bookings) != 1 {
		t.Errorf("Expected 1 booking, got %d", len(bookings))
	}
	if dbCalled {
		t.Error("DB should not be called on cache hit")
	}
}

func TestBookingUseCase_GetMyBookings_CacheMiss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	expectedBookings := []entity.Booking{*testBooking}
	cacheSetCalled := false

	mockCache.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.ErrCacheKeyNotFound)

	mockBookingRepo.EXPECT().
		GetAllBookingsByUserID(testClient.ID).
		Return(expectedBookings, nil)

	mockCache.EXPECT().
		Set(gomock.Any(), gomock.Any(), gomock.Any(), 60*time.Second).
		DoAndReturn(func(ctx context.Context, key string, value any, ttl time.Duration) error {
			cacheSetCalled = true
			return nil
		})

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	bookings, err := useCase.GetMyBookings(testClient.ID)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if len(bookings) != 1 {
		t.Errorf("Expected 1 booking, got %d", len(bookings))
	}
	if !cacheSetCalled {
		t.Error("Cache Set should be called on cache miss")
	}
}

func TestBookingUseCase_GetMyBookings_CacheSetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	expectedBookings := []entity.Booking{*testBooking}

	mockCache.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.ErrCacheKeyNotFound)

	mockBookingRepo.EXPECT().
		GetAllBookingsByUserID(testClient.ID).
		Return(expectedBookings, nil)

	mockCache.EXPECT().
		Set(gomock.Any(), gomock.Any(), gomock.Any(), 60*time.Second).
		Return(errors.New("cache error"))

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	bookings, err := useCase.GetMyBookings(testClient.ID)

	if err != nil {
		t.Errorf("Cache error should not break flow, got: %v", err)
	}
	if len(bookings) != 1 {
		t.Errorf("Expected 1 booking, got %d", len(bookings))
	}
}

func TestBookingUseCase_CancelMyBooking_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockBookingRepo.EXPECT().
		GetBookingByID(testBooking.ID).
		Return(testBooking, nil)

	mockBookingRepo.EXPECT().
		SetBookingCancel(testBooking.ID, testBooking.UserID).
		Return(&entity.Booking{
			ID:     testBooking.ID,
			Status: enum.StatusCancelled,
		}, nil)

	mockCache.EXPECT().
		DeleteByPrefix(gomock.Any(), gomock.Any()).
		Return(nil)

	mockCache.EXPECT().
		Delete(gomock.Any(), gomock.Any()).
		Return(nil)

	mockOutboxRepo.EXPECT().
		CreateEvent(gomock.Cond(func(e *entity.OutboxEvent) bool {
			if e.EventType != string(enum.CancelBookingType) {
				t.Errorf("Expected event type %s, got %s", enum.CancelBookingType, e.EventType)
			}
			return true
		})).
		Return(nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	err := useCase.CancelMyBooking(testBooking.ID, testBooking.UserID)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestBookingUseCase_CancelMyBooking_BookingIsActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockBookingRepo.EXPECT().
		GetBookingByID(testBookingPast.ID).
		Return(testBookingPast, nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	err := useCase.CancelMyBooking(testBookingPast.ID, testBookingPast.UserID)

	if !errors.Is(err, domain.ErrBookingIsActive) {
		t.Errorf("Expected ErrBookingIsActive, got: %v", err)
	}
}

func TestBookingUseCase_CancelMyBooking_BookingNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockBookingRepo.EXPECT().
		GetBookingByID(testBookingPast.ID).
		Return(nil, domain.ErrBookingNotFound)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo, mockCache)

	err := useCase.CancelMyBooking(testBookingPast.ID, testBookingPast.UserID)

	if !errors.Is(err, domain.ErrBookingNotFound) {
		t.Errorf("Expected ErrBookingNotFound, got error: %v", err)
	}
}
