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
	testNotifUserWithAll = &entity.User{
		ID:    1,
		Email: "user@example.com",
		FIO:   "Иванов Иван Иванович",
		Phone: "+79991234567",
		Role:  enum.RoleClient,
		UserNotificationSettings: entity.NotificationSettings{
			IsEmailSend: true,
			IsPhoneSend: true,
		},
	}

	testNotifUserWebOnly = &entity.User{
		ID:    2,
		Email: "web@example.com",
		FIO:   "Петров Петр Петрович",
		Phone: "+79997654321",
		Role:  enum.RoleClient,
		UserNotificationSettings: entity.NotificationSettings{
			IsEmailSend: false,
			IsPhoneSend: false,
		},
	}

	testNotifBooking = &entity.Booking{
		ID:                 100,
		UserID:             1,
		SlotStart:          time.Now().Add(24 * time.Hour),
		SlotEnd:            time.Now().Add(25 * time.Hour),
		Status:             enum.StatusConfirmed,
		ProblemDescription: "Engine not working",
	}

	testNotifNotification = &entity.Notification{
		ID:          1,
		RecipientID: 1,
		Title:       "Test title",
		Body:        "Test body",
		IsRead:      false,
	}
)

func TestNotificationUseCase_CreateNotification_Success_WithAllChannels(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifRepo := NewMockNotificationRepository(ctrl)
	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockUserRepo.EXPECT().
		GetByID(testNotifUserWithAll.ID).
		Return(testNotifUserWithAll, nil)

	mockBookingRepo.EXPECT().
		GetBookingByID(testNotifBooking.ID).
		Return(testNotifBooking, nil)

	mockNotifRepo.EXPECT().
		CreateNotif(gomock.Cond(func(n *entity.Notification) bool {
			if n.RecipientID != testNotifUserWithAll.ID {
				t.Errorf("Expected RecipientID %d, got %d", testNotifUserWithAll.ID, n.RecipientID)
			}
			if n.IsRead != false {
				t.Error("New notification should have IsRead=false")
			}
			if n.Title == "" {
				t.Error("Notification title should not be empty")
			}
			if n.Body == "" {
				t.Error("Notification body should not be empty")
			}
			return true
		})).
		Return(nil)

	mockCache.EXPECT().
		Delete(context.Background(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewNotificationUseCase(mockNotifRepo, mockBookingRepo, mockUserRepo, mockCache)

	params := entity.NotificationParams{
		BookingID: testNotifBooking.ID,
		IP:        "127.0.0.1",
	}

	err := useCase.CreateNotification(testNotifUserWithAll.ID, enum.NewBookingType, params)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
}

func TestNotificationUseCase_CreateNotification_WebOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifRepo := NewMockNotificationRepository(ctrl)
	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockUserRepo.EXPECT().
		GetByID(testNotifUserWebOnly.ID).
		Return(testNotifUserWebOnly, nil)

	mockNotifRepo.EXPECT().
		CreateNotif(gomock.Any()).
		Return(nil)

	mockCache.EXPECT().
		Delete(context.Background(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewNotificationUseCase(mockNotifRepo, mockBookingRepo, mockUserRepo, mockCache)

	params := entity.NotificationParams{}

	err := useCase.CreateNotification(testNotifUserWebOnly.ID, enum.AuthType, params)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
}

func TestNotificationUseCase_CreateNotification_WithoutBooking(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifRepo := NewMockNotificationRepository(ctrl)
	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockUserRepo.EXPECT().
		GetByID(testNotifUserWebOnly.ID).
		Return(testNotifUserWebOnly, nil)

	mockNotifRepo.EXPECT().
		CreateNotif(gomock.Any()).
		Return(nil)

	mockBookingRepo.EXPECT().
		GetBookingByID(gomock.Any()).
		Times(0)

	mockCache.EXPECT().
		Delete(context.Background(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewNotificationUseCase(mockNotifRepo, mockBookingRepo, mockUserRepo, mockCache)

	params := entity.NotificationParams{
		BookingID: 0,
	}

	err := useCase.CreateNotification(testNotifUserWebOnly.ID, enum.AuthType, params)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
}

func TestNotificationUseCase_CreateNotification_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifRepo := NewMockNotificationRepository(ctrl)
	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockUserRepo.EXPECT().
		GetByID(999).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewNotificationUseCase(mockNotifRepo, mockBookingRepo, mockUserRepo, mockCache)

	params := entity.NotificationParams{}

	err := useCase.CreateNotification(999, enum.AuthType, params)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

func TestNotificationUseCase_GetMyNotifications_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifRepo := NewMockNotificationRepository(ctrl)
	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	expectedNotifs := []entity.Notification{*testNotifNotification}

	mockCache.EXPECT().
		Get(context.Background(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, dest any) error {
			if notifsPtr, ok := dest.(*[]entity.Notification); ok {
				*notifsPtr = expectedNotifs
			}
			return nil
		})

	mockNotifRepo.EXPECT().
		GetAllNotificationsByUserID(gomock.Any()).
		Times(0)

	useCase := usecase.NewNotificationUseCase(mockNotifRepo, mockBookingRepo, mockUserRepo, mockCache)

	notifs, err := useCase.GetMyNotifications(testNotifUserWithAll.ID)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if len(notifs) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(notifs))
	}
}

func TestNotificationUseCase_GetMyNotifications_CacheMiss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifRepo := NewMockNotificationRepository(ctrl)
	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	expectedNotifs := []entity.Notification{*testNotifNotification}

	mockCache.EXPECT().
		Get(context.Background(), gomock.Any(), gomock.Any()).
		Return(domain.ErrCacheKeyNotFound)

	mockNotifRepo.EXPECT().
		GetAllNotificationsByUserID(testNotifUserWithAll.ID).
		Return(expectedNotifs, nil)

	mockCache.EXPECT().
		Set(context.Background(), gomock.Any(), gomock.Any(), 60*time.Second).
		Return(nil)

	useCase := usecase.NewNotificationUseCase(mockNotifRepo, mockBookingRepo, mockUserRepo, mockCache)

	notifs, err := useCase.GetMyNotifications(testNotifUserWithAll.ID)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if len(notifs) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(notifs))
	}
}

func TestNotificationUseCase_GetMyNotifications_CacheSetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifRepo := NewMockNotificationRepository(ctrl)
	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	expectedNotifs := []entity.Notification{*testNotifNotification}

	mockCache.EXPECT().
		Get(context.Background(), gomock.Any(), gomock.Any()).
		Return(domain.ErrCacheKeyNotFound)

	mockNotifRepo.EXPECT().
		GetAllNotificationsByUserID(testNotifUserWithAll.ID).
		Return(expectedNotifs, nil)

	mockCache.EXPECT().
		Set(context.Background(), gomock.Any(), gomock.Any(), 60*time.Second).
		Return(errors.New("cache error"))

	useCase := usecase.NewNotificationUseCase(mockNotifRepo, mockBookingRepo, mockUserRepo, mockCache)

	notifs, err := useCase.GetMyNotifications(testNotifUserWithAll.ID)

	if err != nil {
		t.Fatalf("Cache error should not break flow, got: %v", err)
	}
	if len(notifs) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(notifs))
	}
}

func TestNotificationUseCase_UpdateNotificationSettings_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifRepo := NewMockNotificationRepository(ctrl)
	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	newSettings := &entity.NotificationSettings{
		IsEmailSend: false,
		IsPhoneSend: true,
	}

	mockNotifRepo.EXPECT().
		UpdateSettings(testNotifUserWithAll.ID, newSettings).
		Return(nil)

	mockCache.EXPECT().
		Delete(context.Background(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewNotificationUseCase(mockNotifRepo, mockBookingRepo, mockUserRepo, mockCache)

	err := useCase.UpdateNotificationSettings(testNotifUserWithAll.ID, newSettings)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
}

func TestNotificationUseCase_UpdateNotificationSettings_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifRepo := NewMockNotificationRepository(ctrl)
	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	dbError := errors.New("database error")

	mockNotifRepo.EXPECT().
		UpdateSettings(testNotifUserWithAll.ID, gomock.Any()).
		Return(dbError)

	useCase := usecase.NewNotificationUseCase(mockNotifRepo, mockBookingRepo, mockUserRepo, mockCache)

	err := useCase.UpdateNotificationSettings(testNotifUserWithAll.ID, &entity.NotificationSettings{})

	if !errors.Is(err, dbError) {
		t.Errorf("Expected database error, got: %v", err)
	}
}

func TestNotificationUseCase_MarkAllAsRead_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifRepo := NewMockNotificationRepository(ctrl)
	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	mockNotifRepo.EXPECT().
		MarkAllAsRead(testNotifUserWithAll.ID).
		Return(nil)

	mockCache.EXPECT().
		Delete(context.Background(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewNotificationUseCase(mockNotifRepo, mockBookingRepo, mockUserRepo, mockCache)

	err := useCase.MarkAllAsRead(testNotifUserWithAll.ID)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
}

func TestNotificationUseCase_MarkAllAsRead_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNotifRepo := NewMockNotificationRepository(ctrl)
	mockBookingRepo := NewMockBookingRepository(ctrl)
	mockUserRepo := NewMockUserRepository(ctrl)
	mockCache := NewMockCache(ctrl)

	dbError := errors.New("database error")

	mockNotifRepo.EXPECT().
		MarkAllAsRead(testNotifUserWithAll.ID).
		Return(dbError)

	useCase := usecase.NewNotificationUseCase(mockNotifRepo, mockBookingRepo, mockUserRepo, mockCache)

	err := useCase.MarkAllAsRead(testNotifUserWithAll.ID)

	if !errors.Is(err, dbError) {
		t.Errorf("Expected database error, got: %v", err)
	}
}
