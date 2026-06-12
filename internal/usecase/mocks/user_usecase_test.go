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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := NewMockUserRepository(ctrl)

	mockUserRepo.EXPECT().
		GetAll().
		Return([]entity.User{*testUser}, nil)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	_, err := useCase.GetAllUsers()
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestUserUseCase_GetUserByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := NewMockUserRepository(ctrl)

	mockUserRepo.EXPECT().
		GetByID(1).
		Return(testUser, nil)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	user, err := useCase.GetUserByID(1)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if user == nil {
		t.Error("User should not be nil")
	}
	if user.Email != testUser.Email {
		t.Errorf("Expected email %s, got %s", testUser.Email, user.Email)
	}
}

func TestUserUseCase_GetUserByID_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := NewMockUserRepository(ctrl)

	mockUserRepo.EXPECT().
		GetByID(1).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	_, err := useCase.GetUserByID(1)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got error: %v", err)
	}
}

func TestUserUseCase_GetUserByEmail_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := NewMockUserRepository(ctrl)

	mockUserRepo.EXPECT().
		GetByEmail(testUser.Email).
		Return(testUser, nil)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	user, err := useCase.GetUserByEmail(testUser.Email)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if user == nil {
		t.Error("User should not be nil")
	}
	if user.Email != testUser.Email {
		t.Errorf("Expected email %s, got %s", testUser.Email, user.Email)
	}
}

func TestUserUseCase_GetUserByEmail_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := NewMockUserRepository(ctrl)

	mockUserRepo.EXPECT().
		GetByEmail("not_exist_email@example.com").
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	_, err := useCase.GetUserByEmail("not_exist_email@example.com")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got error: %v", err)
	}
}

func TestUserUseCase_GetUserByEmailContext_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := NewMockUserRepository(ctrl)

	mockUserRepo.EXPECT().
		GetByEmailContext(context.Background(), testUser.Email).
		Return(testUser, nil)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	user, err := useCase.GetUserByEmailContext(context.Background(), testUser.Email)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if user == nil {
		t.Error("User should not be nil")
	}
	if user.Email != testUser.Email {
		t.Errorf("Expected email %s, got %s", testUser.Email, user.Email)
	}
}

func TestUserUseCase_GetUserByEmailContext_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := NewMockUserRepository(ctrl)

	mockUserRepo.EXPECT().
		GetByEmailContext(context.Background(), "not_exist_email@example.com").
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	_, err := useCase.GetUserByEmailContext(context.Background(), "not_exist_email@example.com")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got error: %v", err)
	}
}

func TestUserUseCase_CreateUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := NewMockUserRepository(ctrl)

	mockUserRepo.EXPECT().
		EmailExists("new@example.com").
		Return(false, nil)

	mockUserRepo.EXPECT().
		Create(gomock.Any()).
		DoAndReturn(func(user *entity.User) error {
			if user.PasswordHash == "" {
				t.Error("Password hash is empty")
			}
			if user.PasswordHash == "password123" {
				t.Error("Password is not hashed!")
			}

			return nil
		})

	useCase := usecase.NewUserUseCase(mockUserRepo)

	req := &entity.CreateUserRequest{
		Email:    "new@example.com",
		Password: "password123",
		FIO:      "Test User",
		Phone:    "+79991234567",
	}

	user, err := useCase.CreateUser(req)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if user == nil {
		t.Error("User should not be nil")
	}
	if user.Email != "new@example.com" {
		t.Errorf("Expected email 'new@example.com', got '%s'", user.Email)
	}
}

func TestUserUseCase_CreateUser_EmailTaken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := NewMockUserRepository(ctrl)

	mockUserRepo.EXPECT().
		EmailExists("new@example.com").
		Return(true, nil)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	req := &entity.CreateUserRequest{
		Email:    "new@example.com",
		Password: "password123",
		FIO:      "Test User",
		Phone:    "+79991234567",
	}

	_, err := useCase.CreateUser(req)
	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("Expected Email taken, got error: %v", err)
	}
}

func TestUserUseCase_UpdateUserProfile_Success_1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)

	userID := 1
	newFIO := "Updated Name"
	newPhone := "+79999999999"

	mockRepo.EXPECT().
		GetByID(userID).
		Return(&entity.User{ID: userID, Email: "test@example.com"}, nil)

	mockRepo.EXPECT().
		Update(userID, gomock.Cond(func(updates map[string]any) bool {
			if updates["fio"] != newFIO {
				t.Errorf("Expected FIO '%s', got '%v'", newFIO, updates["fio"])
			}
			if updates["phone"] != newPhone {
				t.Errorf("Expected Phone '%s', got '%v'", newPhone, updates["phone"])
			}
			return true
		})).
		Return(&entity.User{
			ID:    userID,
			FIO:   newFIO,
			Phone: newPhone,
		}, nil)

	useCase := usecase.NewUserUseCase(mockRepo)

	req := &entity.UpdateUserProfileRequest{
		FIO:   &newFIO,
		Phone: &newPhone,
	}

	user, err := useCase.UpdateUserProfile(userID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if user.FIO != newFIO {
		t.Errorf("Expected FIO '%s', got '%s'", newFIO, user.FIO)
	}
	if user.Phone != newPhone {
		t.Errorf("Expected Phone '%s', got '%s'", newPhone, user.Phone)
	}
}

func TestUserUseCase_UpdateUserProfile_Success_2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)

	userID := 1
	newFIO := "Updated Name"

	mockRepo.EXPECT().
		GetByID(userID).
		Return(&entity.User{ID: userID, Email: "test@example.com"}, nil)

	mockRepo.EXPECT().
		Update(userID, gomock.Cond(func(updates map[string]any) bool {
			if updates["fio"] != newFIO {
				t.Errorf("Expected FIO '%s', got '%v'", newFIO, updates["fio"])
			}
			return true
		})).
		Return(&entity.User{
			ID:  userID,
			FIO: newFIO,
		}, nil)

	useCase := usecase.NewUserUseCase(mockRepo)

	req := &entity.UpdateUserProfileRequest{
		FIO: &newFIO,
	}

	user, err := useCase.UpdateUserProfile(userID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if user.FIO != newFIO {
		t.Errorf("Expected FIO '%s', got '%s'", newFIO, user.FIO)
	}
}

func TestUserUseCase_UpdateUserProfile_Success_3(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)

	userID := 1
	newPhone := "+79999999999"

	mockRepo.EXPECT().
		GetByID(userID).
		Return(&entity.User{ID: userID, Email: "test@example.com"}, nil)

	mockRepo.EXPECT().
		Update(userID, gomock.Cond(func(updates map[string]any) bool {
			if updates["phone"] != newPhone {
				t.Errorf("Expected Phone '%s', got '%v'", newPhone, updates["phone"])
			}
			return true
		})).
		Return(&entity.User{
			ID:    userID,
			Phone: newPhone,
		}, nil)

	useCase := usecase.NewUserUseCase(mockRepo)

	req := &entity.UpdateUserProfileRequest{
		Phone: &newPhone,
	}

	user, err := useCase.UpdateUserProfile(userID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if user.Phone != newPhone {
		t.Errorf("Expected Phone '%s', got '%s'", newPhone, user.Phone)
	}
}

func TestUserUseCase_UpdateUserProfile_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)

	userID := 1
	newPhone := "+79999999999"

	mockRepo.EXPECT().
		GetByID(userID).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewUserUseCase(mockRepo)

	req := &entity.UpdateUserProfileRequest{
		Phone: &newPhone,
	}

	_, err := useCase.UpdateUserProfile(userID, req)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got error: %v", err)
	}
}

func TestUserUseCase_UpdateUserData_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)

	userID := 1
	newEmail := "new_email@example.com"

	mockRepo.EXPECT().
		GetByID(userID).
		Return(&entity.User{ID: userID, Email: "test@example.com"}, nil)

	mockRepo.EXPECT().
		Update(userID, gomock.Cond(func(updates map[string]any) bool {
			if updates["email"] != newEmail {
				t.Errorf("Expected Email '%s', got '%v'", newEmail, updates["email"])
			}
			return true
		})).
		Return(&entity.User{
			ID:    userID,
			Email: newEmail,
		}, nil)

	useCase := usecase.NewUserUseCase(mockRepo)

	req := map[string]any{"email": newEmail}

	user, err := useCase.UpdateUserData(userID, req)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if user.Email != newEmail {
		t.Errorf("Expected Email '%s', got '%s'", newEmail, user.Email)
	}
}

func TestUserUseCase_UpdateUserData_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)

	userID := 1
	newEmail := "new_email@example.com"

	mockRepo.EXPECT().
		GetByID(userID).
		Return(nil, domain.ErrUserNotFound)

	useCase := usecase.NewUserUseCase(mockRepo)

	req := map[string]any{"email": newEmail}

	_, err := useCase.UpdateUserData(userID, req)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got error: %v", err)
	}
}

func TestUserUseCase_DeleteUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)

	mockRepo.EXPECT().
		Delete(1).
		Return(nil)

	useCase := usecase.NewUserUseCase(mockRepo)
	err := useCase.DeleteUser(1)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestUserUseCase_DeleteUser_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)

	mockRepo.EXPECT().
		Delete(1).
		Return(domain.ErrUserNotFound)

	useCase := usecase.NewUserUseCase(mockRepo)
	err := useCase.DeleteUser(1)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got error: %v", err)
	}
}

func TestUserUseCase_EmailExists_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)

	mockRepo.EXPECT().
		EmailExists("exist@example.com").
		Return(true, nil)

	useCase := usecase.NewUserUseCase(mockRepo)

	_, err := useCase.EmailExists("exist@example.com")

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestUserUseCase_EmailExists_RepoErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)

	dbError := errors.New("database connection failed")

	mockRepo.EXPECT().
		EmailExists("not_exist@example.com").
		Return(false, dbError)

	useCase := usecase.NewUserUseCase(mockRepo)

	_, err := useCase.EmailExists("not_exist@example.com")

	if !errors.Is(err, dbError) {
		t.Errorf("Expected dbError, got error: %v", err)
	}
}
