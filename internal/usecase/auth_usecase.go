package usecase

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/entity"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
)

type AuthUsecase struct {
	userUsecase *UserUsecase
}

func NewAuthUsecase(userUsec *UserUsecase) *AuthUsecase {
	return &AuthUsecase{userUsecase: userUsec}
}

func (a *AuthUsecase) Login(email string, password string) (string, error) {
	user, err := a.userUsecase.GetUserByEmail(email)
	if err != nil {
		return "", ErrInvalidEmail
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", ErrInvalidPassword
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *AuthUsecase) Register(req *entity.CreateUserRequest) (string, error) {
	user, err := a.userUsecase.CreateUser(req)
	if err != nil {
		return "", err
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *AuthUsecase) GetUserByID(id int) (*entity.User, error) {
	return a.userUsecase.GetUserByID(id)
}

func (a *AuthUsecase) UpdateUser(id int, req *entity.UpdateUserRequest) (*entity.User, error) {
	return a.userUsecase.UpdateUser(id, req)
}
