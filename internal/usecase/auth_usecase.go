package usecase

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"

	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	userUseCase *UserUseCase
}

func NewAuthUseCase(userUsec *UserUseCase) *AuthUseCase {
	return &AuthUseCase{userUseCase: userUsec}
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

func (a *AuthUseCase) GetUserByID(id int) (*entity.User, error) {
	return a.userUseCase.GetUserByID(id)
}

func (a *AuthUseCase) UpdateUser(id int, req *entity.UpdateUserRequest) (*entity.User, error) {
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

	updates := map[string]interface{}{
		"password_hash": string(newHashPassword),
	}
	_, err = a.userUseCase.userRepo.Update(user.ID, updates)

	return err
}
