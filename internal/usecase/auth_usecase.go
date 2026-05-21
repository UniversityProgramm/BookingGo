package usecase

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

type UserUseCaseInterface interface {
	GetUserByEmail(email string) (*entity.User, error)
	CreateUser(req *entity.CreateUserRequest) (*entity.User, error)
	GetUserByID(id int) (*entity.User, error)
	UpdateUser(id int, req *entity.UpdateUserRequest) (*entity.User, error)
	ChangePassword(id int, newPassword string) error
	ChangeEmail(id int, newEmail string) error
}

type OutboxAuthRepository interface {
	CreateEvent(event *entity.OutboxEvent) error
}

type AuthUseCase struct {
	userUseCase UserUseCaseInterface
	outboxRepo  OutboxAuthRepository
}

func NewAuthUseCase(userUseCase UserUseCaseInterface, outboxRepo OutboxAuthRepository) *AuthUseCase {
	return &AuthUseCase{userUseCase: userUseCase, outboxRepo: outboxRepo}
}

func (a *AuthUseCase) Login(email string, password string) (string, error) {
	user, err := a.userUseCase.GetUserByEmail(email)
	if err != nil {
		return "", domain.ErrInvalidEmail
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", domain.ErrInvalidPassword
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return "", err
	}

	payload := map[string]any{
		"user_id":    user.ID,
		"booking_id": -1,
		"ip":         "127.0.0.1",
	}
	outboxEvent, err := entity.NewOutboxEvent(enum.AuthType, payload)
	if err != nil {
		logger.Log.Error("[AuthUseCase] Failed to create outboxEvent", "eventType", enum.AuthType, "error", err.Error())
	}

	err = a.outboxRepo.CreateEvent(outboxEvent)
	if err != nil {
		logger.Log.Error("[AuthUseCase] Failed to write outboxEvent", "eventType", enum.AuthType, "error", err.Error())
	}

	return token, nil
}

func (a *AuthUseCase) Register(req *entity.CreateUserRequest) (string, error) {
	user, err := a.userUseCase.CreateUser(req)
	if err != nil {
		return "", err
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *AuthUseCase) GetMe(id int) (*entity.User, error) {
	return a.userUseCase.GetUserByID(id)
}

func (a *AuthUseCase) UpdateMe(id int, req *entity.UpdateUserRequest) (*entity.User, error) {
	return a.userUseCase.UpdateUser(id, req)
}

func (a *AuthUseCase) ChangePassword(id int, req *entity.ChangePasswordRequest) error {
	user, err := a.userUseCase.GetUserByID(id)
	if err != nil {
		return domain.ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword))
	if err != nil {
		return domain.ErrCurPasswordIsNotCorrect
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.NewPassword))
	if err == nil {
		return domain.ErrSamePassword
	}

	newHashPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return a.userUseCase.ChangePassword(user.ID, string(newHashPassword))
}

func (a *AuthUseCase) ChangeEmail(id int, req *entity.ChangeEmailRequest) error {
	user, err := a.userUseCase.GetUserByID(id)
	if err != nil {
		return domain.ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.ConfirmPassword))
	if err != nil {
		return domain.ErrCurPasswordIsNotCorrect
	}

	return a.userUseCase.ChangeEmail(user.ID, req.NewEmail)
}
