package usecase

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"errors"
	"testing"
	"time"
)

type MockUserUseCase struct {
	GetUserByEmailFunc func(email string) (*entity.User, error)
	CreateUserFunc     func(req *entity.CreateUserRequest) (*entity.User, error)
	GetUserByIDFunc    func(id int) (*entity.User, error)
	UpdateUserFunc     func(id int, req *entity.UpdateUserRequest) (*entity.User, error)
	ChangePasswordFunc func(id int, newPassword string) error
}

func (m *MockUserUseCase) GetUserByEmail(email string) (*entity.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(email)
	}
	return nil, domain.ErrNotImplemented
}

func (m *MockUserUseCase) CreateUser(req *entity.CreateUserRequest) (*entity.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(req)
	}
	return nil, domain.ErrNotImplemented
}

func (m *MockUserUseCase) GetUserByID(id int) (*entity.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(id)
	}
	return nil, domain.ErrNotImplemented
}

func (m *MockUserUseCase) UpdateUser(id int, req *entity.UpdateUserRequest) (*entity.User, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(id, req)
	}
	return nil, domain.ErrNotImplemented
}

func (m *MockUserUseCase) ChangePassword(id int, newPassword string) error {
	if m.ChangePasswordFunc != nil {
		return m.ChangePasswordFunc(id, newPassword)
	}
	return domain.ErrNotImplemented
}

var testAuthUser = &entity.User{
	ID:           1,
	Email:        "test@example.com",
	PasswordHash: "$2a$10$C1AtwB2zZtvq5llPqyEkY.YIDB28fNHrVUrcBN1F6CgEDb.YdU0qa", //password123
	FIO:          "Test User",
	Phone:        "+79991234567",
	Role:         enum.RoleClient,
	IsActive:     true,
	CreatedAt:    time.Now(),
	UpdatedAt:    time.Now(),
}

func TestAuthUseCase_Login_InvalidEmail(t *testing.T) {
	mockUserUseCase := &MockUserUseCase{
		GetUserByEmailFunc: func(email string) (*entity.User, error) {
			return nil, domain.ErrInvalidEmail
		},
	}

	authUseCase := NewAuthUseCase(mockUserUseCase)
	_, err := authUseCase.Login("notfound@example.com", "password123")
	if !errors.Is(err, domain.ErrInvalidEmail) {
		t.Errorf("Ожидалась ошибка ErrInvalidEmail, получена: %v", err)
	}
}

func TestAuthUseCase_Login_InvalidPassword(t *testing.T) {
	mockUserUseCase := &MockUserUseCase{
		GetUserByEmailFunc: func(email string) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	authUseCase := NewAuthUseCase(mockUserUseCase)
	_, err := authUseCase.Login("test@example.com", "wrong_password")
	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Errorf("Ожидалась ошибка ErrInvalidPassword, получена: %v", err)
	}
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	mockUserUseCase := &MockUserUseCase{
		GetUserByEmailFunc: func(email string) (*entity.User, error) {
			if email == "test@example.com" {
				return testAuthUser, nil
			}
			return nil, domain.ErrInvalidEmail
		},
	}

	authUseCase := NewAuthUseCase(mockUserUseCase)
	token, err := authUseCase.Login("test@example.com", "password123")

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if token == "" {
		t.Error("Токен не был сгенерирован")
	}
}

func TestAuthUseCase_Register_EmailTaken(t *testing.T) {
	mockUserUseCase := &MockUserUseCase{
		CreateUserFunc: func(req *entity.CreateUserRequest) (*entity.User, error) {
			return nil, domain.ErrEmailTaken
		},
	}

	authUseCase := NewAuthUseCase(mockUserUseCase)

	_, err := authUseCase.Register(&entity.CreateUserRequest{
		Email:    "taken@example.com",
		Password: "qwerty123",
		FIO:      "Test User",
	})

	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("Ожидалась ошибка ErrEmailTaken, получена: %v", err)
	}
}

func TestAuthUseCase_Register_Success(t *testing.T) {
	mockUserUseCase := &MockUserUseCase{
		CreateUserFunc: func(req *entity.CreateUserRequest) (*entity.User, error) {
			return &entity.User{
				ID:    2,
				Email: req.Email,
				FIO:   req.FIO,
				Role:  enum.RoleClient,
			}, nil
		},
	}

	authUseCase := NewAuthUseCase(mockUserUseCase)
	token, err := authUseCase.Register(&entity.CreateUserRequest{
		Email:    "newuser@example.com",
		Password: "qwerty12345",
		FIO:      "new user",
		Phone:    "+79991234567",
	})

	if err != nil {
		t.Fatalf("Ожидался успех, но получена ошибка: %v", err)
	}
	if token == "" {
		t.Error("Токен не был сгенерирован при регистрации")
	}
}

func TestAuthUseCase_GetUserByID_UserNotFound(t *testing.T) {
	mockUserUseCase := &MockUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	authUseCase := NewAuthUseCase(mockUserUseCase)
	_, err := authUseCase.GetUserByID(1)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_GetUserByID_Success(t *testing.T) {
	mockUserUseCase := &MockUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	authUseCase := NewAuthUseCase(mockUserUseCase)
	_, err := authUseCase.GetUserByID(1)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestAuthUseCase_UpdateUser_UserNotFound(t *testing.T) {
	mockUserUseCase := &MockUserUseCase{
		UpdateUserFunc: func(id int, req *entity.UpdateUserRequest) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	authUseCase := NewAuthUseCase(mockUserUseCase)
	_, err := authUseCase.UpdateUser(1, &entity.UpdateUserRequest{
		Phone: new("+79991234567"),
	})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_UpdateUser_Success(t *testing.T) {
	mockUserUseCase := &MockUserUseCase{
		UpdateUserFunc: func(id int, req *entity.UpdateUserRequest) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	authUseCase := NewAuthUseCase(mockUserUseCase)
	_, err := authUseCase.UpdateUser(1, &entity.UpdateUserRequest{
		Phone: new("+79991234567"),
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_UserNotFound(t *testing.T) {
	mockUserUseCase := &MockUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	authUseCase := NewAuthUseCase(mockUserUseCase)
	err := authUseCase.ChangePassword(1, &entity.ChangePasswordRequest{
		CurrentPassword: "old",
		NewPassword:     "new",
	})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_InvalidCurrentPassword(t *testing.T) {
	mockUserUseCase := &MockUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	authUseCase := NewAuthUseCase(mockUserUseCase)
	err := authUseCase.ChangePassword(1, &entity.ChangePasswordRequest{
		CurrentPassword: "wrong",
		NewPassword:     "new",
	})

	if !errors.Is(err, domain.ErrCurPasswordIsNotCorrect) {
		t.Errorf("Ожидалась ошибка ErrCurPasswordIsNotCorrect, получена: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_SamePassword(t *testing.T) {
	mockUserUseCase := &MockUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	authUseCase := NewAuthUseCase(mockUserUseCase)
	err := authUseCase.ChangePassword(1, &entity.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "password123",
	})

	if !errors.Is(err, domain.ErrSamePassword) {
		t.Errorf("Ожидалась ошибка ErrSamePassword, получена: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_Success(t *testing.T) {
	mockUserUseCase := &MockUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
		ChangePasswordFunc: func(id int, newPassword string) error {
			return nil
		},
	}

	authUseCase := NewAuthUseCase(mockUserUseCase)
	err := authUseCase.ChangePassword(1, &entity.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "new",
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}
