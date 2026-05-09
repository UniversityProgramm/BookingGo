package usecase

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	GetAllFunc         func() ([]entity.User, error)
	GetByIDFunc        func(id int) (*entity.User, error)
	GetByEmailFunc     func(email string) (*entity.User, error)
	CreateFunc         func(user *entity.User) error
	UpdateFunc         func(id int, requestUser map[string]interface{}) (*entity.User, error)
	DeleteFunc         func(id int) error
	EmailExistsFunc    func(email string) (bool, error)
	ChangePasswordFunc func(id int, newPassword string) error
}

func (m *MockUserRepository) GetAll() ([]entity.User, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return []entity.User{}, nil
}

func (m *MockUserRepository) GetByID(id int) (*entity.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, domain.ErrNotImplemented
}

func (m *MockUserRepository) GetByEmail(email string) (*entity.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(email)
	}
	return nil, domain.ErrNotImplemented
}

func (m *MockUserRepository) Create(user *entity.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return domain.ErrNotImplemented
}

func (m *MockUserRepository) Update(id int, requestUser map[string]interface{}) (*entity.User, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(id, requestUser)
	}
	return nil, domain.ErrNotImplemented
}

func (m *MockUserRepository) Delete(id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return domain.ErrNotImplemented
}

func (m *MockUserRepository) EmailExists(email string) (bool, error) {
	if m.EmailExistsFunc != nil {
		return m.EmailExistsFunc(email)
	}
	return false, domain.ErrNotImplemented
}

func (m *MockUserRepository) ChangePassword(id int, newPassword string) error {
	if m.ChangePasswordFunc != nil {
		return m.ChangePassword(id, newPassword)
	}
	return domain.ErrNotImplemented
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
}

func TestUserUseCase_GetAllUsers_Success(t *testing.T) {
	mockRepo := &MockUserRepository{
		GetAllFunc: func() ([]entity.User, error) {
			return []entity.User{*testUser}, nil
		},
	}
	useCase := NewUserUseCase(mockRepo)

	_, err := useCase.GetAllUsers()
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_GetUserByID_UserNotFound(t *testing.T) {
	mockRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	useCase := NewUserUseCase(mockRepo)

	_, err := useCase.GetUserByID(1)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestUserUseCase_GetUserByID_Success(t *testing.T) {
	mockRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testUser, nil
		},
	}
	useCase := NewUserUseCase(mockRepo)

	_, err := useCase.GetUserByID(1)
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_GetUserByEmail_UserNotFound(t *testing.T) {
	mockRepo := &MockUserRepository{
		GetByEmailFunc: func(email string) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	useCase := NewUserUseCase(mockRepo)

	_, err := useCase.GetUserByEmail("email")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestUserUseCase_GetUserByEmail_Success(t *testing.T) {
	mockRepo := &MockUserRepository{
		GetByEmailFunc: func(email string) (*entity.User, error) {
			return testUser, nil
		},
	}
	useCase := NewUserUseCase(mockRepo)

	_, err := useCase.GetUserByEmail("email@mail.ru")
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_CreateUser_EmailTaken(t *testing.T) {
	mockRepo := &MockUserRepository{
		EmailExistsFunc: func(email string) (bool, error) {
			return true, nil
		},
	}
	useCase := NewUserUseCase(mockRepo)

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
	mockRepo := &MockUserRepository{
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
	useCase := NewUserUseCase(mockRepo)

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
	mockRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	useCase := NewUserUseCase(mockRepo)
	req := &entity.UpdateUserRequest{
		FIO:   new("new test fio"),
		Phone: new("123123123"),
	}

	_, err := useCase.UpdateUser(1, req)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestUserUseCase_UpdateUser_Success(t *testing.T) {
	updatedFIO := "Updated Name"
	updatedPhone := "+79999999999"

	mockRepo := &MockUserRepository{
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

	useCase := NewUserUseCase(mockRepo)
	req := &entity.UpdateUserRequest{
		FIO:   &updatedFIO,
		Phone: &updatedPhone,
	}

	_, err := useCase.UpdateUser(1, req)
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_DeleteUser_UserNotFound(t *testing.T) {
	mockRepo := &MockUserRepository{
		DeleteFunc: func(id int) error {
			return domain.ErrUserNotFound
		},
	}
	useCase := NewUserUseCase(mockRepo)

	err := useCase.DeleteUser(1)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestUserUseCase_DeleteUser_Success(t *testing.T) {
	mockRepo := &MockUserRepository{
		DeleteFunc: func(id int) error {
			return nil
		},
	}
	useCase := NewUserUseCase(mockRepo)

	err := useCase.DeleteUser(1)
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestUserUseCase_ChangePassword_UserNotFound(t *testing.T) {
	mockRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	useCase := NewUserUseCase(mockRepo)
	newPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}

	err = useCase.ChangePassword(1, string(newPassword))
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestUserUseCase_ChangePassword_Success(t *testing.T) {
	mockRepo := &MockUserRepository{
		GetByIDFunc: func(id int) (*entity.User, error) {
			return testUser, nil
		},
	}
	useCase := NewUserUseCase(mockRepo)
	newPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}

	err = useCase.ChangePassword(1, string(newPassword))
	if errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}
