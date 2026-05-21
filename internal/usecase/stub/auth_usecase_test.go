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

type StubUserUseCase struct {
	GetUserByEmailFunc func(email string) (*entity.User, error)
	CreateUserFunc     func(req *entity.CreateUserRequest) (*entity.User, error)
	GetMeFunc          func(id int) (*entity.User, error)
	UpdateMeFunc       func(id int, req *entity.UpdateUserRequest) (*entity.User, error)
	ChangePasswordFunc func(id int, newPassword string) error
	ChangeEmailFunc    func(id int, newEmail string) error
}

type StubOutboxAuthRepository struct {
	CreateEventFunc func(event *entity.OutboxEvent) error
}

func (m *StubUserUseCase) GetUserByEmail(email string) (*entity.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(email)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserUseCase) CreateUser(req *entity.CreateUserRequest) (*entity.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(req)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserUseCase) GetUserByID(id int) (*entity.User, error) {
	if m.GetMeFunc != nil {
		return m.GetMeFunc(id)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserUseCase) UpdateUser(id int, req *entity.UpdateUserRequest) (*entity.User, error) {
	if m.UpdateMeFunc != nil {
		return m.UpdateMeFunc(id, req)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserUseCase) ChangePassword(id int, newPassword string) error {
	if m.ChangePasswordFunc != nil {
		return m.ChangePasswordFunc(id, newPassword)
	}
	return domain.ErrNotImplemented
}

func (m *StubUserUseCase) ChangeEmail(id int, newEmail string) error {
	if m.ChangeEmailFunc != nil {
		return m.ChangeEmailFunc(id, newEmail)
	}
	return domain.ErrNotImplemented
}

func (m *StubOutboxAuthRepository) CreateEvent(event *entity.OutboxEvent) error {
	if m.CreateEventFunc != nil {
		return m.CreateEventFunc(event)
	}
	return nil
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
	UserNotificationSettings: entity.NotificationSettings{
		IsEmailSend: true,
		IsPhoneSend: false,
	},
}

func TestAuthUseCase_Login_InvalidEmail(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetUserByEmailFunc: func(email string) (*entity.User, error) {
			return nil, domain.ErrInvalidEmail
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
	_, err := authUseCase.Login("notfound@example.com", "password123")
	if !errors.Is(err, domain.ErrInvalidEmail) {
		t.Errorf("Ожидалась ошибка ErrInvalidEmail, получена: %v", err)
	}
}

func TestAuthUseCase_Login_InvalidPassword(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetUserByEmailFunc: func(email string) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
	_, err := authUseCase.Login("test@example.com", "wrong_password")
	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Errorf("Ожидалась ошибка ErrInvalidPassword, получена: %v", err)
	}
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetUserByEmailFunc: func(email string) (*entity.User, error) {
			if email == "test@example.com" {
				return testAuthUser, nil
			}
			return nil, domain.ErrInvalidEmail
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
	token, err := authUseCase.Login("test@example.com", "password123")

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if token == "" {
		t.Error("Токен не был сгенерирован")
	}
}

func TestAuthUseCase_Register_EmailTaken(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		CreateUserFunc: func(req *entity.CreateUserRequest) (*entity.User, error) {
			return nil, domain.ErrEmailTaken
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
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
	stubUserUseCase := &StubUserUseCase{
		CreateUserFunc: func(req *entity.CreateUserRequest) (*entity.User, error) {
			return &entity.User{
				ID:    2,
				Email: req.Email,
				FIO:   req.FIO,
				Role:  enum.RoleClient,
				UserNotificationSettings: entity.NotificationSettings{
					IsEmailSend: true,
					IsPhoneSend: false,
				},
			}, nil
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
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
	stubUserUseCase := &StubUserUseCase{
		GetMeFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
	_, err := authUseCase.GetMe(1)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_GetUserByID_Success(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetMeFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
	_, err := authUseCase.GetMe(1)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestAuthUseCase_UpdateUser_UserNotFound(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		UpdateMeFunc: func(id int, req *entity.UpdateUserRequest) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
	_, err := authUseCase.UpdateMe(1, &entity.UpdateUserRequest{
		Phone: new("+79991234567"),
	})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_UpdateUser_Success(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		UpdateMeFunc: func(id int, req *entity.UpdateUserRequest) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
	_, err := authUseCase.UpdateMe(1, &entity.UpdateUserRequest{
		Phone: new("+79991234567"),
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_UserNotFound(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetMeFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
	err := authUseCase.ChangePassword(1, &entity.ChangePasswordRequest{
		CurrentPassword: "old",
		NewPassword:     "new",
	})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_InvalidCurrentPassword(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetMeFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
	err := authUseCase.ChangePassword(1, &entity.ChangePasswordRequest{
		CurrentPassword: "wrong",
		NewPassword:     "new",
	})

	if !errors.Is(err, domain.ErrCurPasswordIsNotCorrect) {
		t.Errorf("Ожидалась ошибка ErrCurPasswordIsNotCorrect, получена: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_SamePassword(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetMeFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
	err := authUseCase.ChangePassword(1, &entity.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "password123",
	})

	if !errors.Is(err, domain.ErrSamePassword) {
		t.Errorf("Ожидалась ошибка ErrSamePassword, получена: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_Success(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetMeFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
		ChangePasswordFunc: func(id int, newPassword string) error {
			return nil
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
	err := authUseCase.ChangePassword(1, &entity.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "new",
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_UserNotFound(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetMeFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
	err := authUseCase.ChangeEmail(1, &entity.ChangeEmailRequest{
		ConfirmPassword: "old",
		NewEmail:        "newemail@example.com",
	})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_InvalidCurrentPassword(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetMeFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
	err := authUseCase.ChangeEmail(1, &entity.ChangeEmailRequest{
		ConfirmPassword: "wrong",
		NewEmail:        "newemail@example.com",
	})

	if !errors.Is(err, domain.ErrCurPasswordIsNotCorrect) {
		t.Errorf("Ожидалась ошибка ErrCurPasswordIsNotCorrect, получена: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_Success(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetMeFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
		ChangeEmailFunc: func(id int, newEmail string) error {
			return nil
		},
	}

	stubNotifier := &StubOutboxAuthRepository{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier)
	err := authUseCase.ChangeEmail(1, &entity.ChangeEmailRequest{
		ConfirmPassword: "password123",
		NewEmail:        "newemail@example.com",
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}
