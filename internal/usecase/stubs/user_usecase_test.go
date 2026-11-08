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

type StubUserPostgres struct {
	GetAllFunc            func() ([]domain.User, error)
	GetByIDFunc           func(id int) (*domain.User, error)
	GetByEmailFunc        func(email string) (*domain.User, error)
	CreateFunc            func(user *domain.User) error
	UpdateFunc            func(id int, requestUser map[string]interface{}) (*domain.User, error)
	DeleteFunc            func(id int) error
	EmailExistsFunc       func(email string) (bool, error)
	GetByEmailContextFunc func(ctx context.Context, email string) (*domain.User, error)
	GetByIDContextFunc    func(ctx context.Context, id int) (*domain.User, error)
}

func (m *StubUserPostgres) GetAll() ([]domain.User, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return []domain.User{}, nil
}

func (m *StubUserPostgres) GetByID(id int) (*domain.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserPostgres) GetByEmail(email string) (*domain.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(email)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserPostgres) Create(user *domain.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return domain.ErrNotImplemented
}

func (m *StubUserPostgres) Update(id int, requestUser map[string]interface{}) (*domain.User, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(id, requestUser)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserPostgres) Delete(id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return domain.ErrNotImplemented
}

func (m *StubUserPostgres) EmailExists(email string) (bool, error) {
	if m.EmailExistsFunc != nil {
		return m.EmailExistsFunc(email)
	}
	return false, domain.ErrNotImplemented
}

func (m *StubUserPostgres) GetByIDContext(ctx context.Context, id int) (*domain.User, error) {
	if m.GetByIDContextFunc != nil {
		return m.GetByIDContextFunc(ctx, id)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserPostgres) GetByEmailContext(ctx context.Context, email string) (*domain.User, error) {
	if m.GetByEmailContextFunc != nil {
		return m.GetByEmailContextFunc(ctx, email)
	}
	return nil, domain.ErrNotImplemented
}

// Тесты
var testUser = &domain.User{
	ID:           1,
	Email:        "test@example.com",
	PasswordHash: "hashed_password",
	FIO:          "Test User",
	Phone:        "+79991234567",
	Role:         enum.RoleClient,
	IsActive:     true,
	CreatedAt:    time.Now(),
	UpdatedAt:    time.Now(),
	UserNotificationSettings: domain.NotificationSettings{
		IsEmailSend: true,
		IsPhoneSend: false,
	},
}

func TestUserUseCase_GetAllUsers_Success(t *testing.T) {
	stubRepo := &StubUserPostgres{
		GetAllFunc: func() ([]domain.User, error) {
			return []domain.User{*testUser}, nil
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.GetAllUsers()
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_GetUserByID_Success(t *testing.T) {
	stubRepo := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testUser, nil
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.GetUserByID(1)
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_GetUserByID_UserNotFound(t *testing.T) {
	stubRepo := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.GetUserByID(1)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestUserUseCase_GetUserByEmail_Success(t *testing.T) {
	stubRepo := &StubUserPostgres{
		GetByEmailFunc: func(email string) (*domain.User, error) {
			return testUser, nil
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.GetUserByEmail("email@mail.ru")
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_GetUserByEmail_UserNotFound(t *testing.T) {
	stubRepo := &StubUserPostgres{
		GetByEmailFunc: func(email string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.GetUserByEmail("email")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestUserUseCase_GetUserByEmailContext_Success(t *testing.T) {
	stubRepo := &StubUserPostgres{
		GetByEmailContextFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return testUser, nil
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.GetUserByEmailContext(context.Background(), "email@mail.ru")
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_GetUserByEmailContext_UserNotFound(t *testing.T) {
	stubRepo := &StubUserPostgres{
		GetByEmailContextFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.GetUserByEmailContext(context.Background(), "email")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestUserUseCase_CreateUser_Success(t *testing.T) {
	stubRepo := &StubUserPostgres{
		EmailExistsFunc: func(email string) (bool, error) {
			return false, nil
		},
		CreateFunc: func(user *domain.User) error {
			if len(user.PasswordHash) == 0 {
				t.Error("Пароль не был захэширован")
			}
			return nil
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.CreateUser(&dto.CreateUserRequest{
		Email:    "new@mail.ru",
		Password: "qwerty",
		FIO:      "New user",
		Phone:    "123123123",
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_CreateUser_EmailTaken(t *testing.T) {
	stubRepo := &StubUserPostgres{
		EmailExistsFunc: func(email string) (bool, error) {
			return true, nil
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.CreateUser(&dto.CreateUserRequest{
		Email:    "takenemail@mail.ru",
		Password: "123456",
		FIO:      "Test user",
		Phone:    "123123123",
	})

	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("Ожидалась ошибка ErrEmailTaken, получена: %v", err)
	}
}

func TestUserUseCase_UpdateUserProfile_Success(t *testing.T) {
	updatedFIO := "Updated Name"
	updatedPhone := "+79999999999"

	stubRepo := &StubUserPostgres{
		UpdateFunc: func(id int, updates map[string]interface{}) (*domain.User, error) {
			if fio, ok := updates["fio"]; ok {
				if fio != updatedFIO {
					t.Errorf("Неверное значение FIO: %v", fio)
				}
			}
			if phone, ok := updates["phone"]; ok {
				if phone != updatedPhone {
					t.Errorf("Неверное значение Phone: %v", phone)
				}
			}
			updatedUser := *testUser
			updatedUser.FIO = updatedFIO
			updatedUser.Phone = updatedPhone
			return &updatedUser, nil
		},
	}

	useCase := usecase.NewUserUseCase(stubRepo)
	req := &dto.UpdateUserProfileRequest{
		FIO:   &updatedFIO,
		Phone: &updatedPhone,
	}

	_, err := useCase.UpdateUserProfile(1, req)
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_UpdateUserProfile_UserNotFound(t *testing.T) {
	fio := "new test fio"
	phone := "123123123"
	stubRepo := &StubUserPostgres{
		UpdateFunc: func(id int, updates map[string]interface{}) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)
	req := &dto.UpdateUserProfileRequest{
		FIO:   &fio,
		Phone: &phone,
	}

	_, err := useCase.UpdateUserProfile(1, req)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestUserUseCase_UpdateUserData_Success(t *testing.T) {
	stubRepo := &StubUserPostgres{
		UpdateFunc: func(id int, updates map[string]any) (*domain.User, error) {
			if fio, ok := updates["fio"]; ok {
				if fio != "Updated FIO" {
					t.Errorf("Неверное значение FIO: %v", fio)
				}
			}
			updated := *testUser
			updated.FIO = "Updated FIO"
			return &updated, nil
		},
	}

	useCase := usecase.NewUserUseCase(stubRepo)
	updates := map[string]any{"fio": "Updated FIO"}

	result, err := useCase.UpdateUserData(1, updates)
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if result.FIO != "Updated FIO" {
		t.Errorf("FIO не обновлён: %s", result.FIO)
	}
}

func TestUserUseCase_UpdateUserData_UserNotFound(t *testing.T) {
	stubRepo := &StubUserPostgres{
		UpdateFunc: func(id int, updates map[string]any) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	updates := map[string]any{"fio": "New Name"}
	_, err := useCase.UpdateUserData(999, updates)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestUserUseCase_DeleteUser_Success(t *testing.T) {
	stubRepo := &StubUserPostgres{
		DeleteFunc: func(id int) error {
			return nil
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	err := useCase.DeleteUser(1)
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_DeleteUser_UserNotFound(t *testing.T) {
	stubRepo := &StubUserPostgres{
		DeleteFunc: func(id int) error {
			return domain.ErrUserNotFound
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	err := useCase.DeleteUser(1)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestUserUseCase_EmailExists_Success(t *testing.T) {
	stubRepo := &StubUserPostgres{
		EmailExistsFunc: func(email string) (bool, error) {
			if email == "taken@example.com" {
				return true, nil
			}
			return false, nil
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	exists, err := useCase.EmailExists("taken@example.com")
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if !exists {
		t.Error("Ожидалось, что email занят")
	}
}

func TestUserUseCase_EmailExists_Error(t *testing.T) {
	dbErr := errors.New("database error")
	stubRepo := &StubUserPostgres{
		EmailExistsFunc: func(email string) (bool, error) {
			return false, dbErr
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.EmailExists("test@example.com")
	if err == nil {
		t.Fatal("Ожидалась ошибка БД, получена nil")
	}
	if err != dbErr {
		t.Errorf("Ожидалась оригинальная ошибка, получена: %v", err)
	}
}
