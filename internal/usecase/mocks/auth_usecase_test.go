package mocks

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/internal/usecase"
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

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

var userWithTotp = &entity.User{
	ID:            1,
	IsTotpEnabled: true,
	TotpSecret:    "SECRET123",
	PasswordHash:  testAuthUser.PasswordHash,
}

var userWithSecret = &entity.User{
	ID:            3,
	Email:         "user@example.com",
	PasswordHash:  testAuthUser.PasswordHash,
	IsTotpEnabled: false,
	TotpSecret:    "SECRET123",
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByEmail(testAuthUser.Email).
		Return(testAuthUser, nil)

	mockOutboxRepo.EXPECT().
		CreateEvent(gomock.Any()).
		Return(nil).AnyTimes()

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.LoginUserRequest{
		Email:    testAuthUser.Email,
		Password: "password123",
	}

	token, err := useCase.Login(req)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if token == "" {
		t.Error("Token was not generated")
	}
}

func TestAuthUseCase_Login_InvalidEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByEmail("not_exist@example.com").
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.LoginUserRequest{
		Email:    "not_exist@example.com",
		Password: "password123",
	}

	_, err := useCase.Login(req)

	if !errors.Is(err, domain.ErrInvalidEmail) {
		t.Errorf("Expected ErrInvalidEmail, got error: %v", err)
	}
}

func TestAuthUseCase_Login_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByEmail(testAuthUser.Email).
		Return(testAuthUser, nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.LoginUserRequest{
		Email:    testAuthUser.Email,
		Password: "password1234",
	}

	_, err := useCase.Login(req)

	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Errorf("Expected ErrInvalidPassword, got error: %v", err)
	}
}

func TestAuthUseCase_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	newUser := &entity.User{
		ID:    10,
		Email: "new@example.com",
		Role:  enum.RoleClient,
	}

	mockUserUseCase.EXPECT().
		CreateUser(gomock.Cond(func(req *entity.CreateUserRequest) bool {
			return req.Email == "new@example.com"
		})).
		Return(newUser, nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.CreateUserRequest{
		Email:    "new@example.com",
		Password: "password123",
		FIO:      "New User",
	}

	token, err := useCase.Register(req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if token == "" {
		t.Error("Token should not be empty")
	}
}

func TestAuthUseCase_Register_EmailTaken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		CreateUser(gomock.Cond(func(req *entity.CreateUserRequest) bool {
			return req.Email == "taken@example.com"
		})).
		Return(nil, domain.ErrEmailTaken)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.CreateUserRequest{
		Email:    "taken@example.com",
		Password: "password123",
		FIO:      "New User",
	}

	_, err := useCase.Register(req)

	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("Expected ErrEmailTaken, got error: %v", err)
	}
}

func TestAuthUseCase_GetMe_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockCache.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.ErrCacheKeyNotFound)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(testAuthUser, nil)

	mockCache.EXPECT().
		Set(gomock.Any(), gomock.Any(), gomock.Any(), 300*time.Second).
		Return(nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	user, err := useCase.GetMe(testAuthUser.ID)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if user.Email != testAuthUser.Email {
		t.Errorf("Expected email %s, got %s", testAuthUser.Email, user.Email)
	}
}

func TestAuthUseCase_GetMe_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	dbCalled := false

	mockCache.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, dest any) error {
			if userPtr, ok := dest.(*entity.User); ok {
				*userPtr = *testAuthUser
			}
			return nil
		})

	mockUserUseCase.EXPECT().
		GetUserByID(gomock.Any()).
		DoAndReturn(func(id int) (*entity.User, error) {
			dbCalled = true
			return testAuthUser, nil
		}).
		MaxTimes(0)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	user, err := useCase.GetMe(testAuthUser.ID)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if user.Email != testAuthUser.Email {
		t.Errorf("Expected email %s, got %s", testAuthUser.Email, user.Email)
	}
	if dbCalled {
		t.Error("GetUserByID should not be called on cache hit")
	}
}

func TestAuthUseCase_GetMe_CacheMiss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	cacheSetCalled := false

	mockCache.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.ErrCacheKeyNotFound)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(testAuthUser, nil)

	mockCache.EXPECT().
		Set(gomock.Any(), gomock.Any(), gomock.Any(), 300*time.Second).
		DoAndReturn(func(ctx context.Context, key string, value any, ttl time.Duration) error {
			cacheSetCalled = true
			return nil
		})

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	user, err := useCase.GetMe(testAuthUser.ID)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if user.Email != testAuthUser.Email {
		t.Errorf("Expected email %s, got %s", testAuthUser.Email, user.Email)
	}
	if !cacheSetCalled {
		t.Error("Cache Set should be called on cache miss")
	}
}

func TestAuthUseCase_GetMe_CacheSetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockCache.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.ErrCacheKeyNotFound)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(testAuthUser, nil)

	mockCache.EXPECT().
		Set(gomock.Any(), gomock.Any(), gomock.Any(), 300*time.Second).
		Return(errors.New("cache error"))

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	user, err := useCase.GetMe(testAuthUser.ID)

	if err != nil {
		t.Errorf("Cache error should not break flow, got: %v", err)
	}
	if user.Email != testAuthUser.Email {
		t.Errorf("Expected email %s, got %s", testAuthUser.Email, user.Email)
	}
}

func TestAuthUseCase_GetMe_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockCache.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(domain.ErrCacheKeyNotFound)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	_, err := useCase.GetMe(testAuthUser.ID)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got error: %v", err)
	}
}

func TestAuthUseCase_UpdateMe_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	newFIO := "Updated Name"
	req := &entity.UpdateUserProfileRequest{FIO: &newFIO}

	updatedUser := *testAuthUser
	updatedUser.FIO = newFIO

	mockUserUseCase.EXPECT().
		UpdateUserProfile(testAuthUser.ID, req).
		Return(&updatedUser, nil)

	mockCache.EXPECT().
		Delete(gomock.Any(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	user, err := useCase.UpdateMe(testAuthUser.ID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if user.Email != testAuthUser.Email {
		t.Errorf("Expected email %s, got %s", testAuthUser.Email, user.Email)
	}
}

func TestAuthUseCase_UpdateMe_CacheInvalidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	newFIO := "Updated Name"
	req := &entity.UpdateUserProfileRequest{FIO: &newFIO}

	updatedUser := *testAuthUser
	updatedUser.FIO = newFIO

	mockUserUseCase.EXPECT().
		UpdateUserProfile(testAuthUser.ID, req).
		Return(&updatedUser, nil)

	mockCache.EXPECT().
		Delete(gomock.Any(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	_, err := useCase.UpdateMe(testAuthUser.ID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestAuthUseCase_UpdateMe_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	newFIO := "Updated Name"
	req := &entity.UpdateUserProfileRequest{FIO: &newFIO}

	updatedUser := *testAuthUser
	updatedUser.FIO = newFIO

	mockUserUseCase.EXPECT().
		UpdateUserProfile(testAuthUser.ID, req).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	_, err := useCase.UpdateMe(testAuthUser.ID, req)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got error: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(testAuthUser, nil)

	mockUserUseCase.EXPECT().
		UpdateUserData(testAuthUser.ID, gomock.Cond(func(updates map[string]any) bool {
			hash, ok := updates["password_hash"].(string)
			if !ok || hash == "" {
				t.Error("Password hash should not be empty")
			}
			return true
		})).
		Return(testAuthUser, nil)

	mockOutboxRepo.EXPECT().
		CreateEvent(gomock.Any()).
		Return(nil).AnyTimes()

	blacklistService.EXPECT().
		InvalidateAllSessions(context.Background(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newPassword456",
	}

	err := useCase.ChangePassword(testAuthUser.ID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_WrongCurrentPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(testAuthUser, nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangePasswordRequest{
		CurrentPassword: "wong_pass",
		NewPassword:     "newPassword456",
	}

	err := useCase.ChangePassword(testAuthUser.ID, req)

	if !errors.Is(err, domain.ErrCurPasswordIsNotCorrect) {
		t.Errorf("Expected ErrCurPasswordIsNotCorrect, got error: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_SamePassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(testAuthUser, nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "password123",
	}

	err := useCase.ChangePassword(testAuthUser.ID, req)

	if !errors.Is(err, domain.ErrSamePassword) {
		t.Errorf("Expected ErrSamePassword, got error: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "password123",
	}

	err := useCase.ChangePassword(testAuthUser.ID, req)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got error: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_WithTotp_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithTotp.ID).
		Return(userWithTotp, nil)

	mockUserUseCase.EXPECT().
		UpdateUserData(userWithTotp.ID, gomock.Any()).
		Return(userWithTotp, nil)

	mockTotpService.EXPECT().
		ValidateCode(userWithTotp.TotpSecret, "123456").
		Return(true)

	mockOutboxRepo.EXPECT().
		CreateEvent(gomock.Any()).
		Return(nil).AnyTimes()

	blacklistService.EXPECT().
		InvalidateAllSessions(context.Background(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newPassword456",
		OtpCode:         "123456",
	}

	err := useCase.ChangePassword(userWithTotp.ID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_TotpRequired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithTotp.ID).
		Return(userWithTotp, nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newPassword456",
		OtpCode:         "",
	}

	err := useCase.ChangePassword(userWithTotp.ID, req)

	if !errors.Is(err, domain.ErrTotpSecretNotSet) {
		t.Errorf("Expected ErrTotpSecretNotSet, got error: %v", err)
	}
}

func TestAuthUseCase_ChangePassword_InvalidCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithTotp.ID).
		Return(userWithTotp, nil)

	mockTotpService.EXPECT().
		ValidateCode(userWithTotp.TotpSecret, "000000").
		Return(false)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newPassword456",
		OtpCode:         "000000",
	}

	err := useCase.ChangePassword(userWithTotp.ID, req)

	if !errors.Is(err, domain.ErrInvalidTotpCode) {
		t.Errorf("Expected ErrInvalidTotpCode, got error: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(testAuthUser, nil)

	mockUserUseCase.EXPECT().
		EmailExists("newemail@example.com").
		Return(false, nil)

	mockUserUseCase.EXPECT().
		UpdateUserData(testAuthUser.ID, gomock.Any()).
		Return(&entity.User{ID: testAuthUser.ID, Email: "newemail@example.com"}, nil)

	mockCache.EXPECT().
		Delete(gomock.Any(), gomock.Any()).
		Return(nil)

	mockOutboxRepo.EXPECT().
		CreateEvent(gomock.Any()).
		Return(nil).AnyTimes()

	blacklistService.EXPECT().
		InvalidateAllSessions(context.Background(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangeEmailRequest{
		NewEmail:        "newemail@example.com",
		ConfirmPassword: "password123",
	}

	user, err := useCase.ChangeEmail(testAuthUser.ID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if user.Email != "newemail@example.com" {
		t.Errorf("Expected email 'newemail@example.com', got '%s'", user.Email)
	}
}

func TestAuthUseCase_ChangeEmail_CacheInvalidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(testAuthUser, nil)

	mockUserUseCase.EXPECT().
		EmailExists("new@example.com").
		Return(false, nil)

	mockUserUseCase.EXPECT().
		UpdateUserData(testAuthUser.ID, gomock.Any()).
		Return(&entity.User{ID: testAuthUser.ID, Email: "new@example.com"}, nil)

	mockCache.EXPECT().
		Delete(gomock.Any(), gomock.Any()).
		Return(nil)

	mockOutboxRepo.EXPECT().
		CreateEvent(gomock.Any()).
		Return(nil).AnyTimes()

	blacklistService.EXPECT().
		InvalidateAllSessions(context.Background(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangeEmailRequest{
		NewEmail:        "new@example.com",
		ConfirmPassword: "password123",
	}

	_, err := useCase.ChangeEmail(testAuthUser.ID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_EmailTaken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(testAuthUser, nil)

	mockUserUseCase.EXPECT().
		EmailExists("taken@example.com").
		Return(true, nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangeEmailRequest{
		NewEmail:        "taken@example.com",
		ConfirmPassword: "password123",
	}

	_, err := useCase.ChangeEmail(testAuthUser.ID, req)

	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("Expected ErrEmailTaken, got: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_InvalidCurrentPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(testAuthUser, nil)

	mockUserUseCase.EXPECT().
		EmailExists("new@example.com").
		Return(false, nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangeEmailRequest{
		NewEmail:        "new@example.com",
		ConfirmPassword: "wrong_password",
	}

	_, err := useCase.ChangeEmail(testAuthUser.ID, req)

	if !errors.Is(err, domain.ErrCurPasswordIsNotCorrect) {
		t.Errorf("Expected ErrCurPasswordIsNotCorrect, got: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangeEmailRequest{
		NewEmail:        "taken@example.com",
		ConfirmPassword: "password123",
	}

	_, err := useCase.ChangeEmail(testAuthUser.ID, req)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_WithTotp_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithTotp.ID).
		Return(userWithTotp, nil)

	mockUserUseCase.EXPECT().
		EmailExists("newemail@example.com").
		Return(false, nil)

	mockUserUseCase.EXPECT().
		UpdateUserData(userWithTotp.ID, gomock.Any()).
		Return(&entity.User{Email: "newemail@example.com"}, nil)

	mockTotpService.EXPECT().
		ValidateCode(userWithTotp.TotpSecret, "123456").
		Return(true)

	mockCache.EXPECT().
		Delete(gomock.Any(), gomock.Any()).
		Return(nil)

	mockOutboxRepo.EXPECT().
		CreateEvent(gomock.Any()).
		Return(nil).AnyTimes()

	blacklistService.EXPECT().
		InvalidateAllSessions(context.Background(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangeEmailRequest{
		NewEmail:        "newemail@example.com",
		ConfirmPassword: "password123",
		OtpCode:         "123456",
	}

	user, err := useCase.ChangeEmail(testAuthUser.ID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if user.Email != "newemail@example.com" {
		t.Errorf("Expected email 'newemail@example.com', got '%s'", user.Email)
	}
}

func TestAuthUseCase_ChangeEmail_TotpRequired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithTotp.ID).
		Return(userWithTotp, nil)

	mockUserUseCase.EXPECT().
		EmailExists("newemail@example.com").
		Return(false, nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangeEmailRequest{
		NewEmail:        "newemail@example.com",
		ConfirmPassword: "password123",
		OtpCode:         "",
	}

	_, err := useCase.ChangeEmail(testAuthUser.ID, req)

	if !errors.Is(err, domain.ErrTotpSecretNotSet) {
		t.Errorf("Expected ErrTotpSecretNotSet, got error: %v", err)
	}
}

func TestAuthUseCase_ChangeEmail_InvalidCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithTotp.ID).
		Return(userWithTotp, nil)

	mockUserUseCase.EXPECT().
		EmailExists("newemail@example.com").
		Return(false, nil)

	mockTotpService.EXPECT().
		ValidateCode(userWithTotp.TotpSecret, "000000").
		Return(false)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.ChangeEmailRequest{
		NewEmail:        "newemail@example.com",
		ConfirmPassword: "password123",
		OtpCode:         "000000",
	}

	_, err := useCase.ChangeEmail(testAuthUser.ID, req)

	if !errors.Is(err, domain.ErrInvalidTotpCode) {
		t.Errorf("Expected ErrInvalidTotpCode, got error: %v", err)
	}
}

func TestAuthUseCase_SetupTotp_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(testAuthUser, nil)

	mockTotpService.EXPECT().
		GenerateSecret(testAuthUser.Email).
		Return("SECRET123", "otpauth://totp/BookingGo:test@example.com?secret=SECRET123", nil)

	mockUserUseCase.EXPECT().
		UpdateUserData(testAuthUser.ID, gomock.Cond(func(updates map[string]any) bool {
			return updates["totp_secret"] == "SECRET123"
		})).
		Return(testAuthUser, nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	otpUrl, err := useCase.SetupTotp(testAuthUser.ID)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if otpUrl == "" {
		t.Error("OTP URL should not be empty")
	}
}

func TestAuthUseCase_SetupTotp_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	_, err := useCase.SetupTotp(testAuthUser.ID)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got error: %v", err)
	}
}

func TestAuthUseCase_SetupTotp_AlreadyEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithTotp.ID).
		Return(userWithTotp, nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	_, err := useCase.SetupTotp(userWithTotp.ID)

	if !errors.Is(err, domain.ErrTotpAlreadyEnabled) {
		t.Errorf("Expected ErrTotpAlreadyEnabled, got error: %v", err)
	}
}

func TestAuthUseCase_VerifyTotp_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithSecret.ID).
		Return(userWithSecret, nil)

	mockTotpService.EXPECT().
		ValidateCode(userWithTotp.TotpSecret, "123456").
		Return(true)

	mockUserUseCase.EXPECT().
		UpdateUserData(userWithSecret.ID, gomock.Cond(func(updates map[string]any) bool {
			return updates["is_totp_enabled"] == true
		})).Return(userWithSecret, nil)

	mockCache.EXPECT().
		Delete(gomock.Any(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.VerifyTotpRequest{
		OtpCode: "123456",
	}

	err := useCase.VerifyTotp(userWithSecret.ID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestAuthUseCase_VerifyTotp_CacheInvalidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithSecret.ID).
		Return(userWithSecret, nil)

	mockTotpService.EXPECT().
		ValidateCode(userWithSecret.TotpSecret, "123456").
		Return(true)

	mockUserUseCase.EXPECT().
		UpdateUserData(userWithSecret.ID, gomock.Any()).
		Return(userWithSecret, nil)

	mockCache.EXPECT().
		Delete(gomock.Any(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.VerifyTotpRequest{OtpCode: "123456"}
	err := useCase.VerifyTotp(userWithSecret.ID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestAuthUseCase_VerifyTotp_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithSecret.ID).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.VerifyTotpRequest{
		OtpCode: "123456",
	}

	err := useCase.VerifyTotp(userWithSecret.ID, req)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got error: %v", err)
	}
}

func TestAuthUseCase_VerifyTotp_AlreadyEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithTotp.ID).
		Return(userWithTotp, nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.VerifyTotpRequest{
		OtpCode: "123456",
	}

	err := useCase.VerifyTotp(userWithTotp.ID, req)

	if !errors.Is(err, domain.ErrTotpAlreadyEnabled) {
		t.Errorf("Expected ErrTotpAlreadyEnabled, got error: %v", err)
	}
}

func TestAuthUseCase_VerifyTotp_SecretNotSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(testAuthUser, nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.VerifyTotpRequest{
		OtpCode: "123456",
	}

	err := useCase.VerifyTotp(testAuthUser.ID, req)

	if !errors.Is(err, domain.ErrTotpSecretNotSet) {
		t.Errorf("Expected ErrTotpSecretNotSet, got error: %v", err)
	}
}

func TestAuthUseCase_VerifyTotp_InvalidOtpCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithSecret.ID).
		Return(userWithSecret, nil)

	mockTotpService.EXPECT().
		ValidateCode(userWithSecret.TotpSecret, "123456").
		Return(false)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.VerifyTotpRequest{
		OtpCode: "123456",
	}

	err := useCase.VerifyTotp(userWithSecret.ID, req)

	if !errors.Is(err, domain.ErrInvalidTotpCode) {
		t.Errorf("Expected ErrInvalidTotpCode, got error: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithTotp.ID).
		Return(userWithTotp, nil)

	mockTotpService.EXPECT().
		ValidateCode(userWithTotp.TotpSecret, "123456").
		Return(true)

	mockUserUseCase.EXPECT().
		UpdateUserData(userWithTotp.ID, gomock.Cond(func(updates map[string]any) bool {
			return updates["is_totp_enabled"] == false
		})).
		Return(userWithTotp, nil)

	mockCache.EXPECT().
		Delete(gomock.Any(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.DisableTotpRequest{
		OtpCode:         "123456",
		CurrentPassword: "password123",
	}

	err := useCase.DisableTotp(userWithTotp.ID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_CacheInvalidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithTotp.ID).
		Return(userWithTotp, nil)

	mockTotpService.EXPECT().
		ValidateCode(userWithTotp.TotpSecret, "123456").
		Return(true)

	mockUserUseCase.EXPECT().
		UpdateUserData(userWithTotp.ID, gomock.Any()).
		Return(userWithTotp, nil)

	mockCache.EXPECT().
		Delete(gomock.Any(), gomock.Any()).
		Return(nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.DisableTotpRequest{
		OtpCode:         "123456",
		CurrentPassword: "password123",
	}

	err := useCase.DisableTotp(userWithTotp.ID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithTotp.ID).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.DisableTotpRequest{
		OtpCode:         "123456",
		CurrentPassword: "password123",
	}

	err := useCase.DisableTotp(userWithTotp.ID, req)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got error: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_AlreadyDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(testAuthUser.ID).
		Return(testAuthUser, nil)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.DisableTotpRequest{
		OtpCode:         "123456",
		CurrentPassword: "password123",
	}

	err := useCase.DisableTotp(testAuthUser.ID, req)

	if !errors.Is(err, domain.ErrTotpAlreadyDisabled) {
		t.Errorf("Expected ErrTotpAlreadyDisabled, got error: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithTotp.ID).
		Return(userWithTotp, nil)

	mockTotpService.EXPECT().
		ValidateCode(userWithTotp.TotpSecret, "123456").
		Return(true)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.DisableTotpRequest{
		OtpCode:         "123456",
		CurrentPassword: "wrong",
	}

	err := useCase.DisableTotp(userWithTotp.ID, req)

	if !errors.Is(err, domain.ErrCurPasswordIsNotCorrect) {
		t.Errorf("Expected ErrCurPasswordIsNotCorrect, got error: %v", err)
	}
}

func TestAuthUseCase_DisableTotp_InvalidOtpCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	blacklistService := NewMockBlacklist(ctrl)

	mockUserUseCase.EXPECT().
		GetUserByID(userWithTotp.ID).
		Return(userWithTotp, nil)

	mockTotpService.EXPECT().
		ValidateCode(userWithTotp.TotpSecret, "000000").
		Return(false)

	useCase := usecase.NewAuthUseCase(mockUserUseCase, mockOutboxRepo, mockTotpService, mockCache, blacklistService)

	req := &entity.DisableTotpRequest{
		OtpCode:         "000000",
		CurrentPassword: "password123",
	}

	err := useCase.DisableTotp(userWithTotp.ID, req)

	if !errors.Is(err, domain.ErrInvalidTotpCode) {
		t.Errorf("Expected ErrInvalidTotpCode, got error: %v", err)
	}
}

func TestAuthUseCase_Logout_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := NewMockUserUseCaseInterface(ctrl)
	mockOutboxRepo := NewMockOutboxRepository(ctrl)
	mockTotpService := NewMockTotpService(ctrl)
	mockCache := NewMockCache(ctrl)
	mockBlacklist := NewMockBlacklist(ctrl)

	jti := "test-jti-123"
	expiresAt := time.Now().Add(1 * time.Hour)

	mockBlacklist.EXPECT().
		AddToBlacklist(gomock.Any(), jti, expiresAt).
		Return(nil)

	useCase := usecase.NewAuthUseCase(
		mockUserUseCase,
		mockOutboxRepo,
		mockTotpService,
		mockCache,
		mockBlacklist,
	)

	err := useCase.Logout(1, jti, expiresAt)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
}

func TestAuthUseCase_Logout_BlacklistError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBlacklist := NewMockBlacklist(ctrl)

	blacklistError := errors.New("redis connection failed")

	mockBlacklist.EXPECT().
		AddToBlacklist(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(blacklistError)

	useCase := usecase.NewAuthUseCase(nil, nil, nil, nil, mockBlacklist)

	err := useCase.Logout(1, "test-jti", time.Now().Add(1*time.Hour))

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !errors.Is(err, blacklistError) {
		t.Errorf("Expected blacklist error, got: %v", err)
	}
}
