package usecase

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/entity"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail            = errors.New("invalid email")
	ErrInvalidPassword         = errors.New("invalid password")
	ErrCurPasswordIsNotCorrect = errors.New("current password is not correct")
	ErrSamePassword            = errors.New("entered password is the same as current")
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

func (a *AuthUsecase) ChangePassword(id int, req *entity.ChangePasswordRequest) error {
	user, err := a.userUsecase.GetUserByID(id)
	if err != nil {
		return ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword))
	if err != nil {
		return ErrCurPasswordIsNotCorrect
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.NewPassword))
	if err == nil {
		return ErrSamePassword
	}

	newHashPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"password_hash": string(newHashPassword),
	}
	_, err = a.userUsecase.userRepo.Update(user.ID, updates)

	return err
}
