package stub

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/internal/usecase"
	"errors"
	"testing"
	"time"
)

type StubBookingRepository struct {
	CreateFunc                 func(booking *entity.Booking) error
	IsSlotAvailableFunc        func(start, end time.Time) (bool, error)
	GetAllBookingsByUserIDFunc func(id int) ([]entity.Booking, error)
	GetBookingByIDFunc         func(id int) (*entity.Booking, error)
	SetBookingCompleteFunc     func(bookingID int) (*entity.Booking, error)
	SetBookingCancelFunc       func(bookingID, userID int) (*entity.Booking, error)
	GetAllFunc                 func() ([]entity.Booking, error)
}

type StubOutboxBookingRepository struct {
	CreateEventFunc func(event *entity.OutboxEvent) error
}

func (m *StubBookingRepository) Create(booking *entity.Booking) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(booking)
	}
	return domain.ErrNotImplemented
}

func (m *StubBookingRepository) IsSlotAvailable(start, end time.Time) (bool, error) {
	if m.IsSlotAvailableFunc != nil {
		return m.IsSlotAvailableFunc(start, end)
	}
	return false, domain.ErrNotImplemented
}

func (m *StubBookingRepository) GetAllBookingsByUserID(id int) ([]entity.Booking, error) {
	if m.GetAllBookingsByUserIDFunc != nil {
		return m.GetAllBookingsByUserIDFunc(id)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubBookingRepository) GetBookingByID(id int) (*entity.Booking, error) {
	if m.GetBookingByIDFunc != nil {
		return m.GetBookingByIDFunc(id)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubBookingRepository) SetBookingComplete(bookingID int) (*entity.Booking, error) {
	if m.SetBookingCompleteFunc != nil {
		return m.SetBookingCompleteFunc(bookingID)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubBookingRepository) SetBookingCancel(bookingID, userID int) (*entity.Booking, error) {
	if m.SetBookingCancelFunc != nil {
		return m.SetBookingCancelFunc(bookingID, userID)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubBookingRepository) GetAll() ([]entity.Booking, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubOutboxBookingRepository) CreateEvent(event *entity.OutboxEvent) error {
	if m.CreateEventFunc != nil {
		return m.CreateEventFunc(event)
	}
	return nil
}

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

	testAdmin = &entity.User{
		ID:       2,
		Email:    "admin@example.com",
		Role:     enum.RoleAdmin,
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

func TestBookingUseCase_CreateBooking_PastTime(t *testing.T) {
	stubBookingRepo := &StubBookingRepository{}
	stubUserRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testClient, nil
		},
	}

	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)
	req := &entity.CreateBookingRequest{
		SlotStart:          time.Now().Add(-1 * time.Hour),
		ProblemDescription: "test",
	}

	_, err := useCase.CreateBooking(testClient.ID, req)

	if !errors.Is(err, domain.ErrInvalidTimeRange) {
		t.Errorf("Ожидалась ошибка ErrInvalidTimeRange, получена: %v", err)
	}
}

func TestBookingUseCase_CreateBooking_SlotNotAvailable(t *testing.T) {
	stubBookingRepo := &StubBookingRepository{
		IsSlotAvailableFunc: func(start, end time.Time) (bool, error) {
			return false, nil
		},
	}
	stubUserRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testClient, nil
		},
	}

	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)

	futureTime := time.Now().Add(24 * time.Hour)
	req := &entity.CreateBookingRequest{
		SlotStart:          futureTime,
		ProblemDescription: "test",
	}

	_, err := useCase.CreateBooking(testClient.ID, req)

	if !errors.Is(err, domain.ErrSlotNotAvailable) {
		t.Errorf("Ожидалась ошибка ErrSlotNotAvailable, получена: %v", err)
	}
}

func TestBookingUseCase_CreateBooking_UserNotFound(t *testing.T) {
	stubBookingRepo := &StubBookingRepository{}
	stubUserRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)
	req := &entity.CreateBookingRequest{
		SlotStart:          time.Now().Add(24 * time.Hour),
		ProblemDescription: "test",
	}

	_, err := useCase.CreateBooking(999, req)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestBookingUseCase_CreateBooking_Success(t *testing.T) {
	futureTime := time.Now().Add(24 * time.Hour) // завтра

	stubBookingRepo := &StubBookingRepository{
		IsSlotAvailableFunc: func(start, end time.Time) (bool, error) {
			return true, nil
		},
		CreateFunc: func(booking *entity.Booking) error {
			return nil
		},
	}

	stubUserRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testClient, nil
		},
	}

	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)

	req := &entity.CreateBookingRequest{
		SlotStart:          futureTime,
		ProblemDescription: "Тестовая проблема",
	}

	booking, err := useCase.CreateBooking(testClient.ID, req)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if booking == nil {
		t.Fatal("Запись не была создана")
	}
}

func TestBookingUseCase_GetMyBookings_Success(t *testing.T) {
	expectedBookings := []entity.Booking{*testBooking}

	stubBookingRepo := &StubBookingRepository{
		GetAllBookingsByUserIDFunc: func(id int) ([]entity.Booking, error) {
			if id != testClient.ID {
				t.Errorf("Неверный UserID: %d", id)
			}
			return expectedBookings, nil
		},
	}
	stubUserRepo := &StubUserRepository{} // не используется в этом методе

	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)
	_, err := useCase.GetMyBookings(testClient.ID)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestBookingUseCase_GetMyBookings_Error(t *testing.T) {
	dbError := errors.New("database error")
	stubBookingRepo := &StubBookingRepository{
		GetAllBookingsByUserIDFunc: func(id int) ([]entity.Booking, error) {
			return nil, dbError
		},
	}
	stubUserRepo := &StubUserRepository{}

	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)
	_, err := useCase.GetMyBookings(testClient.ID)

	if err == nil {
		t.Fatal("Ожидалась ошибка, получена nil")
	}
	if !errors.Is(err, dbError) {
		t.Errorf("Ожидалась оригинальная ошибка, получена: %v", err)
	}
}

func TestBookingUseCase_CancelMyBooking_BookingNotFound(t *testing.T) {
	stubBookingRepo := &StubBookingRepository{
		GetBookingByIDFunc: func(id int) (*entity.Booking, error) {
			return nil, domain.ErrBookingNotFound
		},
	}
	stubUserRepo := &StubUserRepository{}

	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)
	err := useCase.CancelMyBooking(999, testClient.ID)

	if !errors.Is(err, domain.ErrBookingNotFound) {
		t.Errorf("Ожидалась ошибка ErrBookingNotFound, получена: %v", err)
	}
}

func TestBookingUseCase_CancelMyBooking_BookingAlreadyActive(t *testing.T) {
	stubBookingRepo := &StubBookingRepository{
		GetBookingByIDFunc: func(id int) (*entity.Booking, error) {
			return testBookingPast, nil
		},
	}
	stubUserRepo := &StubUserRepository{}

	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)
	err := useCase.CancelMyBooking(10, testClient.ID)

	if !errors.Is(err, domain.ErrBookingAlreadyActive) {
		t.Errorf("Ожидалась ошибка ErrBookingAlreadyActive, получена: %v", err)
	}

}

func TestBookingUseCase_CancelMyBooking_Success(t *testing.T) {
	stubBookingRepo := &StubBookingRepository{
		GetBookingByIDFunc: func(id int) (*entity.Booking, error) {
			return testBooking, nil
		},
		SetBookingCancelFunc: func(bookingID, userID int) (*entity.Booking, error) {
			return &entity.Booking{}, nil
		},
	}
	stubUserRepo := &StubUserRepository{}

	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)
	err := useCase.CancelMyBooking(10, testClient.ID)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestBookingUseCase_CreateBooking_OverlappingSlots(t *testing.T) {
	existingStart := time.Now().Add(24 * time.Hour)
	existingEnd := time.Now().Add(25 * time.Hour)

	newStart := time.Now().Add(24 * time.Hour).Add(30 * time.Minute)

	stubBookingRepo := &StubBookingRepository{
		IsSlotAvailableFunc: func(start, end time.Time) (bool, error) {
			if existingStart.Before(end) && start.Before(existingEnd) {
				return false, nil
			}
			return true, nil
		},
	}
	stubUserRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testClient, nil
		},
	}

	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)
	req := &entity.CreateBookingRequest{
		SlotStart:          newStart,
		ProblemDescription: "Тест пересечения",
	}

	_, err := useCase.CreateBooking(testClient.ID, req)

	if !errors.Is(err, domain.ErrSlotNotAvailable) {
		t.Errorf("Ожидалась ошибка ErrSlotNotAvailable для пересекающегося слота, получена: %v", err)
	}
}

func TestBookingUseCase_CompleteBooking_BookingNotFound(t *testing.T) {
	stubBookingRepo := &StubBookingRepository{
		GetBookingByIDFunc: func(id int) (*entity.Booking, error) {
			return nil, domain.ErrBookingNotFound
		},
	}
	stubUserRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testStaff, nil
		},
	}

	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)
	_, err := useCase.CompleteBookingByID(100)

	if !errors.Is(err, domain.ErrBookingNotFound) {
		t.Errorf("Ожидалась ошибка ErrBookingNotFound, получена: %v", err)
	}
}

func TestBookingUseCase_CompleteBooking_NotFinishedYet(t *testing.T) {
	stubBookingRepo := &StubBookingRepository{
		GetBookingByIDFunc: func(id int) (*entity.Booking, error) {
			return testBooking, nil
		},
	}
	stubUserRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testStaff, nil
		},
	}

	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)
	_, err := useCase.CompleteBookingByID(101)

	if !errors.Is(err, domain.ErrBookingNotFinished) {
		t.Errorf("Ожидалась ошибка ErrBookingNotFinished, получена: %v", err)
	}
}

func TestBookingUseCase_CompleteBooking_Success(t *testing.T) {
	stubBookingRepo := &StubBookingRepository{
		GetBookingByIDFunc: func(id int) (*entity.Booking, error) {
			return testBookingPast, nil
		},
		SetBookingCompleteFunc: func(bookingID int) (*entity.Booking, error) {
			updated := *testBookingPast
			updated.Status = enum.StatusCompleted
			return &updated, nil
		},
	}
	stubUserRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testStaff, nil
		},
	}

	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)
	result, err := useCase.CompleteBookingByID(100)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if result.Status != enum.StatusCompleted {
		t.Errorf("Статус не обновлен: %s", result.Status)
	}
}

func TestBookingUseCase_GetAllBookings_OnlyForStaffOrAdmin(t *testing.T) {
	stubUserRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testClient, nil
		},
	}

	stubBookingRepo := &StubBookingRepository{
		GetAllFunc: func() ([]entity.Booking, error) {
			return []entity.Booking{}, nil
		},
	}

	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)

	_, err := useCase.GetAllBookings(testClient.ID)

	if !errors.Is(err, domain.ErrOnlyForStaffOrAdmin) {
		t.Errorf("Ожидалась ошибка ErrOnlyForStaffOrAdmin, получена: %v", err)
	}
}

func TestBookingUseCase_GetAllBookings_UserNotFound(t *testing.T) {
	stubUserRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubBookingRepo := &StubBookingRepository{}
	stubNotifier := &StubOutboxBookingRepository{}
	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)

	_, err := useCase.GetAllBookings(999)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestBookingUseCase_GetAllBookings_StaffSuccess(t *testing.T) {
	expectedBookings := []entity.Booking{*testBooking, *testBookingPast}

	stubUserRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testStaff, nil
		},
	}

	stubBookingRepo := &StubBookingRepository{
		GetAllFunc: func() ([]entity.Booking, error) {
			return expectedBookings, nil
		},
	}

	stubNotifier := &StubOutboxBookingRepository{}

	useCase := usecase.NewBookingUseCase(stubBookingRepo, stubUserRepo, stubNotifier)

	_, err := useCase.GetAllBookings(testStaff.ID)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}
