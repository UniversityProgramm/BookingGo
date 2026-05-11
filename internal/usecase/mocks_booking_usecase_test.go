package usecase

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"errors"
	"testing"
	"time"
)

type MockBookingRepository struct {
	CreateFunc              func(booking *entity.Booking) error
	IsSlotAvailableFunc     func(start, end time.Time) (bool, error)
	GetBookingsByUserIDFunc func(id int) ([]entity.Booking, error)
	GetBookingByIDFunc      func(id int) (*entity.Booking, error)
	DeleteBookingByIDFunc   func(bookingID, userID int) error
	SetBookingStatusFunc    func(bookingID int, newStatus enum.Status) (*entity.Booking, error)
}

func (m *MockBookingRepository) Create(booking *entity.Booking) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(booking)
	}
	return domain.ErrNotImplemented
}

func (m *MockBookingRepository) IsSlotAvailable(start, end time.Time) (bool, error) {
	if m.IsSlotAvailableFunc != nil {
		return m.IsSlotAvailableFunc(start, end)
	}
	return false, domain.ErrNotImplemented
}

func (m *MockBookingRepository) GetBookingsByUserID(id int) ([]entity.Booking, error) {
	if m.GetBookingsByUserIDFunc != nil {
		return m.GetBookingsByUserIDFunc(id)
	}
	return nil, domain.ErrNotImplemented
}

func (m *MockBookingRepository) GetBookingByID(id int) (*entity.Booking, error) {
	if m.GetBookingByIDFunc != nil {
		return m.GetBookingByIDFunc(id)
	}
	return nil, domain.ErrNotImplemented
}

func (m *MockBookingRepository) DeleteBookingByID(bookingID, userID int) error {
	if m.DeleteBookingByIDFunc != nil {
		return m.DeleteBookingByIDFunc(bookingID, userID)
	}
	return domain.ErrNotImplemented
}

func (m *MockBookingRepository) SetBookingStatus(bookingID int, newStatus enum.Status) (*entity.Booking, error) {
	if m.SetBookingStatusFunc != nil {
		return m.SetBookingStatusFunc(bookingID, newStatus)
	}
	return nil, domain.ErrNotImplemented
}

var (
	testClient = &entity.User{
		ID:       1,
		Email:    "client@example.com",
		Role:     enum.RoleClient,
		IsActive: true,
	}

	testStaff = &entity.User{
		ID:       3,
		Email:    "staff@mail.ru",
		Role:     enum.RoleStaff,
		IsActive: true,
	}

	testAdmin = &entity.User{
		ID:       2,
		Email:    "admin@example.com",
		Role:     enum.RoleAdmin,
		IsActive: true,
	}

	testBooking = &entity.Booking{
		ID:                 10,
		UserID:             1,
		SlotStart:          time.Now(),
		SlotEnd:            time.Now().Add(time.Hour),
		Status:             enum.StatusConfirmed,
		ProblemDescription: "problem",
	}

	testBookingPast = &entity.Booking{
		ID:                 10,
		UserID:             1,
		SlotStart:          time.Now().Add(-2 * time.Hour),
		SlotEnd:            time.Now().Add(-1 * time.Hour),
		Status:             enum.StatusConfirmed,
		ProblemDescription: "problem",
	}
)

func TestBookingUseCase_CreateBooking_NotClient(t *testing.T) {
	mockBookingRepo := &MockBookingRepository{}
	mockUserRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testAdmin, nil
		},
	}

	useCase := NewBookingUseCase(mockBookingRepo, mockUserRepo)
	req := &entity.CreateBookingRequest{
		SlotStart:          time.Now().Add(24 * time.Hour),
		ProblemDescription: "test",
	}

	_, err := useCase.CreateBooking(testAdmin.ID, req)

	if !errors.Is(err, domain.ErrOnlyForClient) {
		t.Errorf("Ожидалась ошибка ErrOnlyForClient, получена: %v", err)
	}
}

func TestBookingUseCase_CreateBooking_PastTime(t *testing.T) {
	mockBookingRepo := &MockBookingRepository{}
	mockUserRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testClient, nil
		},
	}

	useCase := NewBookingUseCase(mockBookingRepo, mockUserRepo)
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
	mockBookingRepo := &MockBookingRepository{
		IsSlotAvailableFunc: func(start, end time.Time) (bool, error) {
			return false, nil
		},
	}
	mockUserRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testClient, nil
		},
	}

	useCase := NewBookingUseCase(mockBookingRepo, mockUserRepo)

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
	mockBookingRepo := &MockBookingRepository{}
	mockUserRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	useCase := NewBookingUseCase(mockBookingRepo, mockUserRepo)
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

	mockBookingRepo := &MockBookingRepository{
		IsSlotAvailableFunc: func(start, end time.Time) (bool, error) {
			return true, nil
		},
		CreateFunc: func(booking *entity.Booking) error {
			return nil
		},
	}

	mockUserRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testClient, nil
		},
	}

	useCase := NewBookingUseCase(mockBookingRepo, mockUserRepo)

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

	mockBookingRepo := &MockBookingRepository{
		GetBookingsByUserIDFunc: func(id int) ([]entity.Booking, error) {
			if id != testClient.ID {
				t.Errorf("Неверный UserID: %d", id)
			}
			return expectedBookings, nil
		},
	}
	mockUserRepo := &MockUserRepository{} // не используется в этом методе

	useCase := NewBookingUseCase(mockBookingRepo, mockUserRepo)
	_, err := useCase.GetMyBookings(testClient.ID)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestBookingUseCase_GetMyBookings_Error(t *testing.T) {
	dbError := errors.New("database error")
	mockBookingRepo := &MockBookingRepository{
		GetBookingsByUserIDFunc: func(id int) ([]entity.Booking, error) {
			return nil, dbError
		},
	}
	mockUserRepo := &MockUserRepository{}

	useCase := NewBookingUseCase(mockBookingRepo, mockUserRepo)
	_, err := useCase.GetMyBookings(testClient.ID)

	if err == nil {
		t.Fatal("Ожидалась ошибка, получена nil")
	}
	if !errors.Is(err, dbError) {
		t.Errorf("Ожидалась оригинальная ошибка, получена: %v", err)
	}
}

func TestBookingUseCase_DeleteMyBooking_BookingNotFound(t *testing.T) {
	mockBookingRepo := &MockBookingRepository{
		DeleteBookingByIDFunc: func(bookingID, userID int) error {
			return domain.ErrBookingNotFound
		},
	}
	mockUserRepo := &MockUserRepository{}

	useCase := NewBookingUseCase(mockBookingRepo, mockUserRepo)
	err := useCase.DeleteMyBooking(999, testClient.ID)

	if !errors.Is(err, domain.ErrBookingNotFound) {
		t.Errorf("Ожидалась ошибка ErrBookingNotFound, получена: %v", err)
	}
}

func TestBookingUseCase_DeleteMyBooking_Success(t *testing.T) {
	mockBookingRepo := &MockBookingRepository{
		DeleteBookingByIDFunc: func(bookingID, userID int) error {
			return nil
		},
	}
	mockUserRepo := &MockUserRepository{}

	useCase := NewBookingUseCase(mockBookingRepo, mockUserRepo)
	err := useCase.DeleteMyBooking(100, testClient.ID)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestBookingUseCase_CreateBooking_OverlappingSlots(t *testing.T) {
	existingStart := time.Now().Add(24 * time.Hour)
	existingEnd := time.Now().Add(25 * time.Hour)

	newStart := time.Now().Add(24 * time.Hour).Add(30 * time.Minute)

	mockBookingRepo := &MockBookingRepository{
		IsSlotAvailableFunc: func(start, end time.Time) (bool, error) {
			if existingStart.Before(end) && start.Before(existingEnd) {
				return false, nil
			}
			return true, nil
		},
	}
	mockUserRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testClient, nil
		},
	}

	useCase := NewBookingUseCase(mockBookingRepo, mockUserRepo)
	req := &entity.CreateBookingRequest{
		SlotStart:          newStart,
		ProblemDescription: "Тест пересечения",
	}

	_, err := useCase.CreateBooking(testClient.ID, req)

	if !errors.Is(err, domain.ErrSlotNotAvailable) {
		t.Errorf("Ожидалась ошибка ErrSlotNotAvailable для пересекающегося слота, получена: %v", err)
	}
}

func TestBookingUseCase_CompleteBooking_NotStaff(t *testing.T) {
	mockBookingRepo := &MockBookingRepository{}
	mockUserRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testClient, nil
		},
	}

	useCase := NewBookingUseCase(mockBookingRepo, mockUserRepo)
	_, err := useCase.ChangeBookingStatus(100, 2)

	if !errors.Is(err, domain.ErrOnlyForStaff) {
		t.Errorf("Ожидалась ошибка ErrOnlyForStaff, получена: %v", err)
	}
}

func TestBookingUseCase_CompleteBooking_BookingNotFound(t *testing.T) {
	mockBookingRepo := &MockBookingRepository{
		GetBookingByIDFunc: func(id int) (*entity.Booking, error) {
			return nil, domain.ErrBookingNotFound
		},
	}
	mockUserRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testStaff, nil
		},
	}

	useCase := NewBookingUseCase(mockBookingRepo, mockUserRepo)
	_, err := useCase.ChangeBookingStatus(100, 3)

	if !errors.Is(err, domain.ErrBookingNotFound) {
		t.Errorf("Ожидалась ошибка ErrBookingNotFound, получена: %v", err)
	}
}

func TestBookingUseCase_CompleteBooking_NotFinishedYet(t *testing.T) {
	mockBookingRepo := &MockBookingRepository{
		GetBookingByIDFunc: func(id int) (*entity.Booking, error) {
			return testBooking, nil
		},
	}
	mockUserRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testStaff, nil
		},
	}

	useCase := NewBookingUseCase(mockBookingRepo, mockUserRepo)
	_, err := useCase.ChangeBookingStatus(101, 1)

	if !errors.Is(err, domain.ErrBookingNotFinished) {
		t.Errorf("Ожидалась ошибка ErrBookingNotFinished, получена: %v", err)
	}
}

func TestBookingUseCase_CompleteBooking_Success(t *testing.T) {
	mockBookingRepo := &MockBookingRepository{
		GetBookingByIDFunc: func(id int) (*entity.Booking, error) {
			return testBookingPast, nil
		},
		SetBookingStatusFunc: func(bookingID int, completeStatus enum.Status) (*entity.Booking, error) {
			updated := *testBookingPast
			updated.Status = completeStatus
			return &updated, nil
		},
	}
	mockUserRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testStaff, nil
		},
	}

	useCase := NewBookingUseCase(mockBookingRepo, mockUserRepo)
	result, err := useCase.ChangeBookingStatus(100, 1)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if result.Status != enum.StatusCompleted {
		t.Errorf("Статус не обновлен: %s", result.Status)
	}
}
