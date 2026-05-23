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
	UpdateUserProfile(id int, req *entity.UpdateUserRequest) (*entity.User, error)
	UpdateUserData(id int, updates map[string]any) (*entity.User, error)
	EmailExists(email string) (bool, error)
}

type OutboxAuthRepository interface {
	CreateEvent(event *entity.OutboxEvent) error
}

type TotpService interface {
	GenerateSecret(email string) (string, string, error)
	ValidateCode(secret, code string) bool
}

type AuthUseCase struct {
	userUseCase UserUseCaseInterface
	outboxRepo  OutboxAuthRepository
	totpService TotpService
}

func NewAuthUseCase(userUseCase UserUseCaseInterface, outboxRepo OutboxAuthRepository, totpService TotpService) *AuthUseCase {
	return &AuthUseCase{userUseCase: userUseCase, outboxRepo: outboxRepo, totpService: totpService}
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
	return a.userUseCase.UpdateUserProfile(id, req)
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

	if user.IsTotpEnabled {
		if req.OtpCode == "" {
			return domain.ErrTotpSecretNotSet
		}
		if !a.totpService.ValidateCode(user.TotpSecret, req.OtpCode) {
			return domain.ErrInvalidTotpCode
		}
	}

	updates := map[string]any{
		"password_hash": string(newHashPassword),
	}
	_, err = a.userUseCase.UpdateUserData(user.ID, updates)

	payload := map[string]any{
		"user_id":    user.ID,
		"booking_id": -1,
		"ip":         "127.0.0.1",
	}
	outboxEvent, err := entity.NewOutboxEvent(enum.ChangePasswordType, payload)
	if err != nil {
		logger.Log.Error("[AuthUseCase] Failed to create outboxEvent", "eventType", enum.ChangePasswordType, "error", err.Error())
	}

	err = a.outboxRepo.CreateEvent(outboxEvent)
	if err != nil {
		logger.Log.Error("[AuthUseCase] Failed to write outboxEvent", "eventType", enum.ChangePasswordType, "error", err.Error())
	}

	return err
}

func (a *AuthUseCase) ChangeEmail(id int, req *entity.ChangeEmailRequest) (*entity.User, error) {
	user, err := a.userUseCase.GetUserByID(id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	exists, err := a.userUseCase.EmailExists(req.NewEmail)
	if exists {
		return nil, domain.ErrEmailTaken
	}
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.ConfirmPassword))
	if err != nil {
		return nil, domain.ErrCurPasswordIsNotCorrect
	}

	if user.IsTotpEnabled {
		if req.OtpCode == "" {
			return nil, domain.ErrTotpSecretNotSet
		}
		if !a.totpService.ValidateCode(user.TotpSecret, req.OtpCode) {
			return nil, domain.ErrInvalidTotpCode
		}
	}

	updates := map[string]any{
		"email": req.NewEmail,
	}
	updatedUser, err := a.userUseCase.UpdateUserData(user.ID, updates)

	payload := map[string]any{
		"user_id":    user.ID,
		"booking_id": -1,
		"ip":         "127.0.0.1",
	}
	outboxEvent, err := entity.NewOutboxEvent(enum.ChangeEmailType, payload)
	if err != nil {
		logger.Log.Error("[AuthUseCase] Failed to create outboxEvent", "eventType", enum.ChangeEmailType, "error", err.Error())
	}

	err = a.outboxRepo.CreateEvent(outboxEvent)
	if err != nil {
		logger.Log.Error("[AuthUseCase] Failed to write outboxEvent", "eventType", enum.ChangeEmailType, "error", err.Error())
	}

	return updatedUser, err
}

func (a *AuthUseCase) SetupTotp(id int) (string, error) {
	user, err := a.userUseCase.GetUserByID(id)
	if err != nil {
		return "", domain.ErrUserNotFound
	}
	if user.IsTotpEnabled {
		return "", domain.ErrTotpAlreadyEnabled
	}

	secret, otpUrl, err := a.totpService.GenerateSecret(user.Email)
	if err != nil {
		return "", err
	}
	updates := map[string]any{
		"totp_secret": secret,
	}

	_, err = a.userUseCase.UpdateUserData(user.ID, updates)
	if err != nil {
		return "", err
	}

	return otpUrl, nil
}

func (a *AuthUseCase) VerifyTotp(id int, req *entity.VerifyTotpRequest) error {
	user, err := a.userUseCase.GetUserByID(id)
	if err != nil {
		return domain.ErrUserNotFound
	}
	if user.IsTotpEnabled {
		return domain.ErrTotpAlreadyEnabled
	}
	if user.TotpSecret == "" {
		return domain.ErrTotpSecretNotSet
	}

	if !a.totpService.ValidateCode(user.TotpSecret, req.OtpCode) {
		return domain.ErrInvalidTotpCode
	}

	updates := map[string]any{
		"is_totp_enabled": true,
	}
	_, err = a.userUseCase.UpdateUserData(user.ID, updates)
	if err != nil {
		return err
	}

	return nil
}

func (a *AuthUseCase) DisableTotp(id int, req *entity.DisableTotpRequest) error {
	user, err := a.userUseCase.GetUserByID(id)
	if err != nil {
		return domain.ErrUserNotFound
	}
	if !user.IsTotpEnabled {
		return domain.ErrTotpAlreadyDisabled
	}
	if user.TotpSecret == "" {
		return domain.ErrTotpSecretNotSet
	}

	if !a.totpService.ValidateCode(user.TotpSecret, req.OtpCode) {
		return domain.ErrInvalidTotpCode
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword))
	if err != nil {
		return domain.ErrCurPasswordIsNotCorrect
	}

	updates := map[string]any{
		"is_totp_enabled": false,
	}
	_, err = a.userUseCase.UpdateUserData(user.ID, updates)
	if err != nil {
		return err
	}

	return nil
}
