package stubs

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/internal/usecase"
	"context"
	"errors"
	"testing"
	"time"
)

type StubUserUseCase struct {
	GetUserByEmailFunc        func(email string) (*entity.User, error)
	CreateUserFunc            func(req *entity.CreateUserRequest) (*entity.User, error)
	GetUserByIDFunc           func(id int) (*entity.User, error)
	UpdateUserProfileFunc     func(id int, req *entity.UpdateUserProfileRequest) (*entity.User, error)
	UpdateUserDataFunc        func(id int, updates map[string]any) (*entity.User, error)
	EmailExistsFunc           func(email string) (bool, error)
	GetUserByEmailContextFunc func(ctx context.Context, email string) (*entity.User, error)
}

type StubOutboxRepository struct {
	CreateEventFunc        func(event *entity.OutboxEvent) error
	CreateEventContextFunc func(ctx context.Context, event *entity.OutboxEvent) error
}

type StubTotpService struct {
	GenerateSecretFunc func(email string) (string, string, error)
	ValidateCodeFunc   func(secret, code string) bool
}

func (m *StubUserUseCase) GetUserByEmail(email string) (*entity.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(email)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserUseCase) GetUserByEmailContext(ctx context.Context, email string) (*entity.User, error) {
	if m.GetUserByEmailContextFunc != nil {
		return m.GetUserByEmailContextFunc(ctx, email)
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
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(id)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserUseCase) UpdateUserProfile(id int, req *entity.UpdateUserProfileRequest) (*entity.User, error) {
	if m.UpdateUserProfileFunc != nil {
		return m.UpdateUserProfileFunc(id, req)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserUseCase) UpdateUserData(id int, updates map[string]any) (*entity.User, error) {
	if m.UpdateUserDataFunc != nil {
		return m.UpdateUserDataFunc(id, updates)
	}
	return nil, domain.ErrNotImplemented
}

func (m *StubUserUseCase) EmailExists(email string) (bool, error) {
	if m.EmailExistsFunc != nil {
		return m.EmailExistsFunc(email)
	}

	return false, domain.ErrNotImplemented
}

func (m *StubOutboxRepository) CreateEvent(event *entity.OutboxEvent) error {
	if m.CreateEventFunc != nil {
		return m.CreateEventFunc(event)
	}
	return nil
}

func (m *StubOutboxRepository) CreateEventContext(ctx context.Context, event *entity.OutboxEvent) error {
	if m.CreateEventContextFunc != nil {
		return m.CreateEventContextFunc(ctx, event)
	}
	return nil
}

func (s StubTotpService) GenerateSecret(email string) (string, string, error) {
	if s.GenerateSecretFunc != nil {
		return s.GenerateSecretFunc(email)
	}
	return "", "", domain.ErrNotImplemented
}

func (s StubTotpService) ValidateCode(secret, code string) bool {
	if s.ValidateCodeFunc != nil {
		return s.ValidateCodeFunc(secret, code)
	}
	return false
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

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
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

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
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

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
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

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
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

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
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
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
	_, err := authUseCase.GetMe(1)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_GetUserByID_Success(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
	_, err := authUseCase.GetMe(1)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestAuthUseCase_UpdateUser_UserNotFound(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		UpdateUserProfileFunc: func(id int, req *entity.UpdateUserProfileRequest) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
	_, err := authUseCase.UpdateMe(1, &entity.UpdateUserProfileRequest{
		Phone: new("+79991234567"),
	})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_UpdateUser_Success(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		UpdateUserProfileFunc: func(id int, req *entity.UpdateUserProfileRequest) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
	_, err := authUseCase.UpdateMe(1, &entity.UpdateUserProfileRequest{
		Phone: new("+79991234567"),
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_UserNotFound(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
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
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
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
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
	err := authUseCase.ChangePassword(1, &entity.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "password123",
	})

	if !errors.Is(err, domain.ErrSamePassword) {
		t.Errorf("Ожидалась ошибка ErrSamePassword, получена: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_TotpRequired(t *testing.T) {
	userWithTotp := &entity.User{
		ID:            1,
		IsTotpEnabled: true,
		TotpSecret:    "SECRET123",
		PasswordHash:  testAuthUser.PasswordHash,
	}
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	err := authUseCase.ChangePassword(1, &entity.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newpass",
		OtpCode:         "",
	})

	if !errors.Is(err, domain.ErrTotpSecretNotSet) {
		t.Errorf("Ожидалась ошибка ErrTotpSecretNotSet (требуется код), получена: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_TotpInvalidCode(t *testing.T) {
	userWithTotp := &entity.User{
		ID:            1,
		IsTotpEnabled: true,
		TotpSecret:    "SECRET123",
		PasswordHash:  testAuthUser.PasswordHash,
	}
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{
		ValidateCodeFunc: func(secret, code string) bool {
			return false
		},
	}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	err := authUseCase.ChangePassword(1, &entity.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newpass",
		OtpCode:         "wrongcode",
	})

	if !errors.Is(err, domain.ErrInvalidTotpCode) {
		t.Errorf("Ожидалась ошибка ErrInvalidTotpCode, получена: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_WithTotp_Success(t *testing.T) {
	userWithTotp := &entity.User{
		ID:            1,
		IsTotpEnabled: true,
		TotpSecret:    "SECRET123",
		PasswordHash:  testAuthUser.PasswordHash,
	}
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return userWithTotp, nil
		},
		UpdateUserDataFunc: func(id int, updates map[string]any) (*entity.User, error) {
			if updates["password_hash"] == "" {
				t.Error("password_hash не был обновлён")
			}
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxRepository{
		CreateEventFunc: func(event *entity.OutboxEvent) error {
			if event.EventType != string(enum.ChangePasswordType) {
				t.Errorf("Неверный тип события: %s", event.EventType)
			}
			return nil
		},
	}
	stubTotp := &StubTotpService{
		ValidateCodeFunc: func(secret, code string) bool {
			return secret == "SECRET123" && code == "123456"
		},
	}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	err := authUseCase.ChangePassword(1, &entity.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newSecurePass!",
		OtpCode:         "123456",
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_Success(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
		UpdateUserDataFunc: func(id int, updates map[string]any) (*entity.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
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
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
	_, err := authUseCase.ChangeEmail(1, &entity.ChangeEmailRequest{
		ConfirmPassword: "old",
		NewEmail:        "newemail@example.com",
		OtpCode:         "111111",
	})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_InvalidCurrentPassword(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
		EmailExistsFunc: func(email string) (bool, error) {
			return false, nil
		},
	}

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
	_, err := authUseCase.ChangeEmail(1, &entity.ChangeEmailRequest{
		ConfirmPassword: "wrong",
		NewEmail:        "newemail@example.com",
		OtpCode:         "111111",
	})

	if !errors.Is(err, domain.ErrCurPasswordIsNotCorrect) {
		t.Errorf("Ожидалась ошибка ErrCurPasswordIsNotCorrect, получена: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_EmailExists(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
		EmailExistsFunc: func(email string) (bool, error) {
			return true, nil
		},
	}

	stubNotifier := &StubOutboxRepository{}
	stubTotpService := &StubTotpService{}
	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubNotifier, stubTotpService)
	_, err := authUseCase.ChangeEmail(1, &entity.ChangeEmailRequest{
		ConfirmPassword: "password123",
		NewEmail:        "newemail@example.com",
	})

	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("Ожидалась ошибка ErrEmailTaken, получена: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_TotpRequired(t *testing.T) {
	userWithTotp := &entity.User{
		ID:            1,
		IsTotpEnabled: true,
		TotpSecret:    "SECRET123",
		PasswordHash:  testAuthUser.PasswordHash,
	}
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return userWithTotp, nil
		},
		EmailExistsFunc: func(email string) (bool, error) {
			return false, nil
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	_, err := authUseCase.ChangeEmail(1, &entity.ChangeEmailRequest{
		NewEmail:        "new@example.com",
		ConfirmPassword: "password123",
		OtpCode:         "",
	})

	if !errors.Is(err, domain.ErrTotpSecretNotSet) {
		t.Errorf("Ожидалась ошибка ErrTotpSecretNotSet (требуется код), получена: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_WithTotp_Success(t *testing.T) {
	userWithTotp := &entity.User{
		ID:            1,
		Email:         "old@example.com",
		IsTotpEnabled: true,
		TotpSecret:    "SECRET123",
		PasswordHash:  testAuthUser.PasswordHash,
	}

	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return userWithTotp, nil
		},
		EmailExistsFunc: func(email string) (bool, error) {
			return false, nil
		},
		UpdateUserDataFunc: func(id int, updates map[string]any) (*entity.User, error) {
			if updates["email"] != "new@example.com" {
				t.Errorf("Неверный email в updates: %v", updates["email"])
			}

			updated := *userWithTotp
			updated.Email = "new@example.com"
			return &updated, nil
		},
	}

	stubOutbox := &StubOutboxRepository{
		CreateEventFunc: func(event *entity.OutboxEvent) error {
			if event.EventType != string(enum.ChangeEmailType) {
				t.Errorf("Неверный тип события: %s", event.EventType)
			}
			return nil
		},
	}

	stubTotp := &StubTotpService{
		ValidateCodeFunc: func(secret, code string) bool {
			return true
		},
	}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	updatedUser, err := authUseCase.ChangeEmail(1, &entity.ChangeEmailRequest{
		NewEmail:        "new@example.com",
		ConfirmPassword: "password123",
		OtpCode:         "123456",
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if updatedUser.Email != "new@example.com" {
		t.Errorf("Email не обновился: ожидался 'new@example.com', получен '%s'", updatedUser.Email)
	}
}

func TestAuthUseCase_SetupTotp_UserNotFound(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	_, err := authUseCase.SetupTotp(999)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_SetupTotp_AlreadyEnabled(t *testing.T) {
	userWithTotp := &entity.User{
		ID:            1,
		Email:         "test@example.com",
		IsTotpEnabled: true,
	}
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	_, err := authUseCase.SetupTotp(1)

	if !errors.Is(err, domain.ErrTotpAlreadyEnabled) {
		t.Errorf("Ожидалась ошибка ErrTotpAlreadyEnabled, получена: %v", err)
	}
}

func TestAuthUseCase_SetupTotp_Success(t *testing.T) {
	expectedSecret := "JBSWY3DPEHPK3PXP"
	expectedURL := "otpauth://totp/BookingGo:test@example.com?secret=JBSWY3DPEHPK3PXP&issuer=BookingGo"

	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return testAuthUser, nil
		},
		UpdateUserDataFunc: func(id int, updates map[string]any) (*entity.User, error) {
			if updates["totp_secret"] != expectedSecret {
				t.Errorf("Неверный секрет в updates: %v", updates["totp_secret"])
			}
			return testAuthUser, nil
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{
		GenerateSecretFunc: func(email string) (string, string, error) {
			if email != "test@example.com" {
				t.Errorf("Неверный email в GenerateSecret: %s", email)
			}
			return expectedSecret, expectedURL, nil
		},
	}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	otpURL, err := authUseCase.SetupTotp(1)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if otpURL != expectedURL {
		t.Errorf("Неверный otpauth URL: ожидался %s, получен %s", expectedURL, otpURL)
	}
}

func TestAuthUseCase_VerifyTotp_UserNotFound(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	err := authUseCase.VerifyTotp(999, &entity.VerifyTotpRequest{OtpCode: "123456"})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_VerifyTotp_AlreadyEnabled(t *testing.T) {
	userWithTotp := &entity.User{
		ID:            1,
		IsTotpEnabled: true,
		TotpSecret:    "SECRET123",
	}
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	err := authUseCase.VerifyTotp(1, &entity.VerifyTotpRequest{OtpCode: "123456"})

	if !errors.Is(err, domain.ErrTotpAlreadyEnabled) {
		t.Errorf("Ожидалась ошибка ErrTotpAlreadyEnabled, получена: %v", err)
	}
}

func TestAuthUseCase_VerifyTotp_SecretNotSet(t *testing.T) {
	userWithoutSecret := &entity.User{
		ID:            1,
		IsTotpEnabled: false,
		TotpSecret:    "",
	}
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return userWithoutSecret, nil
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	err := authUseCase.VerifyTotp(1, &entity.VerifyTotpRequest{OtpCode: "123456"})

	if !errors.Is(err, domain.ErrTotpSecretNotSet) {
		t.Errorf("Ожидалась ошибка ErrTotpSecretNotSet, получена: %v", err)
	}
}

func TestAuthUseCase_VerifyTotp_InvalidCode(t *testing.T) {
	userWithSecret := &entity.User{
		ID:            1,
		IsTotpEnabled: false,
		TotpSecret:    "SECRET123",
	}
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return userWithSecret, nil
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{
		ValidateCodeFunc: func(secret, code string) bool {
			return false
		},
	}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	err := authUseCase.VerifyTotp(1, &entity.VerifyTotpRequest{OtpCode: "wrongcode"})

	if !errors.Is(err, domain.ErrInvalidTotpCode) {
		t.Errorf("Ожидалась ошибка ErrInvalidTotpCode, получена: %v", err)
	}
}

func TestAuthUseCase_VerifyTotp_Success(t *testing.T) {
	userWithSecret := &entity.User{
		ID:            1,
		IsTotpEnabled: false,
		TotpSecret:    "SECRET123",
	}
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return userWithSecret, nil
		},
		UpdateUserDataFunc: func(id int, updates map[string]any) (*entity.User, error) {
			if updates["is_totp_enabled"] != true {
				t.Errorf("Ожидалось is_totp_enabled=true, получено: %v", updates["is_totp_enabled"])
			}
			return userWithSecret, nil
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{
		ValidateCodeFunc: func(secret, code string) bool {
			return secret == "SECRET123" && code == "123456"
		},
	}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	err := authUseCase.VerifyTotp(1, &entity.VerifyTotpRequest{OtpCode: "123456"})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_UserNotFound(t *testing.T) {
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	err := authUseCase.DisableTotp(999, &entity.DisableTotpRequest{
		CurrentPassword: "pass",
		OtpCode:         "123456",
	})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_NotEnabled(t *testing.T) {
	userWithoutTotp := &entity.User{
		ID:            1,
		IsTotpEnabled: false,
	}
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return userWithoutTotp, nil
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	err := authUseCase.DisableTotp(1, &entity.DisableTotpRequest{
		CurrentPassword: "pass",
		OtpCode:         "123456",
	})

	if !errors.Is(err, domain.ErrTotpAlreadyDisabled) {
		t.Errorf("Ожидалась ошибка ErrTotpAlreadyDisabled, получена: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_InvalidOtpCode(t *testing.T) {
	userWithTotp := &entity.User{
		ID:            1,
		IsTotpEnabled: true,
		TotpSecret:    "SECRET123",
		PasswordHash:  testAuthUser.PasswordHash, // хэш "password123"
	}
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{
		ValidateCodeFunc: func(secret, code string) bool {
			return false
		},
	}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	err := authUseCase.DisableTotp(1, &entity.DisableTotpRequest{
		CurrentPassword: "password123",
		OtpCode:         "wrongcode",
	})

	if !errors.Is(err, domain.ErrInvalidTotpCode) {
		t.Errorf("Ожидалась ошибка ErrInvalidTotpCode, получена: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_WrongPassword(t *testing.T) {
	userWithTotp := &entity.User{
		ID:            1,
		IsTotpEnabled: true,
		TotpSecret:    "SECRET123",
		PasswordHash:  testAuthUser.PasswordHash,
	}
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{
		ValidateCodeFunc: func(secret, code string) bool {
			return true
		},
	}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	err := authUseCase.DisableTotp(1, &entity.DisableTotpRequest{
		CurrentPassword: "wrongpass",
		OtpCode:         "123456",
	})

	if !errors.Is(err, domain.ErrCurPasswordIsNotCorrect) {
		t.Errorf("Ожидалась ошибка ErrCurPasswordIsNotCorrect, получена: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_Success(t *testing.T) {
	userWithTotp := &entity.User{
		ID:            1,
		IsTotpEnabled: true,
		TotpSecret:    "SECRET123",
		PasswordHash:  testAuthUser.PasswordHash,
	}
	stubUserUseCase := &StubUserUseCase{
		GetUserByIDFunc: func(id int) (*entity.User, error) {
			return userWithTotp, nil
		},
		UpdateUserDataFunc: func(id int, updates map[string]any) (*entity.User, error) {
			if updates["is_totp_enabled"] != false {
				t.Errorf("Ожидалось is_totp_enabled=false, получено: %v", updates["is_totp_enabled"])
			}
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxRepository{}
	stubTotp := &StubTotpService{
		ValidateCodeFunc: func(secret, code string) bool {
			return true
		},
	}

	authUseCase := usecase.NewAuthUseCase(stubUserUseCase, stubOutbox, stubTotp)
	err := authUseCase.DisableTotp(1, &entity.DisableTotpRequest{
		CurrentPassword: "password123",
		OtpCode:         "123456",
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}
