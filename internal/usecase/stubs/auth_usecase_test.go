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

	"github.com/redis/go-redis/v9"
)

type StubOutboxPostgres struct {
	CreateEventFunc        func(event *domain.OutboxEvent) error
	CreateEventContextFunc func(ctx context.Context, event *domain.OutboxEvent) error
}

type StubTotpService struct {
	GenerateSecretFunc func(email string) (string, string, error)
	ValidateCodeFunc   func(secret, code string) bool
}

type StubCacheService struct {
	GetFunc              func(ctx context.Context, key string, dest any) error
	SetFunc              func(ctx context.Context, key string, value any, ttl time.Duration) error
	DeleteFunc           func(ctx context.Context, key string) error
	DeleteByPrefixFunc   func(ctx context.Context, prefix string) error
	IncrementWithTTLFunc func(ctx context.Context, key string, ttl time.Duration) (int64, error)
	GetClientFunc        func() *redis.Client
	CloseFunc            func() error
}

type StubBlacklistService struct {
	AddToBlacklistFunc        func(ctx context.Context, jti string, expiresAt time.Time) error
	IsInBlacklistFunc         func(ctx context.Context, jti string) bool
	InvalidateAllSessionsFunc func(ctx context.Context, userID int) error
	IsSessionValidFunc        func(ctx context.Context, userID int, issuedAt time.Time) bool
}

func (m *StubOutboxPostgres) CreateEvent(event *domain.OutboxEvent) error {
	if m.CreateEventFunc != nil {
		return m.CreateEventFunc(event)
	}
	return nil
}

func (m *StubOutboxPostgres) CreateEventContext(ctx context.Context, event *domain.OutboxEvent) error {
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

func (sc StubCacheService) Get(ctx context.Context, key string, dest any) error {
	if sc.GetFunc != nil {
		return sc.GetFunc(ctx, key, dest)
	}
	return domain.ErrCacheKeyNotFound
}

func (sc StubCacheService) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if sc.SetFunc != nil {
		return sc.SetFunc(ctx, key, value, ttl)
	}
	return nil
}

func (sc StubCacheService) Delete(ctx context.Context, key string) error {
	if sc.DeleteFunc != nil {
		return sc.DeleteFunc(ctx, key)
	}
	return domain.ErrNotImplemented
}

func (sc StubCacheService) DeleteByPrefix(ctx context.Context, prefix string) error {
	if sc.DeleteByPrefixFunc != nil {
		return sc.DeleteByPrefixFunc(ctx, prefix)
	}
	return domain.ErrNotImplemented
}

func (sc StubCacheService) IncrementWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if sc.IncrementWithTTLFunc != nil {
		return sc.IncrementWithTTLFunc(ctx, key, ttl)
	}
	return 0, domain.ErrNotImplemented
}

func (sc StubCacheService) GetClient() *redis.Client {
	if sc.GetClientFunc != nil {
		return sc.GetClientFunc()
	}
	return nil
}

func (sc StubCacheService) Close() error {
	if sc.CloseFunc != nil {
		return sc.CloseFunc()
	}
	return nil
}

func (s *StubBlacklistService) AddToBlacklist(ctx context.Context, jti string, expiresAt time.Time) error {
	if s.AddToBlacklistFunc != nil {
		return s.AddToBlacklistFunc(ctx, jti, expiresAt)
	}
	return domain.ErrNotImplemented
}

func (s *StubBlacklistService) IsInBlacklist(ctx context.Context, jti string) bool {
	if s.IsInBlacklistFunc != nil {
		return s.IsInBlacklistFunc(ctx, jti)
	}
	return false
}

func (s *StubBlacklistService) InvalidateAllSessions(ctx context.Context, userID int) error {
	if s.InvalidateAllSessionsFunc != nil {
		return s.InvalidateAllSessionsFunc(ctx, userID)
	}
	return domain.ErrNotImplemented
}

func (s *StubBlacklistService) IsSessionValid(ctx context.Context, userID int, issuedAt time.Time) bool {
	if s.IsSessionValidFunc != nil {
		return s.IsSessionValidFunc(ctx, userID, issuedAt)
	}
	return true
}

var testAuthUser = &domain.User{
	ID:           1,
	Email:        "test@example.com",
	PasswordHash: "$2a$10$C1AtwB2zZtvq5llPqyEkY.YIDB28fNHrVUrcBN1F6CgEDb.YdU0qa", //password123
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

var userWithTotp = &domain.User{
	ID:            1,
	IsTotpEnabled: true,
	TotpSecret:    "SECRET123",
	PasswordHash:  testAuthUser.PasswordHash,
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	outboxCalled := false

	stubUserPostgres := &StubUserPostgres{
		GetByEmailFunc: func(email string) (*domain.User, error) {
			if email == "test@example.com" {
				return testAuthUser, nil
			}
			return nil, domain.ErrInvalidEmail
		},
	}

	stubNotifier := &StubOutboxPostgres{
		CreateEventFunc: func(event *domain.OutboxEvent) error {
			outboxCalled = true
			if event.EventType != string(enum.AuthType) {
				t.Errorf("Неверный тип события: %s", event.EventType)
			}
			return nil
		},
	}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	loginReq := &dto.LoginUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	token, err := authUseCase.Login(loginReq)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if token == "" {
		t.Error("Токен не был сгенерирован")
	}
	if !outboxCalled {
		t.Error("Outbox событие должно было быть создано")
	}
}

func TestAuthUseCase_Login_InvalidEmail(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByEmailFunc: func(email string) (*domain.User, error) {
			return nil, domain.ErrInvalidEmail
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	loginReq := &dto.LoginUserRequest{
		Email:    "not_found@example.com",
		Password: "password123",
	}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	_, err := authUseCase.Login(loginReq)
	if !errors.Is(err, domain.ErrInvalidEmail) {
		t.Errorf("Ожидалась ошибка ErrInvalidEmail, получена: %v", err)
	}
}

func TestAuthUseCase_Login_InvalidPassword(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByEmailFunc: func(email string) (*domain.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	loginReq := &dto.LoginUserRequest{
		Email:    "test@example.com",
		Password: "wrong_password",
	}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	_, err := authUseCase.Login(loginReq)
	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Errorf("Ожидалась ошибка ErrInvalidPassword, получена: %v", err)
	}
}

func TestAuthUseCase_Register_Success(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		EmailExistsFunc: func(email string) (bool, error) {
			return false, nil
		},
		CreateFunc: func(user *domain.User) error {
			user.ID = 2
			return nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	token, err := authUseCase.Register(&dto.CreateUserRequest{
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

func TestAuthUseCase_Register_EmailTaken(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		EmailExistsFunc: func(email string) (bool, error) {
			return true, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	_, err := authUseCase.Register(&dto.CreateUserRequest{
		Email:    "taken@example.com",
		Password: "qwerty123",
		FIO:      "Test User",
	})

	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("Ожидалась ошибка ErrEmailTaken, получена: %v", err)
	}
}

func TestAuthUseCase_GetMe_Success(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	_, err := authUseCase.GetMe(1)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestAuthUseCase_GetMe_CacheHit(t *testing.T) {
	dbCalled := false

	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			dbCalled = true
			return testAuthUser, nil
		},
	}

	stubCache := &StubCacheService{
		GetFunc: func(ctx context.Context, key string, dest any) error {
			if userPtr, ok := dest.(*domain.User); ok {
				*userPtr = *testAuthUser
			}
			return nil
		},
	}
	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCache, stubBlacklistService)
	user, err := authUseCase.GetMe(1)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if user.Email != testAuthUser.Email {
		t.Errorf("Ожидался email %s, получен %s", testAuthUser.Email, user.Email)
	}
	if dbCalled {
		t.Error("GetByID не должен вызываться при redis hit")
	}
}

func TestAuthUseCase_GetMe_CacheMiss(t *testing.T) {
	cacheSetCalled := false

	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testAuthUser, nil
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
	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCache, stubBlacklistService)
	user, err := authUseCase.GetMe(1)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if user.Email != testAuthUser.Email {
		t.Errorf("Ожидался email %s, получен %s", testAuthUser.Email, user.Email)
	}
	if !cacheSetCalled {
		t.Error("Cache Set должен вызываться при redis miss")
	}
}

func TestAuthUseCase_GetMe_CacheSetError(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testAuthUser, nil
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
	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCache, stubBlacklistService)
	user, err := authUseCase.GetMe(1)

	if err != nil {
		t.Fatalf("Ошибка кэша не должна ломать флоу, получена: %v", err)
	}
	if user.Email != testAuthUser.Email {
		t.Errorf("Ожидался email %s, получен %s", testAuthUser.Email, user.Email)
	}
}

func TestAuthUseCase_GetMe_UserNotFound_NoCacheSet(t *testing.T) {
	cacheSetCalled := false

	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
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
	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCache, stubBlacklistService)
	_, err := authUseCase.GetMe(999)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
	if cacheSetCalled {
		t.Error("Cache Set не должен вызываться, если пользователь не найден")
	}
}

func TestAuthUseCase_GetMe_UserNotFound(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	_, err := authUseCase.GetMe(1)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_UpdateMe_Success(t *testing.T) {
	cacheUserInvalidated := false

	stubUserPostgres := &StubUserPostgres{
		UpdateFunc: func(id int, updates map[string]any) (*domain.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{
		DeleteFunc: func(ctx context.Context, key string) error {
			cacheUserInvalidated = true
			return nil
		},
	}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	_, err := authUseCase.UpdateMe(1, &dto.UpdateUserProfileRequest{
		Phone: new("+79991234567"),
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if !cacheUserInvalidated {
		t.Error("Ожидалась инвалидация кэша пользователя, но она не была вызвана")
	}
}

func TestAuthUseCase_UpdateMe_UserNotFound(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		UpdateFunc: func(id int, updates map[string]any) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	_, err := authUseCase.UpdateMe(1, &dto.UpdateUserProfileRequest{
		Phone: new("+79991234567"),
	})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_Success(t *testing.T) {
	passwordUpdated := false
	outboxCalled := false
	sessionsInvalidated := false

	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testAuthUser, nil
		},
		UpdateFunc: func(id int, updates map[string]any) (*domain.User, error) {
			passwordUpdated = true
			if updates["password_hash"] == "" {
				t.Error("password_hash не был обновлён")
			}
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{
		CreateEventFunc: func(event *domain.OutboxEvent) error {
			outboxCalled = true
			if event.EventType != string(enum.ChangePasswordType) {
				t.Errorf("Неверный тип события: %s", event.EventType)
			}
			return nil
		},
	}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{
		InvalidateAllSessionsFunc: func(ctx context.Context, userID int) error {
			sessionsInvalidated = true
			if userID != testAuthUser.ID {
				t.Errorf("Expected userID %d, got %d", testAuthUser.ID, userID)
			}
			return nil
		},
	}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	err := authUseCase.ChangePassword(1, &dto.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "new",
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if !outboxCalled {
		t.Error("Outbox событие должно было быть создано")
	}
	if !passwordUpdated {
		t.Error("Пароль должен быть обновлен")
	}
	if !sessionsInvalidated {
		t.Error("Все сессии должны быть деактивированы")
	}
}

func TestAuthUseCase_ChangePassword_WithTotp_Success(t *testing.T) {
	outboxCalled := false
	sessionsInvalidated := false

	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return userWithTotp, nil
		},
		UpdateFunc: func(id int, updates map[string]any) (*domain.User, error) {
			if updates["password_hash"] == "" {
				t.Error("password_hash не был обновлён")
			}
			return userWithTotp, nil
		},
	}
	stubNotifier := &StubOutboxPostgres{
		CreateEventFunc: func(event *domain.OutboxEvent) error {
			outboxCalled = true
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
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{
		InvalidateAllSessionsFunc: func(ctx context.Context, userID int) error {
			sessionsInvalidated = true
			return nil
		},
	}
	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotp, stubCacheService, stubBlacklistService)
	err := authUseCase.ChangePassword(1, &dto.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newSecurePass!",
		OtpCode:         "123456",
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if !outboxCalled {
		t.Error("Outbox событие должно было быть создано")
	}
	if !sessionsInvalidated {
		t.Error("Все сессии должны быть деактивированы")
	}
}

func TestAuthUseCase_ChangePassword_UserNotFound(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	err := authUseCase.ChangePassword(1, &dto.ChangePasswordRequest{
		CurrentPassword: "old",
		NewPassword:     "new",
	})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_InvalidCurrentPassword(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	err := authUseCase.ChangePassword(1, &dto.ChangePasswordRequest{
		CurrentPassword: "wrong",
		NewPassword:     "new",
	})

	if !errors.Is(err, domain.ErrCurPasswordIsNotCorrect) {
		t.Errorf("Ожидалась ошибка ErrCurPasswordIsNotCorrect, получена: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_SamePassword(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testAuthUser, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	err := authUseCase.ChangePassword(1, &dto.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "password123",
	})

	if !errors.Is(err, domain.ErrSamePassword) {
		t.Errorf("Ожидалась ошибка ErrSamePassword, получена: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_TotpRequired(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	err := authUseCase.ChangePassword(1, &dto.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newpass",
		OtpCode:         "",
	})

	if !errors.Is(err, domain.ErrTotpSecretNotSet) {
		t.Errorf("Ожидалась ошибка ErrTotpSecretNotSet (требуется код), получена: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_TotpInvalidCode(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{
		ValidateCodeFunc: func(secret, code string) bool {
			return false
		},
	}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	err := authUseCase.ChangePassword(1, &dto.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newpass",
		OtpCode:         "wrongcode",
	})

	if !errors.Is(err, domain.ErrInvalidTotpCode) {
		t.Errorf("Ожидалась ошибка ErrInvalidTotpCode, получена: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_Success(t *testing.T) {
	cacheUserInvalidated := false
	outboxCalled := false
	sessionsInvalidated := false

	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testAuthUser, nil
		},
		EmailExistsFunc: func(email string) (bool, error) {
			return false, nil
		},
		UpdateFunc: func(id int, updates map[string]any) (*domain.User, error) {
			if updates["email"] != "new@example.com" {
				t.Errorf("Неверный email в updates: %v", updates["email"])
			}

			updated := *testAuthUser
			updated.Email = "new@example.com"
			return &updated, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{
		CreateEventFunc: func(event *domain.OutboxEvent) error {
			outboxCalled = true
			if event.EventType != string(enum.ChangeEmailType) {
				t.Errorf("Неверный тип события: %s", event.EventType)
			}
			return nil
		},
	}

	stubTotp := &StubTotpService{}
	stubCacheService := &StubCacheService{
		DeleteFunc: func(ctx context.Context, key string) error {
			cacheUserInvalidated = true
			return nil
		},
	}
	stubBlacklistService := &StubBlacklistService{
		InvalidateAllSessionsFunc: func(ctx context.Context, userID int) error {
			sessionsInvalidated = true
			if userID != testAuthUser.ID {
				t.Errorf("Expected userID %d, got %d", testAuthUser.ID, userID)
			}
			return nil
		},
	}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotp, stubCacheService, stubBlacklistService)
	updatedUser, err := authUseCase.ChangeEmail(1, &dto.ChangeEmailRequest{
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
	if !cacheUserInvalidated {
		t.Error("Ожидалась инвалидация кэша пользователя, но она не была вызвана")
	}
	if !outboxCalled {
		t.Error("Outbox событие должно было быть создано")
	}
	if !sessionsInvalidated {
		t.Error("Все сессии должны быть деактивированы")
	}
}

func TestAuthUseCase_ChangeEmail_WithTotp_Success(t *testing.T) {
	sessionsInvalidated := false

	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return userWithTotp, nil
		},
		EmailExistsFunc: func(email string) (bool, error) {
			return false, nil
		},
		UpdateFunc: func(id int, updates map[string]any) (*domain.User, error) {
			if updates["email"] != "new@example.com" {
				t.Errorf("Неверный email в updates: %v", updates["email"])
			}

			updated := *userWithTotp
			updated.Email = "new@example.com"
			return &updated, nil
		},
	}

	stubOutbox := &StubOutboxPostgres{
		CreateEventFunc: func(event *domain.OutboxEvent) error {
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
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{
		InvalidateAllSessionsFunc: func(ctx context.Context, userID int) error {
			sessionsInvalidated = true
			if userID != testAuthUser.ID {
				t.Errorf("Expected userID %d, got %d", testAuthUser.ID, userID)
			}
			return nil
		},
	}
	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	updatedUser, err := authUseCase.ChangeEmail(1, &dto.ChangeEmailRequest{
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
	if !sessionsInvalidated {
		t.Error("Все сессии должны быть деактивированы")
	}
}

func TestAuthUseCase_ChangeEmail_UserNotFound(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	_, err := authUseCase.ChangeEmail(1, &dto.ChangeEmailRequest{
		ConfirmPassword: "old",
		NewEmail:        "newemail@example.com",
		OtpCode:         "111111",
	})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_InvalidCurrentPassword(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testAuthUser, nil
		},
		EmailExistsFunc: func(email string) (bool, error) {
			return false, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	_, err := authUseCase.ChangeEmail(1, &dto.ChangeEmailRequest{
		ConfirmPassword: "wrong",
		NewEmail:        "newemail@example.com",
		OtpCode:         "111111",
	})

	if !errors.Is(err, domain.ErrCurPasswordIsNotCorrect) {
		t.Errorf("Ожидалась ошибка ErrCurPasswordIsNotCorrect, получена: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_EmailExists(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testAuthUser, nil
		},
		EmailExistsFunc: func(email string) (bool, error) {
			return true, nil
		},
	}

	stubNotifier := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubNotifier, stubTotpService, stubCacheService, stubBlacklistService)
	_, err := authUseCase.ChangeEmail(1, &dto.ChangeEmailRequest{
		ConfirmPassword: "password123",
		NewEmail:        "newemail@example.com",
	})

	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("Ожидалась ошибка ErrEmailTaken, получена: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_TotpRequired(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return userWithTotp, nil
		},
		EmailExistsFunc: func(email string) (bool, error) {
			return false, nil
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	_, err := authUseCase.ChangeEmail(1, &dto.ChangeEmailRequest{
		NewEmail:        "new@example.com",
		ConfirmPassword: "password123",
		OtpCode:         "",
	})

	if !errors.Is(err, domain.ErrTotpSecretNotSet) {
		t.Errorf("Ожидалась ошибка ErrTotpSecretNotSet (требуется код), получена: %v", err)
	}
}

func TestAuthUseCase_SetupTotp_Success(t *testing.T) {
	expectedSecret := "JBSWY3DPEHPK3PXP"
	expectedURL := "otpauth://totp/BookingGo:test@example.com?secret=JBSWY3DPEHPK3PXP&issuer=BookingGo"

	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return testAuthUser, nil
		},
		UpdateFunc: func(id int, updates map[string]any) (*domain.User, error) {
			if updates["totp_secret"] != expectedSecret {
				t.Errorf("Неверный секрет в updates: %v", updates["totp_secret"])
			}
			return testAuthUser, nil
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{
		GenerateSecretFunc: func(email string) (string, string, error) {
			if email != "test@example.com" {
				t.Errorf("Неверный email в GenerateSecret: %s", email)
			}
			return expectedSecret, expectedURL, nil
		},
	}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	otpURL, err := authUseCase.SetupTotp(1)

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if otpURL != expectedURL {
		t.Errorf("Неверный otpauth URL: ожидался %s, получен %s", expectedURL, otpURL)
	}
}

func TestAuthUseCase_SetupTotp_UserNotFound(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	_, err := authUseCase.SetupTotp(999)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_SetupTotp_AlreadyEnabled(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	_, err := authUseCase.SetupTotp(1)

	if !errors.Is(err, domain.ErrTotpAlreadyEnabled) {
		t.Errorf("Ожидалась ошибка ErrTotpAlreadyEnabled, получена: %v", err)
	}
}

func TestAuthUseCase_VerifyTotp_Success(t *testing.T) {
	cacheUserInvalidated := false

	userWithSecret := &domain.User{
		ID:            1,
		IsTotpEnabled: false,
		TotpSecret:    "SECRET123",
	}
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return userWithSecret, nil
		},
		UpdateFunc: func(id int, updates map[string]any) (*domain.User, error) {
			if updates["is_totp_enabled"] != true {
				t.Errorf("Ожидалось is_totp_enabled=true, получено: %v", updates["is_totp_enabled"])
			}
			return userWithSecret, nil
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{
		ValidateCodeFunc: func(secret, code string) bool {
			return secret == "SECRET123" && code == "123456"
		},
	}
	stubCacheService := &StubCacheService{
		DeleteFunc: func(ctx context.Context, key string) error {
			cacheUserInvalidated = true
			return nil
		},
	}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	err := authUseCase.VerifyTotp(1, &dto.VerifyTotpRequest{OtpCode: "123456"})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if !cacheUserInvalidated {
		t.Error("Ожидалась инвалидация кэша пользователя, но она не была вызвана")
	}
}

func TestAuthUseCase_VerifyTotp_UserNotFound(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	err := authUseCase.VerifyTotp(999, &dto.VerifyTotpRequest{OtpCode: "123456"})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_VerifyTotp_AlreadyEnabled(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	err := authUseCase.VerifyTotp(1, &dto.VerifyTotpRequest{OtpCode: "123456"})

	if !errors.Is(err, domain.ErrTotpAlreadyEnabled) {
		t.Errorf("Ожидалась ошибка ErrTotpAlreadyEnabled, получена: %v", err)
	}
}

func TestAuthUseCase_VerifyTotp_SecretNotSet(t *testing.T) {
	userWithoutSecret := &domain.User{
		ID:            1,
		IsTotpEnabled: false,
		TotpSecret:    "",
	}
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return userWithoutSecret, nil
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	err := authUseCase.VerifyTotp(1, &dto.VerifyTotpRequest{OtpCode: "123456"})

	if !errors.Is(err, domain.ErrTotpSecretNotSet) {
		t.Errorf("Ожидалась ошибка ErrTotpSecretNotSet, получена: %v", err)
	}
}

func TestAuthUseCase_VerifyTotp_InvalidCode(t *testing.T) {
	userWithSecret := &domain.User{
		ID:            1,
		IsTotpEnabled: false,
		TotpSecret:    "SECRET123",
	}
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return userWithSecret, nil
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{
		ValidateCodeFunc: func(secret, code string) bool {
			return false
		},
	}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	err := authUseCase.VerifyTotp(1, &dto.VerifyTotpRequest{OtpCode: "wrongcode"})

	if !errors.Is(err, domain.ErrInvalidTotpCode) {
		t.Errorf("Ожидалась ошибка ErrInvalidTotpCode, получена: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_Success(t *testing.T) {
	cacheUserInvalidated := false

	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return userWithTotp, nil
		},
		UpdateFunc: func(id int, updates map[string]any) (*domain.User, error) {
			if updates["is_totp_enabled"] != false {
				t.Errorf("Ожидалось is_totp_enabled=false, получено: %v", updates["is_totp_enabled"])
			}
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{
		ValidateCodeFunc: func(secret, code string) bool {
			return true
		},
	}
	stubCacheService := &StubCacheService{
		DeleteFunc: func(ctx context.Context, key string) error {
			cacheUserInvalidated = true
			return nil
		},
	}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	err := authUseCase.DisableTotp(1, &dto.DisableTotpRequest{
		CurrentPassword: "password123",
		OtpCode:         "123456",
	})

	if err != nil {
		t.Fatalf("Ожидался успех, получена ошибка: %v", err)
	}
	if !cacheUserInvalidated {
		t.Error("Ожидалась инвалидация кэша пользователя, но она не была вызвана")
	}
}

func TestAuthUseCase_DisableTotp_UserNotFound(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	err := authUseCase.DisableTotp(999, &dto.DisableTotpRequest{
		CurrentPassword: "pass",
		OtpCode:         "123456",
	})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Ожидалась ошибка ErrUserNotFound, получена: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_NotEnabled(t *testing.T) {
	userWithoutTotp := &domain.User{
		ID:            1,
		IsTotpEnabled: false,
	}
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return userWithoutTotp, nil
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	err := authUseCase.DisableTotp(1, &dto.DisableTotpRequest{
		CurrentPassword: "pass",
		OtpCode:         "123456",
	})

	if !errors.Is(err, domain.ErrTotpAlreadyDisabled) {
		t.Errorf("Ожидалась ошибка ErrTotpAlreadyDisabled, получена: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_InvalidOtpCode(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{
		ValidateCodeFunc: func(secret, code string) bool {
			return false
		},
	}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	err := authUseCase.DisableTotp(1, &dto.DisableTotpRequest{
		CurrentPassword: "password123",
		OtpCode:         "wrongcode",
	})

	if !errors.Is(err, domain.ErrInvalidTotpCode) {
		t.Errorf("Ожидалась ошибка ErrInvalidTotpCode, получена: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_WrongPassword(t *testing.T) {
	stubUserPostgres := &StubUserPostgres{
		GetByIDFunc: func(id int) (*domain.User, error) {
			return userWithTotp, nil
		},
	}
	stubOutbox := &StubOutboxPostgres{}
	stubTotp := &StubTotpService{
		ValidateCodeFunc: func(secret, code string) bool {
			return true
		},
	}
	stubCacheService := &StubCacheService{}
	stubBlacklistService := &StubBlacklistService{}

	authUseCase := usecase.NewAuthUseCase(stubUserPostgres, stubOutbox, stubTotp, stubCacheService, stubBlacklistService)
	err := authUseCase.DisableTotp(1, &dto.DisableTotpRequest{
		CurrentPassword: "wrongpass",
		OtpCode:         "123456",
	})

	if !errors.Is(err, domain.ErrCurPasswordIsNotCorrect) {
		t.Errorf("Ожидалась ошибка ErrCurPasswordIsNotCorrect, получена: %v", err)
	}
}

func TestAuthUseCase_Logout_Success(t *testing.T) {
	blacklistCalled := false
	var receivedJTI string
	var receivedExpiresAt time.Time

	stubBlacklist := &StubBlacklistService{
		AddToBlacklistFunc: func(ctx context.Context, jti string, expiresAt time.Time) error {
			blacklistCalled = true
			receivedJTI = jti
			receivedExpiresAt = expiresAt
			return nil
		},
	}

	stubUserPostgres := &StubUserPostgres{}
	stubOutboxRepo := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewAuthUseCase(
		stubUserPostgres,
		stubOutboxRepo,
		stubTotpService,
		stubCacheService,
		stubBlacklist,
	)

	userID := 1
	jti := "test-jti-123"
	expiresAt := time.Now().Add(1 * time.Hour)

	err := useCase.Logout(userID, jti, expiresAt)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if !blacklistCalled {
		t.Error("AddToBlacklist should be called")
	}
	if receivedJTI != jti {
		t.Errorf("Expected JTI %s, got %s", jti, receivedJTI)
	}
	if receivedExpiresAt.Sub(expiresAt) > time.Second || receivedExpiresAt.Sub(expiresAt) < -time.Second {
		t.Errorf("Expected expiresAt %v, got %v", expiresAt, receivedExpiresAt)
	}
}

func TestAuthUseCase_Logout_BlacklistError(t *testing.T) {
	blacklistError := errors.New("redis connection failed")

	stubBlacklist := &StubBlacklistService{
		AddToBlacklistFunc: func(ctx context.Context, jti string, expiresAt time.Time) error {
			return blacklistError
		},
	}

	stubUserPostgres := &StubUserPostgres{}
	stubOutboxRepo := &StubOutboxPostgres{}
	stubTotpService := &StubTotpService{}
	stubCacheService := &StubCacheService{}

	useCase := usecase.NewAuthUseCase(
		stubUserPostgres,
		stubOutboxRepo,
		stubTotpService,
		stubCacheService,
		stubBlacklist,
	)

	err := useCase.Logout(1, "test-jti", time.Now().Add(1*time.Hour))

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !errors.Is(err, blacklistError) {
		t.Errorf("Expected blacklist error, got: %v", err)
	}
}
