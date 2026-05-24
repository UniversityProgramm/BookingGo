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

type StubUserRepository struct {
	GetAllFunc      func() ([]entity.User, error)
	GetByIDFunc     func(id int) (*entity.User, error)
	GetByEmailFunc  func(email string) (*entity.User, error)
	CreateFunc      func(user *entity.User) error
	UpdateFunc      func(id int, requestUser map[string]interface{}) (*entity.User, error)
	DeleteFunc      func(id int) error
	EmailExistsFunc func(email string) (bool, error)
}

func (m *StubUserRepository) GetAll() ([]entity.User, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return []entity.User{}, nil
}

func (m *StubUserRepository) GetByID(id int) (*entity.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserRepository) GetByEmail(email string) (*entity.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(email)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserRepository) Create(user *entity.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return domain.ErrNotImplemented
}

func (m *StubUserRepository) Update(id int, requestUser map[string]interface{}) (*entity.User, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(id, requestUser)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserRepository) Delete(id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return domain.ErrNotImplemented
}

func (m *StubUserRepository) EmailExists(email string) (bool, error) {
	if m.EmailExistsFunc != nil {
		return m.EmailExistsFunc(email)
	}
	return false, domain.ErrNotImplemented
}

// Тесты
var testUser = &entity.User{
	ID:           1,
	Email:        "test@example.com",
	PasswordHash: "hashed_password",
	FIO:          "Test User",
	Phone:        "+79991234567",
	Role:         enum.RoleClient,
	IsActive:     true,
	CreatedAt:    time.Now(),
	UpdatedAt:    time.Now(),
	UserNotificationSettings: entity.NotificationSettings{
		IsEmailSend: true,
		IsPhoneSend: false,
	},
}

func TestUserUseCase_GetAllUsers_Success(t *testing.T) {
	stubRepo := &StubUserRepository{
		GetAllFunc: func() ([]entity.User, error) {
			return []entity.User{*testUser}, nil
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.GetAllUsers()
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_GetUserByID_UserNotFound(t *testing.T) {
	stubRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.GetUserByID(1)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestUserUseCase_GetUserByID_Success(t *testing.T) {
	stubRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testUser, nil
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.GetUserByID(1)
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_GetUserByEmail_UserNotFound(t *testing.T) {
	stubRepo := &StubUserRepository{
		GetByEmailFunc: func(email string) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.GetUserByEmail("email")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestUserUseCase_GetUserByEmail_Success(t *testing.T) {
	stubRepo := &StubUserRepository{
		GetByEmailFunc: func(email string) (*entity.User, error) {
			return testUser, nil
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.GetUserByEmail("email@mail.ru")
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_CreateUser_EmailTaken(t *testing.T) {
	stubRepo := &StubUserRepository{
		EmailExistsFunc: func(email string) (bool, error) {
			return true, nil
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.CreateUser(&entity.CreateUserRequest{
		Email:    "takenemail@mail.ru",
		Password: "123456",
		FIO:      "Test user",
		Phone:    "123123123",
	})

	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("Ожидалась ошибка ErrEmailTaken, получена: %v", err)
	}
}

func TestUserUseCase_CreateUser_Success(t *testing.T) {
	stubRepo := &StubUserRepository{
		EmailExistsFunc: func(email string) (bool, error) {
			return false, nil
		},
		CreateFunc: func(user *entity.User) error {
			if len(user.PasswordHash) == 0 {
				t.Error("Пароль не был захэширован")
			}
			return nil
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	_, err := useCase.CreateUser(&entity.CreateUserRequest{
		Email:    "new@mail.ru",
		Password: "qwerty",
		FIO:      "New user",
		Phone:    "123123123",
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_UpdateUser_UserNotFound(t *testing.T) {
	stubRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)
	req := &entity.UpdateUserRequest{
		FIO:   new("new test fio"),
		Phone: new("123123123"),
	}

	_, err := useCase.UpdateUserProfile(1, req)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestUserUseCase_UpdateUser_Success(t *testing.T) {
	updatedFIO := "Updated Name"
	updatedPhone := "+79999999999"

	stubRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testUser, nil
		},
		UpdateFunc: func(id int, updates map[string]interface{}) (*entity.User, error) {
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
			// Возвращаем обновлённого пользователя
			updatedUser := *testUser
			updatedUser.FIO = updatedFIO
			updatedUser.Phone = updatedPhone
			return &updatedUser, nil
		},
	}

	useCase := usecase.NewUserUseCase(stubRepo)
	req := &entity.UpdateUserRequest{
		FIO:   &updatedFIO,
		Phone: &updatedPhone,
	}

	_, err := useCase.UpdateUserProfile(1, req)
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_DeleteUser_UserNotFound(t *testing.T) {
	stubRepo := &StubUserRepository{
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

func TestUserUseCase_DeleteUser_Success(t *testing.T) {
	stubRepo := &StubUserRepository{
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

func TestUserUseCase_UpdateUserData_UserNotFound(t *testing.T) {
	stubRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
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

func TestUserUseCase_UpdateUserData_Success(t *testing.T) {
	stubRepo := &StubUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testUser, nil
		},
		UpdateFunc: func(id int, updates map[string]any) (*entity.User, error) {
			// Проверяем, что переданы правильные данные
			if fio, ok := updates["fio"]; ok {
				if fio != "Updated FIO" {
					t.Errorf("Неверное значение FIO: %v", fio)
				}
			}
			// Возвращаем обновлённого пользователя
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

func TestUserUseCase_EmailExists_True(t *testing.T) {
	stubRepo := &StubUserRepository{
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

func TestUserUseCase_EmailExists_False(t *testing.T) {
	stubRepo := &StubUserRepository{
		EmailExistsFunc: func(email string) (bool, error) {
			return false, nil
		},
	}
	useCase := usecase.NewUserUseCase(stubRepo)

	exists, err := useCase.EmailExists("free@example.com")
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if exists {
		t.Error("Ожидалось, что email свободен")
	}
}

func TestUserUseCase_EmailExists_Error(t *testing.T) {
	dbErr := errors.New("database error")
	stubRepo := &StubUserRepository{
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
