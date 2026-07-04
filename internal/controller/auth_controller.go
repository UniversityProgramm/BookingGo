package controller

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/usecase"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authUseCase *usecase.AuthUseCase
}

func NewAuthController(authUseCase *usecase.AuthUseCase) *AuthController {
	return &AuthController{authUseCase: authUseCase}
}

func (ac *AuthController) Login(c *gin.Context) {
	var req entity.LoginUserRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	token, err := ac.authUseCase.Login(&req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidEmail):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email"})
		case errors.Is(err, domain.ErrInvalidPassword):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный пароль"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка входа"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (ac *AuthController) Register(c *gin.Context) {
	var req entity.CreateUserRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	token, err := ac.authUseCase.Register(&req)
	if err != nil {
		if errors.Is(err, domain.ErrEmailTaken) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Этот Email занят"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка регистрации"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"token": token})
}

func (ac *AuthController) GetMe(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	user, err := ac.authUseCase.GetMe(currentUser.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки профиля"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

func (ac *AuthController) UpdateMe(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	userID := currentUser.UserID
	var req entity.UpdateUserProfileRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	updatedUser, err := ac.authUseCase.UpdateMe(userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		} else if errors.Is(err, domain.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "Email уже занят"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления данных профиля"})
		}
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

func (ac *AuthController) ChangePassword(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	userID := currentUser.UserID
	var req entity.ChangePasswordRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	err := ac.authUseCase.ChangePassword(userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		} else if errors.Is(err, domain.ErrCurPasswordIsNotCorrect) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Введен неверный текущий пароль"})
		} else if errors.Is(err, domain.ErrSamePassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Новый пароль совпадает с текущим"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка смены пароля"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пароль успешно изменен"})
}

func (ac *AuthController) ChangeEmail(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	userID := currentUser.UserID
	var req entity.ChangeEmailRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	user, err := ac.authUseCase.ChangeEmail(userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		} else if errors.Is(err, domain.ErrCurPasswordIsNotCorrect) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Введен неверный текущий пароль"})
		} else if errors.Is(err, domain.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "Данная почта уже занята"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка смены почты"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Почта успешно изменена", "user": user})
}

func (ac *AuthController) SetupTotp(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	otpUrl, err := ac.authUseCase.SetupTotp(currentUser.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		} else if errors.Is(err, domain.ErrTotpAlreadyEnabled) {
			c.JSON(http.StatusConflict, gin.H{"error": "2FA уже включен"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка настройки 2FA"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"manual_entry_key": otpUrl})
}

func (ac *AuthController) VerifyTotp(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	var req entity.VerifyTotpRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	err := ac.authUseCase.VerifyTotp(currentUser.UserID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		} else if errors.Is(err, domain.ErrTotpAlreadyEnabled) {
			c.JSON(http.StatusConflict, gin.H{"error": "2FA уже подтвержден и включен"})
		} else if errors.Is(err, domain.ErrTotpSecretNotSet) {
			c.JSON(http.StatusPreconditionRequired, gin.H{"error": "Сначала настройте 2FA"})
		} else if errors.Is(err, domain.ErrInvalidTotpCode) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный код"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка подтверждения 2FA"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA подтвержден и включен"})
}

func (ac *AuthController) DisableTotp(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	var req entity.DisableTotpRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	err := ac.authUseCase.DisableTotp(currentUser.UserID, &req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		case errors.Is(err, domain.ErrTotpAlreadyDisabled):
			c.JSON(http.StatusConflict, gin.H{"error": "2FA уже выключен"})
		case errors.Is(err, domain.ErrTotpSecretNotSet):
			c.JSON(http.StatusPreconditionRequired, gin.H{"error": "Сначала настройте 2FA"})
		case errors.Is(err, domain.ErrInvalidTotpCode):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный код"})
		case errors.Is(err, domain.ErrCurPasswordIsNotCorrect):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Введен неверный текущий пароль"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отключения 2FA"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA выключен"})
}
