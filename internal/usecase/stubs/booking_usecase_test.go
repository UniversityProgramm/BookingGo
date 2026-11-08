package stubs

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/dto"
	"BookingGo/internal/enum"
	"BookingGo/internal/usecase"
	"context"
	"errors"
	"testing"
	"time"
)

type StubBookingPostgres struct {
	CreateFunc                 func(ctx context.Context, booking *domain.Booking) error
	IsSlotAvailableFunc        func(ctx context.Context, start, end time.Time) (bool, error)
	GetAllBookingsByUserIDFunc func(id int) ([]domain.Booking, error)
	GetBookingByIDFunc         func(id int) (*domain.Booking, error)
	SetBookingCompleteFunc     func(bookingID int) (*domain.Booking, error)
	SetBookingCancelFunc       func(bookingID, userID int) (*domain.Booking, error)
	GetAllFunc                 func() ([]domain.Booking, error)
}

func (m *StubBookingPostgres) Create(ctx context.Context, booking *domain.Booking) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, booking)
	}
	return domain.ErrNotImplemented
}

func (m *StubBookingPostgres) IsSlotAvailable(ctx context.Context, start, end time.Time) (bool, error) {
	if m.IsSlotAvailableFunc != nil {
		return m.IsSlotAvailableFunc(ctx, start, end)
	}
	return false, domain.ErrNotImplemented
}

func (m *StubBookingPostgres) GetAllBookingsByUserID(id int) ([]domain.Booking, error) {
	if m.GetAllBookingsByUserIDFunc != nil {
		return m.GetAllBookingsByUserIDFunc(id)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubBookingPostgres) GetBookingByID(id int) (*domain.Booking, error) {
	if m.GetBookingByIDFunc != nil {
		return m.GetBookingByIDFunc(id)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubBookingPostgres) SetBookingComplete(bookingID int) (*domain.Booking, error) {
	if m.SetBookingCompleteFunc != nil {
		return m.SetBookingCompleteFunc(bookingID)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubBookingPostgres) SetBookingCancel(bookingID, userID int) (*domain.Booking, error) {
	if m.SetBookingCancelFunc != nil {
		return m.SetBookingCancelFunc(bookingID, userID)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubBookingPostgres) GetAll() ([]domain.Booking, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return nil, domain.ErrNotImplemented
}

var (
	testClient = &domain.User{
		ID:       1,
		Email:    "client@example.com",
		Role:     enum.RoleClient,
		IsActive: true,
		UserNotificationSettings: domain.NotificationSettings{
			IsEmailSend: true,
			IsPhoneSend: false,
		},
	}

	testStaff = &domain.User{
		ID:       3,
		Email:    "staff@mail.ru",
		Role:     enum.RoleStaff,
		IsActive: true,
		UserNotificationSettings: domain.NotificationSettings{
			IsEmailSend: true,
			IsPhoneSend: false,
		},
	}

	testBooking = &domain.Booking{
		ID:                 10,
		UserID:             1,
		SlotStart:          time.Now().Add(time.Hour),
		SlotEnd:            time.Now().Add(2 * time.Hour),
		Status:             enum.StatusConfirmed,
		ProblemDescription: "problem",
	}

	testBookingPast = &domain.Booking{
		ID:                 11,
		UserID:             1,
		SlotStart:          time.Now().Add(-2 * time.Hour),
		SlotEnd:            time.Now().Add(-1 * time.Hour),
		Status:             enum.StatusConfirmed,
		ProblemDescription: "problem",
	}
)

func TestBookingUseCase_CreateBooking_Success(t *testing.T) {
	outboxCalled := false
	futureTime := time.Now().Add(24 * time.Hour)
	cacheSlotInvalidated := false
	cacheUserBookingsInvalidated := false

	stubBookingRepo := &StubBookingPostgres{
		IsSlotAvailableFunc: func(ctx context.Context, start, end time.Time) (bool, error) {
			return true, nil
		},
		CreateFunc: func(ctx context.Context, booking *domain.Booking) error {
			booking.ID = 100
			return nil
		},
	}

	stubUserRepo := &StubUserPostgres{
		GetByIDContextFunc: func(ctx context.Context, id int) (*domain.User, error) {
			return testClient, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{
		CreateEventContextFunc: func(ctx context.Context, event *domain.OutboxEvent) error {
			outboxCalled = true
			if event.EventType != string(enum.NewBookingType) {
				t.Errorf("Неверный тип события: %s", event.EventType)
			}
			return nil
		},
	}
	stubCacheService := &StubCacheService{
		DeleteByPrefixFunc: func(ctx context.Context, prefix string) error {
			cacheSlotInvalidated = true
			return nil
		},
		DeleteFunc: func(ctx context.Context, key string) error {
			cacheUserBookingsInvalidated = true
			return nil
		},
	}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)

	req := &dto.CreateBookingRequest{
		SlotStart:          futureTime,
		ProblemDescription: "Тестовая проблема",
	}

	booking, err := useCase.CreateBooking(context.Background(), testClient.ID, req)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if booking == nil {
		t.Fatal("Запись не была создана")
	}
	if !cacheSlotInvalidated {
		t.Error("Ожидалась инвалидация кэша слотов, но она не была вызвана")
	}
	if !cacheUserBookingsInvalidated {
		t.Error("Ожидалась инвалидация кэша записей пользователя, но она не была вызвана")
	}
	if !outboxCalled {
		t.Error("Outbox событие должно было быть создано")
	}
}

func TestBookingUseCase_CreateBooking_CacheHit(t *testing.T) {
	futureTime := time.Now().Add(24 * time.Hour)
	slotCheckedInDB := false

	stubBookingRepo := &StubBookingPostgres{
		IsSlotAvailableFunc: func(ctx context.Context, start, end time.Time) (bool, error) {
			slotCheckedInDB = true
			return true, nil
		},
		CreateFunc: func(ctx context.Context, booking *domain.Booking) error {
			return nil
		},
	}

	stubUserRepo := &StubUserPostgres{
		GetByIDContextFunc: func(ctx context.Context, id int) (*domain.User, error) {
			return testClient, nil
		},
	}

	stubCache := &StubCacheService{
		GetFunc: func(ctx context.Context, key string, dest any) error {
			if boolPtr, ok := dest.(*bool); ok {
				*boolPtr = true
			}
			return nil
		},
	}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, &StubOutboxPostgres{}, stubCache)

	req := &dto.CreateBookingRequest{
		SlotStart:          futureTime,
		ProblemDescription: "Тест кэша",
	}

	_, err := useCase.CreateBooking(context.Background(), testClient.ID, req)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if slotCheckedInDB {
		t.Error("IsSlotAvailable не должен вызываться при redis hit")
	}
}

func TestBookingUseCase_CreateBooking_OverlappingSlots(t *testing.T) {
	existingStart := time.Now().Add(24 * time.Hour)
	existingEnd := time.Now().Add(25 * time.Hour)

	newStart := time.Now().Add(24 * time.Hour).Add(30 * time.Minute)

	stubBookingRepo := &StubBookingPostgres{
		IsSlotAvailableFunc: func(ctx context.Context, start, end time.Time) (bool, error) {
			if existingStart.Before(end) && start.Before(existingEnd) {
				return false, nil
			}
			return true, nil
		},
	}
	stubUserRepo := &StubUserPostgres{
		GetByIDContextFunc: func(ctx context.Context, id int) (*domain.User, error) {
			return testClient, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)
	req := &dto.CreateBookingRequest{
		SlotStart:          newStart,
		ProblemDescription: "Тест пересечения",
	}

	_, err := useCase.CreateBooking(context.Background(), testClient.ID, req)

	if !errors.Is(err, domain.ErrSlotNotAvailable) {
		t.Errorf("Ожидалась ошибка ErrSlotNotAvailable для пересекающегося слота, получена: %v", err)
	}
}

func TestBookingUseCase_CreateBooking_PastTime(t *testing.T) {
	stubBookingRepo := &StubBookingPostgres{}
	stubUserRepo := &StubUserPostgres{
		GetByIDContextFunc: func(ctx context.Context, id int) (*domain.User, error) {
			return testClient, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)
	req := &dto.CreateBookingRequest{
		SlotStart:          time.Now().Add(-1 * time.Hour),
		ProblemDescription: "test",
	}

	_, err := useCase.CreateBooking(context.Background(), testClient.ID, req)

	if !errors.Is(err, domain.ErrInvalidTimeRange) {
		t.Errorf("Ожидалась ошибка ErrInvalidTimeRange, получена: %v", err)
	}
}

func TestBookingUseCase_CreateBooking_SlotNotAvailable(t *testing.T) {
	stubBookingRepo := &StubBookingPostgres{
		IsSlotAvailableFunc: func(ctx context.Context, start, end time.Time) (bool, error) {
			return false, nil
		},
	}
	stubUserRepo := &StubUserPostgres{
		GetByIDContextFunc: func(ctx context.Context, id int) (*domain.User, error) {
			return testClient, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)

	futureTime := time.Now().Add(24 * time.Hour)
	req := &dto.CreateBookingRequest{
		SlotStart:          futureTime,
		ProblemDescription: "test",
	}

	_, err := useCase.CreateBooking(context.Background(), testClient.ID, req)

	if !errors.Is(err, domain.ErrSlotNotAvailable) {
		t.Errorf("Ожидалась ошибка ErrSlotNotAvailable, получена: %v", err)
	}
}

func TestBookingUseCase_CreateBooking_UserNotFound(t *testing.T) {
	stubBookingRepo := &StubBookingPostgres{}
	stubUserRepo := &StubUserPostgres{
		GetByIDContextFunc: func(ctx context.Context, id int) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)
	req := &dto.CreateBookingRequest{
		SlotStart:          time.Now().Add(24 * time.Hour),
		ProblemDescription: "test",
	}

	_, err := useCase.CreateBooking(context.Background(), 999, req)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestBookingUseCase_IsSlotAvailableCached_CacheHit(t *testing.T) {
	start := time.Now().Add(24 * time.Hour)
	end := start.Add(time.Hour)
	dbCalled := false

	stubBookingRepo := &StubBookingPostgres{
		IsSlotAvailableFunc: func(ctx context.Context, start, end time.Time) (bool, error) {
			dbCalled = true
			return true, nil
		},
	}

	stubCache := &StubCacheService{
		GetFunc: func(ctx context.Context, key string, dest any) error {
			if boolPtr, ok := dest.(*bool); ok {
				*boolPtr = true
			}
			return nil
		},
	}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, &StubUserPostgres{}, &StubOutboxPostgres{}, stubCache)

	available, err := useCase.IsSlotAvailableCached(context.Background(), start, end)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if !available {
		t.Error("Ожидалось available=true из кэша")
	}
	if dbCalled {
		t.Error("БД не должна вызываться при redis hit")
	}
}

func TestBookingUseCase_IsSlotAvailableCached_CacheMiss(t *testing.T) {
	start := time.Now().Add(24 * time.Hour)
	end := start.Add(time.Hour)
	cacheSetCalled := false

	stubBookingRepo := &StubBookingPostgres{
		IsSlotAvailableFunc: func(ctx context.Context, start, end time.Time) (bool, error) {
			return true, nil
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

	useCase := usecase.NewBookingUseCase(stubBookingRepo, &StubUserPostgres{}, &StubOutboxPostgres{}, stubCache)

	available, err := useCase.IsSlotAvailableCached(context.Background(), start, end)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if !available {
		t.Error("Ожидалось available=true из БД")
	}
	if !cacheSetCalled {
		t.Error("Cache Set должен вызываться при redis miss")
	}
}

func TestBookingUseCase_IsSlotAvailableCached_CacheReturnsFalse(t *testing.T) {
	start := time.Now().Add(24 * time.Hour)
	end := start.Add(time.Hour)
	dbCalled := false

	stubBookingRepo := &StubBookingPostgres{
		IsSlotAvailableFunc: func(ctx context.Context, start, end time.Time) (bool, error) {
			dbCalled = true
			return false, nil
		},
	}

	stubCache := &StubCacheService{
		GetFunc: func(ctx context.Context, key string, dest any) error {
			if boolPtr, ok := dest.(*bool); ok {
				*boolPtr = false
			}
			return nil
		},
	}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, &StubUserPostgres{}, &StubOutboxPostgres{}, stubCache)

	available, err := useCase.IsSlotAvailableCached(context.Background(), start, end)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if available {
		t.Error("Ожидалось available=false из кэша")
	}
	if dbCalled {
		t.Error("БД не должна вызываться при redis hit")
	}
}

func TestBookingUseCase_IsSlotAvailableCached_DBError(t *testing.T) {
	start := time.Now().Add(24 * time.Hour)
	end := start.Add(time.Hour)
	dbError := errors.New("database error")

	stubBookingRepo := &StubBookingPostgres{
		IsSlotAvailableFunc: func(ctx context.Context, start, end time.Time) (bool, error) {
			return false, dbError
		},
	}

	stubCache := &StubCacheService{
		GetFunc: func(ctx context.Context, key string, dest any) error {
			return domain.ErrCacheKeyNotFound
		},
	}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, &StubUserPostgres{}, &StubOutboxPostgres{}, stubCache)

	_, err := useCase.IsSlotAvailableCached(context.Background(), start, end)

	if err == nil {
		t.Fatal("Ожидалась ошибка БД, получена nil")
	}
	if !errors.Is(err, dbError) {
		t.Errorf("Ожидалась ошибка БД, получена: %v", err)
	}
}

func TestBookingUseCase_IsSlotAvailableCached_CacheSetError(t *testing.T) {
	start := time.Now().Add(24 * time.Hour)
	end := start.Add(time.Hour)
	cacheError := errors.New("redis error")

	stubBookingRepo := &StubBookingPostgres{
		IsSlotAvailableFunc: func(ctx context.Context, start, end time.Time) (bool, error) {
			return true, nil
		},
	}

	stubCache := &StubCacheService{
		GetFunc: func(ctx context.Context, key string, dest any) error {
			return domain.ErrCacheKeyNotFound
		},
		SetFunc: func(ctx context.Context, key string, value any, ttl time.Duration) error {
			return cacheError
		},
	}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, &StubUserPostgres{}, &StubOutboxPostgres{}, stubCache)

	available, err := useCase.IsSlotAvailableCached(context.Background(), start, end)

	if err != nil {
		t.Fatalf("Ошибка кэша не должна ломать флоу, получена: %v", err)
	}
	if !available {
		t.Error("Ожидалось available=true из БД")
	}
}

func TestBookingUseCase_GetAllBookings_Staff_Success(t *testing.T) {
	expectedBookings := []domain.Booking{*testBooking, *testBookingPast}

	stubUserRepo := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testStaff, nil
		},
	}

	stubBookingRepo := &StubBookingPostgres{
		GetAllFunc: func() ([]domain.Booking, error) {
			return expectedBookings, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)

	_, err := useCase.GetAllBookings(testStaff.ID)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestBookingUseCase_GetAllBookings_OnlyForStaffOrAdmin(t *testing.T) {
	stubUserRepo := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testClient, nil
		},
	}

	stubBookingRepo := &StubBookingPostgres{
		GetAllFunc: func() ([]domain.Booking, error) {
			return []domain.Booking{}, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)

	_, err := useCase.GetAllBookings(testClient.ID)

	if !errors.Is(err, domain.ErrOnlyForStaffOrAdmin) {
		t.Errorf("Ожидалась ошибка ErrOnlyForStaffOrAdmin, получена: %v", err)
	}
}

func TestBookingUseCase_GetAllBookings_UserNotFound(t *testing.T) {
	stubUserRepo := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubBookingRepo := &StubBookingPostgres{}
	stubNotifier := &StubOutboxPostgres{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)

	_, err := useCase.GetAllBookings(999)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestBookingUseCase_CompleteBookingByID_Success(t *testing.T) {
	outboxCalled := false

	stubBookingRepo := &StubBookingPostgres{
		GetBookingByIDFunc: func(id int) (*domain.Booking, error) {
			return testBookingPast, nil
		},
		SetBookingCompleteFunc: func(bookingID int) (*domain.Booking, error) {
			updated := *testBookingPast
			updated.Status = enum.StatusCompleted
			return &updated, nil
		},
	}
	stubUserRepo := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testStaff, nil
		},
	}
	stubNotifier := &StubOutboxPostgres{
		CreateEventFunc: func(event *domain.OutboxEvent) error {
			outboxCalled = true
			if event.EventType != string(enum.CompleteBookingType) {
				t.Errorf("Неверный тип события: %s", event.EventType)
			}
			return nil
		},
	}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)
	result, err := useCase.CompleteBookingByID(100)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if result.Status != enum.StatusCompleted {
		t.Errorf("Статус не обновлен: %s", result.Status)
	}
	if !outboxCalled {
		t.Error("Outbox событие должно было быть создано")
	}
}

func TestBookingUseCase_CompleteBookingByID_BookingNotFound(t *testing.T) {
	stubBookingRepo := &StubBookingPostgres{
		GetBookingByIDFunc: func(id int) (*domain.Booking, error) {
			return nil, domain.ErrBookingNotFound
		},
	}
	stubUserRepo := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testStaff, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)
	_, err := useCase.CompleteBookingByID(100)

	if !errors.Is(err, domain.ErrBookingNotFound) {
		t.Errorf("Ожидалась ошибка ErrBookingNotFound, получена: %v", err)
	}
}

func TestBookingUseCase_CompleteBookingByID_NotFinishedYet(t *testing.T) {
	stubBookingRepo := &StubBookingPostgres{
		GetBookingByIDFunc: func(id int) (*domain.Booking, error) {
			return testBooking, nil
		},
	}
	stubUserRepo := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testStaff, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)
	_, err := useCase.CompleteBookingByID(101)

	if !errors.Is(err, domain.ErrBookingNotFinished) {
		t.Errorf("Ожидалась ошибка ErrBookingNotFinished, получена: %v", err)
	}
}

func TestBookingUseCase_GetMyBookings_Success(t *testing.T) {
	expectedBookings := []domain.Booking{*testBooking}

	stubBookingRepo := &StubBookingPostgres{
		GetAllBookingsByUserIDFunc: func(id int) ([]domain.Booking, error) {
			if id != testClient.ID {
				t.Errorf("Неверный UserID: %d", id)
			}
			return expectedBookings, nil
		},
	}
	stubUserRepo := &StubUserPostgres{}

	stubNotifier := &StubOutboxPostgres{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)
	_, err := useCase.GetMyBookings(testClient.ID)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestBookingUseCase_GetMyBookings_CacheHit(t *testing.T) {
	expectedBookings := []domain.Booking{*testBooking}

	stubBookingRepo := &StubBookingPostgres{
		GetAllBookingsByUserIDFunc: func(id int) ([]domain.Booking, error) {
			return expectedBookings, nil
		},
	}

	stubCache := &StubCacheService{
		GetFunc: func(ctx context.Context, key string, dest any) error {
			if boolPtr, ok := dest.(*bool); ok {
				*boolPtr = true
			}
			return nil
		},
	}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, &StubUserPostgres{}, &StubOutboxPostgres{}, stubCache)

	_, err := useCase.GetMyBookings(testClient.ID)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestBookingUseCase_GetMyBookings_CacheMiss(t *testing.T) {
	expectedBookings := []domain.Booking{*testBooking}
	cacheSetCalled := false

	stubBookingRepo := &StubBookingPostgres{
		GetAllBookingsByUserIDFunc: func(id int) ([]domain.Booking, error) {
			return expectedBookings, nil
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

	useCase := usecase.NewBookingUseCase(stubBookingRepo, &StubUserPostgres{}, &StubOutboxPostgres{}, stubCache)

	_, err := useCase.GetMyBookings(testClient.ID)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if !cacheSetCalled {
		t.Error("Cache Set должен вызываться при redis miss")
	}
}

func TestBookingUseCase_GetMyBookings_Error(t *testing.T) {
	dbError := errors.New("database error")
	stubBookingRepo := &StubBookingPostgres{
		GetAllBookingsByUserIDFunc: func(id int) ([]domain.Booking, error) {
			return nil, dbError
		},
	}
	stubUserRepo := &StubUserPostgres{}

	stubNotifier := &StubOutboxPostgres{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)
	_, err := useCase.GetMyBookings(testClient.ID)

	if err == nil {
		t.Fatal("Ожидалась ошибка, получена nil")
	}
	if !errors.Is(err, dbError) {
		t.Errorf("Ожидалась оригинальная ошибка, получена: %v", err)
	}
}

func TestBookingUseCase_CancelMyBooking_Success(t *testing.T) {
	cacheSlotInvalidated := false
	cacheUserBookingsInvalidated := false
	outboxCalled := false

	stubBookingRepo := &StubBookingPostgres{
		GetBookingByIDFunc: func(id int) (*domain.Booking, error) {
			return testBooking, nil
		},
		SetBookingCancelFunc: func(bookingID, userID int) (*domain.Booking, error) {
			return &domain.Booking{}, nil
		},
	}
	stubUserRepo := &StubUserPostgres{}
	stubNotifier := &StubOutboxPostgres{
		CreateEventFunc: func(event *domain.OutboxEvent) error {
			outboxCalled = true
			if event.EventType != string(enum.CancelBookingType) {
				t.Errorf("Неверный тип события: %s", event.EventType)
			}
			return nil
		},
	}
	stubCacheService := &StubCacheService{
		DeleteByPrefixFunc: func(ctx context.Context, prefix string) error {
			cacheSlotInvalidated = true
			return nil
		},
		DeleteFunc: func(ctx context.Context, key string) error {
			cacheUserBookingsInvalidated = true
			return nil
		},
	}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)
	err := useCase.CancelMyBooking(10, testClient.ID)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if !cacheSlotInvalidated {
		t.Error("Ожидалась инвалидация кэша слотов, но она не была вызвана")
	}
	if !cacheUserBookingsInvalidated {
		t.Error("Ожидалась инвалидация кэша записей пользователя, но она не была вызвана")
	}
	if !outboxCalled {
		t.Error("Outbox событие должно было быть создано")
	}
}

func TestBookingUseCase_CancelMyBooking_BookingNotFound(t *testing.T) {
	stubBookingRepo := &StubBookingPostgres{
		GetBookingByIDFunc: func(id int) (*domain.Booking, error) {
			return nil, domain.ErrBookingNotFound
		},
	}
	stubUserRepo := &StubUserPostgres{}

	stubNotifier := &StubOutboxPostgres{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)
	err := useCase.CancelMyBooking(999, testClient.ID)

	if !errors.Is(err, domain.ErrBookingNotFound) {
		t.Errorf("Ожидалась ошибка ErrBookingNotFound, получена: %v", err)
	}
}

func TestBookingUseCase_CancelMyBooking_BookingAlreadyActive(t *testing.T) {
	stubBookingRepo := &StubBookingPostgres{
		GetBookingByIDFunc: func(id int) (*domain.Booking, error) {
			return testBookingPast, nil
		},
	}
	stubUserRepo := &StubUserPostgres{}

	stubNotifier := &StubOutboxPostgres{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier, stubCacheService)
	err := useCase.CancelMyBooking(10, testClient.ID)

	if !errors.Is(err, domain.ErrBookingIsActive) {
		t.Errorf("Ожидалась ошибка ErrBookingAlreadyActive, получена: %v", err)
	}

}
