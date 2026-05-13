package usecase

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type UserUseCaseInterface interface {
	GetUserByEmail(email string) (*entity.User, error)
	CreateUser(req *entity.CreateUserRequest) (*entity.User, error)
	GetUserByID(id int) (*entity.User, error)
	UpdateUser(id int, req *entity.UpdateUserRequest) (*entity.User, error)
	ChangePassword(id int, newPassword string) error
}

type AuthNotifier interface {
	CreateNotification(userID int, notificationType enum.TypeOfNotification, params entity.NotificationParams) error
}

type AuthUseCase struct {
	userUseCase UserUseCaseInterface
	notifier    AuthNotifier
}

func NewAuthUseCase(userUseCase UserUseCaseInterface, notifier AuthNotifier) *AuthUseCase {
	return &AuthUseCase{userUseCase: userUseCase, notifier: notifier}
}

func (a *AuthUseCase) SendAuthNotification(userID int, bookingType enum.TypeOfNotification, params entity.NotificationParams) {
	err := a.notifier.CreateNotification(userID, bookingType, params)
	if err != nil {
		log.Printf("[Notifications] Ошибка при создании уведомления авторизации: %v", err.Error())
	}
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

	params := entity.NotificationParams{
		IP: "127.0.0.1",
	}
	a.SendAuthNotification(user.ID, enum.AuthType, params)

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
