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

	mockBookingRepo.EXPECT().
		IsSlotAvailable(ctx, futureTime, gomock.Any()).
		Return(true, nil)

	mockUserRepo.EXPECT().
		GetByIDContext(ctx, testClient.ID).
		Return(testClient, nil)

	mockBookingRepo.EXPECT().
		Create(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, b *entity.Booking) error {
			b.ID = 100
			return nil
		})

	mockOutboxRepo.EXPECT().
		CreateEventContext(ctx, gomock.Cond(func(e *entity.OutboxEvent) bool {
			if e.EventType != string(enum.NewBookingType) {
				t.Errorf("Expected event type %s, got %s", enum.NewBookingType, e.EventType)
			}
			return true
		})).
		Return(nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo)

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

func TestBookingUseCase_CreateBooking_PastTimeRange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pastTime := time.Now().Add(-24 * time.Hour)
	ctx := context.Background()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)

	mockUserRepo.EXPECT().
		GetByIDContext(ctx, testClient.ID).
		Return(testClient, nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo)

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

	mockUserRepo.EXPECT().
		GetByIDContext(ctx, testClient.ID).
		Return(testClient, nil)

	mockBookingRepo.EXPECT().
		IsSlotAvailable(ctx, futureTime, gomock.Any()).
		Return(false, nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo)

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

	mockUserRepo.EXPECT().
		GetByIDContext(ctx, testClient.ID).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo)

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

	mockUserRepo.EXPECT().
		GetByID(testStaff.ID).
		Return(testStaff, nil)

	expectedBookings := []entity.Booking{*testBooking}
	mockBookingRepo.EXPECT().
		GetAll().
		Return(expectedBookings, nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo)

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

	mockUserRepo.EXPECT().
		GetByID(testClient.ID).
		Return(testClient, nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo)

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

	mockUserRepo.EXPECT().
		GetByID(testStaff.ID).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo)

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

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo)

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

	mockBookingRepo.EXPECT().
		GetBookingByID(testBookingPast.ID).
		Return(nil, domain.ErrBookingNotFound)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo)

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

	mockBookingRepo.EXPECT().
		GetBookingByID(testBooking.ID).
		Return(testBooking, nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo)

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

	expectedBookings := []entity.Booking{*testBooking}
	mockBookingRepo.EXPECT().
		GetAllBookingsByUserID(testClient.ID).
		Return(expectedBookings, nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo)

	_, err := useCase.GetMyBookings(testClient.ID)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestBookingUseCase_CancelMyBooking_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)

	mockBookingRepo.EXPECT().
		GetBookingByID(testBooking.ID).
		Return(testBooking, nil)

	mockBookingRepo.EXPECT().
		SetBookingCancel(testBooking.ID, testBooking.UserID).
		Return(&entity.Booking{
			ID:     testBooking.ID,
			Status: enum.StatusCancelled,
		}, nil)

	mockOutboxRepo.EXPECT().
		CreateEvent(gomock.Cond(func(e *entity.OutboxEvent) bool {
			if e.EventType != string(enum.CancelBookingType) {
				t.Errorf("Expected event type %s, got %s", enum.CancelBookingType, e.EventType)
			}
			return true
		})).
		Return(nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo)

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

	mockBookingRepo.EXPECT().
		GetBookingByID(testBookingPast.ID).
		Return(testBookingPast, nil)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo)

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

	mockBookingRepo.EXPECT().
		GetBookingByID(testBookingPast.ID).
		Return(nil, domain.ErrBookingNotFound)

	useCase := usecase.NewBookingUseCase(mockBookingRepo, mockUserRepo, mockOutboxRepo)

	err := useCase.CancelMyBooking(testBookingPast.ID, testBookingPast.UserID)

	if !errors.Is(err, domain.ErrBookingNotFound) {
		t.Errorf("Expected ErrBookingNotFound, got error: %v", err)
	}
}
