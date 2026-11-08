package stubs

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/enum"
	"BookingGo/internal/usecase"
	"context"
	"errors"
	"testing"
	"time"
)

type StubNotificationPostgres struct {
	UpdateSettingsFunc              func(userID int, req *domain.NotificationSettings) error
	GetAllNotificationsByUserIDFunc func(userID int) ([]domain.Notification, error)
	MarkAllAsReadFunc               func(recipientID int) error
	CreateNotifFunc                 func(notification *domain.Notification) error
}

func (m *StubNotificationPostgres) UpdateSettings(userID int, req *domain.NotificationSettings) error {
	if m.UpdateSettingsFunc != nil {
		return m.UpdateSettingsFunc(userID, req)
	}
	return domain.ErrNotImplemented
}

func (m *StubNotificationPostgres) GetAllNotificationsByUserID(userID int) ([]domain.Notification, error) {
	if m.GetAllNotificationsByUserIDFunc != nil {
		return m.GetAllNotificationsByUserIDFunc(userID)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubNotificationPostgres) MarkAllAsRead(recipientID int) error {
	if m.MarkAllAsReadFunc != nil {
		return m.MarkAllAsReadFunc(recipientID)
	}
	return domain.ErrNotImplemented
}

func (m *StubNotificationPostgres) CreateNotif(notification *domain.Notification) error {
	if m.CreateNotifFunc != nil {
		return m.CreateNotifFunc(notification)
	}
	return domain.ErrNotImplemented
}

var (
	testNotifUserWithAll = &domain.User{
		ID:    1,
		Email: "user@example.com",
		FIO:   "Иванов Иван Иванович",
		Phone: "+79991234567",
		Role:  enum.RoleClient,
		UserNotificationSettings: domain.NotificationSettings{
			IsEmailSend: true,
			IsPhoneSend: true,
		},
	}

	testNotifUserWebOnly = &domain.User{
		ID:    2,
		Email: "web@example.com",
		FIO:   "Петров Петр Петрович",
		Phone: "+79997654321",
		Role:  enum.RoleClient,
		UserNotificationSettings: domain.NotificationSettings{
			IsEmailSend: false,
			IsPhoneSend: false,
		},
	}

	testNotifBooking = &domain.Booking{
		ID:                 100,
		UserID:             1,
		SlotStart:          time.Now().Add(24 * time.Hour),
		SlotEnd:            time.Now().Add(25 * time.Hour),
		Status:             enum.StatusConfirmed,
		ProblemDescription: "Engine not working",
	}

	testNotifNotification = &domain.Notification{
		ID:          1,
		RecipientID: 1,
		Title:       "Test title",
		Body:        "Test body",
		IsRead:      false,
	}
)

func TestNotificationUseCase_CreateNotification_Success_WithAllChannels(t *testing.T) {
	notifCreated := false
	cacheInvalidated := false

	stubNotifRepo := &StubNotificationPostgres{
		CreateNotifFunc: func(notification *domain.Notification) error {
			notifCreated = true
			if notification.RecipientID != testNotifUserWithAll.ID {
				t.Errorf("Expected RecipientID %d, got %d", testNotifUserWithAll.ID, notification.RecipientID)
			}
			if notification.IsRead != false {
				t.Error("New notification should have IsRead=false")
			}
			if notification.Title == "" {
				t.Error("Notification title should not be empty")
			}
			if notification.Body == "" {
				t.Error("Notification body should not be empty")
			}
			return nil
		},
	}

	stubUserRepo := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testNotifUserWithAll, nil
		},
	}

	stubBookingRepo := &StubBookingPostgres{
		GetBookingByIDFunc: func(id int) (*domain.Booking, error) {
			return testNotifBooking, nil
		},
	}

	stubCache := &StubCacheService{
		DeleteFunc: func(ctx context.Context, key string) error {
			cacheInvalidated = true
			return nil
		},
	}

	useCase := usecase.NewNotificationUseCase(stubNotifRepo, stubBookingRepo, stubUserRepo, stubCache)

	params := domain.NotificationParams{
		BookingID: testNotifBooking.ID,
		IP:        "127.0.0.1",
	}

	err := useCase.CreateNotification(testNotifUserWithAll.ID, enum.NewBookingType, params)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if !notifCreated {
		t.Error("Notification should be created")
	}
	if !cacheInvalidated {
		t.Error("Notifications redis should be invalidated")
	}
}

func TestNotificationUseCase_CreateNotification_WebOnly(t *testing.T) {
	notifCreated := false

	stubNotifRepo := &StubNotificationPostgres{
		CreateNotifFunc: func(notification *domain.Notification) error {
			notifCreated = true
			return nil
		},
	}

	stubUserRepo := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testNotifUserWebOnly, nil
		},
	}

	stubBookingRepo := &StubBookingPostgres{}
	stubCache := &StubCacheService{}

	useCase := usecase.NewNotificationUseCase(stubNotifRepo, stubBookingRepo, stubUserRepo, stubCache)

	params := domain.NotificationParams{
		BookingID: testNotifBooking.ID,
	}

	err := useCase.CreateNotification(testNotifUserWebOnly.ID, enum.AuthType, params)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if !notifCreated {
		t.Error("Notification should be created")
	}
}

func TestNotificationUseCase_CreateNotification_WithoutBooking(t *testing.T) {
	notifCreated := false
	bookingRepoCalled := false

	stubNotifRepo := &StubNotificationPostgres{
		CreateNotifFunc: func(notification *domain.Notification) error {
			notifCreated = true
			return nil
		},
	}

	stubUserRepo := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testNotifUserWebOnly, nil
		},
	}

	stubBookingRepo := &StubBookingPostgres{
		GetBookingByIDFunc: func(id int) (*domain.Booking, error) {
			bookingRepoCalled = true
			return nil, nil
		},
	}

	stubCache := &StubCacheService{}

	useCase := usecase.NewNotificationUseCase(stubNotifRepo, stubBookingRepo, stubUserRepo, stubCache)

	params := domain.NotificationParams{
		BookingID: 0,
	}

	err := useCase.CreateNotification(testNotifUserWebOnly.ID, enum.AuthType, params)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if !notifCreated {
		t.Error("Notification should be created")
	}
	if bookingRepoCalled {
		t.Error("BookingRepo should not be called when BookingID == 0")
	}
}

func TestNotificationUseCase_CreateNotification_UserNotFound(t *testing.T) {
	stubNotifRepo := &StubNotificationPostgres{}
	stubUserRepo := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	stubBookingRepo := &StubBookingPostgres{}
	stubCache := &StubCacheService{}

	useCase := usecase.NewNotificationUseCase(stubNotifRepo, stubBookingRepo, stubUserRepo, stubCache)

	err := useCase.CreateNotification(999, enum.AuthType, domain.NotificationParams{})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

func TestNotificationUseCase_GetMyNotifications_CacheHit(t *testing.T) {
	expectedNotifs := []domain.Notification{*testNotifNotification}
	dbCalled := false

	stubNotifRepo := &StubNotificationPostgres{
		GetAllNotificationsByUserIDFunc: func(userID int) ([]domain.Notification, error) {
			dbCalled = true
			return expectedNotifs, nil
		},
	}

	stubCache := &StubCacheService{
		GetFunc: func(ctx context.Context, key string, dest any) error {
			if notifsPtr, ok := dest.(*[]domain.Notification); ok {
				*notifsPtr = expectedNotifs
			}
			return nil
		},
	}

	useCase := usecase.NewNotificationUseCase(stubNotifRepo, nil, nil, stubCache)

	notifs, err := useCase.GetMyNotifications(testNotifUserWithAll.ID)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if len(notifs) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(notifs))
	}
	if dbCalled {
		t.Error("DB should not be called on redis hit")
	}
}

func TestNotificationUseCase_GetMyNotifications_CacheMiss(t *testing.T) {
	expectedNotifs := []domain.Notification{*testNotifNotification}
	cacheSetCalled := false

	stubNotifRepo := &StubNotificationPostgres{
		GetAllNotificationsByUserIDFunc: func(userID int) ([]domain.Notification, error) {
			return expectedNotifs, nil
		},
	}

	stubCache := &StubCacheService{
		GetFunc: func(ctx context.Context, key string, dest any) error {
			return domain.ErrCacheKeyNotFound
		},
		SetFunc: func(ctx context.Context, key string, value any, ttl time.Duration) error {
			cacheSetCalled = true
			return nil
		},
	}

	useCase := usecase.NewNotificationUseCase(stubNotifRepo, nil, nil, stubCache)

	notifs, err := useCase.GetMyNotifications(testNotifUserWithAll.ID)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if len(notifs) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(notifs))
	}
	if !cacheSetCalled {
		t.Error("Cache Set should be called on redis miss")
	}
}

func TestNotificationUseCase_GetMyNotifications_CacheSetError(t *testing.T) {
	expectedNotifs := []domain.Notification{*testNotifNotification}

	stubNotifRepo := &StubNotificationPostgres{
		GetAllNotificationsByUserIDFunc: func(userID int) ([]domain.Notification, error) {
			return expectedNotifs, nil
		},
	}

	stubCache := &StubCacheService{
		GetFunc: func(ctx context.Context, key string, dest any) error {
			return domain.ErrCacheKeyNotFound
		},
		SetFunc: func(ctx context.Context, key string, value any, ttl time.Duration) error {
			return errors.New("redis error")
		},
	}

	useCase := usecase.NewNotificationUseCase(stubNotifRepo, nil, nil, stubCache)

	notifs, err := useCase.GetMyNotifications(testNotifUserWithAll.ID)

	if err != nil {
		t.Fatalf("Cache error should not break flow, got: %v", err)
	}
	if len(notifs) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(notifs))
	}
}

func TestNotificationUseCase_UpdateNotificationSettings_Success(t *testing.T) {
	settingsUpdated := false
	cacheInvalidated := false

	stubNotifRepo := &StubNotificationPostgres{
		UpdateSettingsFunc: func(userID int, req *domain.NotificationSettings) error {
			settingsUpdated = true
			if userID != testNotifUserWithAll.ID {
				t.Errorf("Expected userID %d, got %d", testNotifUserWithAll.ID, userID)
			}
			return nil
		},
	}

	stubCache := &StubCacheService{
		DeleteFunc: func(ctx context.Context, key string) error {
			cacheInvalidated = true
			return nil
		},
	}

	useCase := usecase.NewNotificationUseCase(stubNotifRepo, nil, nil, stubCache)

	newSettings := &domain.NotificationSettings{
		IsEmailSend: false,
		IsPhoneSend: true,
	}

	err := useCase.UpdateNotificationSettings(testNotifUserWithAll.ID, newSettings)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if !settingsUpdated {
		t.Error("Settings should be updated")
	}
	if !cacheInvalidated {
		t.Error("User redis should be invalidated")
	}
}

func TestNotificationUseCase_UpdateNotificationSettings_DBError(t *testing.T) {
	dbError := errors.New("database error")

	stubNotifRepo := &StubNotificationPostgres{
		UpdateSettingsFunc: func(userID int, req *domain.NotificationSettings) error {
			return dbError
		},
	}

	stubCache := &StubCacheService{}

	useCase := usecase.NewNotificationUseCase(stubNotifRepo, nil, nil, stubCache)

	err := useCase.UpdateNotificationSettings(testNotifUserWithAll.ID, &domain.NotificationSettings{})

	if !errors.Is(err, dbError) {
		t.Errorf("Expected database error, got: %v", err)
	}
}

func TestNotificationUseCase_MarkAllAsRead_Success(t *testing.T) {
	markedAsRead := false
	cacheInvalidated := false

	stubNotifRepo := &StubNotificationPostgres{
		MarkAllAsReadFunc: func(recipientID int) error {
			markedAsRead = true
			if recipientID != testNotifUserWithAll.ID {
				t.Errorf("Expected recipientID %d, got %d", testNotifUserWithAll.ID, recipientID)
			}
			return nil
		},
	}

	stubCache := &StubCacheService{
		DeleteFunc: func(ctx context.Context, key string) error {
			cacheInvalidated = true
			return nil
		},
	}

	useCase := usecase.NewNotificationUseCase(stubNotifRepo, nil, nil, stubCache)

	err := useCase.MarkAllAsRead(testNotifUserWithAll.ID)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if !markedAsRead {
		t.Error("Notifications should be marked as read")
	}
	if !cacheInvalidated {
		t.Error("Notifications redis should be invalidated")
	}
}

func TestNotificationUseCase_MarkAllAsRead_DBError(t *testing.T) {
	dbError := errors.New("database error")

	stubNotifRepo := &StubNotificationPostgres{
		MarkAllAsReadFunc: func(recipientID int) error {
			return dbError
		},
	}

	stubCache := &StubCacheService{}

	useCase := usecase.NewNotificationUseCase(stubNotifRepo, nil, nil, stubCache)

	err := useCase.MarkAllAsRead(testNotifUserWithAll.ID)

	if !errors.Is(err, dbError) {
		t.Errorf("Expected database error, got: %v", err)
	}
}
