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

// Login godoc
// @Summary      Вход в систему
// @Description  Аутентификация пользователя по email и паролю. Возвращает JWT токен для дальнейших запросов.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      entity.LoginUserRequest  true  "Данные для входа"
// @Success      200      {object}  map[string]string        "JWT токен"
// @Failure      400      {object}  map[string]string        "Неверный формат данных"
// @Failure      401      {object}  map[string]string        "Неверный email или пароль"
// @Failure      500      {object}  map[string]string        "Внутренняя ошибка сервера"
// @Router       /auth/login [post]
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

// Register godoc
// @Summary      Регистрация нового пользователя
// @Description  Создаёт нового пользователя с ролью client. Возвращает JWT токен.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      entity.CreateUserRequest  true  "Данные пользователя"
// @Success      201      {object}  map[string]string         "JWT токен"
// @Failure      400      {object}  map[string]string         "Неверный формат или email занят"
// @Failure      500      {object}  map[string]string         "Внутренняя ошибка сервера"
// @Router       /auth/register [post]
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

// Logout godoc
// @Summary      Выход из системы
// @Description  Отзыв текущего JWT токена (добавление в blacklist). Требует авторизации.
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string  "Сообщение об успешном выходе"
// @Failure      401  {object}  map[string]string  "Требуется авторизация"
// @Failure      500  {object}  map[string]string  "Ошибка при выходе"
// @Router       /auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	err := ac.authUseCase.Logout(currentUser.UserID, currentUser.ID, currentUser.ExpiresAt.Time)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при выходе"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен успешно"})
}

// GetMe godoc
// @Summary      Получить текущий профиль
// @Description  Возвращает данные авторизованного пользователя
// @Tags         profile
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  entity.User              "Данные пользователя"
// @Failure      401  {object}  map[string]string        "Требуется авторизация"
// @Failure      404  {object}  map[string]string        "Пользователь не найден"
// @Failure      500  {object}  map[string]string        "Внутренняя ошибка сервера"
// @Router       /me/profile [get]
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

// UpdateMe godoc
// @Summary      Обновить профиль
// @Description  Обновляет ФИО и телефон пользователя
// @Tags         profile
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      entity.UpdateUserProfileRequest  true  "Данные для обновления"
// @Success      200      {object}  entity.User                      "Обновлённый пользователь"
// @Failure      400      {object}  map[string]string                "Неверный формат"
// @Failure      401      {object}  map[string]string                "Требуется авторизация"
// @Failure      404      {object}  map[string]string                "Пользователь не найден"
// @Failure      409      {object}  map[string]string                "Email уже занят"
// @Failure      500      {object}  map[string]string                "Внутренняя ошибка сервера"
// @Router       /me/profile [put]
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

// ChangePassword godoc
// @Summary      Сменить пароль
// @Description  Изменяет пароль пользователя. Требует подтверждения текущего пароля и кода TOTP (если 2FA включён).
// @Tags         profile
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      entity.ChangePasswordRequest  true  "Текущий и новый пароль"
// @Success      200  {object}  map[string]string  "Пароль успешно изменён"
// @Failure      400  {object}  map[string]string  "Неверный текущий пароль / новый пароль совпадает / неверный TOTP"
// @Failure      401  {object}  map[string]string  "Требуется авторизация"
// @Failure      404  {object}  map[string]string  "Пользователь не найден"
// @Failure      500  {object}  map[string]string  "Внутренняя ошибка сервера"
// @Router       /me/profile/password [put]
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

// ChangeEmail godoc
// @Summary      Сменить email
// @Description  Изменяет email пользователя. Требует подтверждения пароля и кода TOTP (если 2FA включён).
// @Tags         profile
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      entity.ChangeEmailRequest  true  "Новый email и подтверждение"
// @Success      200  {object}  map[string]interface{}  "Пользователь с новым email"
// @Failure      400  {object}  map[string]string       "Неверный пароль / неверный TOTP"
// @Failure      401  {object}  map[string]string       "Требуется авторизация"
// @Failure      404  {object}  map[string]string       "Пользователь не найден"
// @Failure      409  {object}  map[string]string       "Email уже занят"
// @Failure      500  {object}  map[string]string       "Внутренняя ошибка сервера"
// @Router       /me/profile/email [put]
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

// SetupTotp godoc
// @Summary      Настроить 2FA
// @Description  Генерирует секрет TOTP и возвращает otpauth URL для сканирования в приложении-аутентификаторе (Google Authenticator, Authy и т.д.)
// @Tags         totp
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string  "OTP URL для настройки"
// @Failure      401  {object}  map[string]string  "Требуется авторизация"
// @Failure      404  {object}  map[string]string  "Пользователь не найден"
// @Failure      409  {object}  map[string]string  "2FA уже включён"
// @Failure      500  {object}  map[string]string  "Внутренняя ошибка сервера"
// @Router       /me/otp/setup [post]
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

// VerifyTotp godoc
// @Summary      Подтвердить 2FA
// @Description  Подтверждает код TOTP и включает двухфакторную аутентификацию
// @Tags         totp
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      entity.VerifyTotpRequest  true  "Код TOTP"
// @Success      200  {object}  map[string]string  "2FA успешно включён"
// @Failure      400  {object}  map[string]string  "Неверный код"
// @Failure      401  {object}  map[string]string  "Требуется авторизация"
// @Failure      404  {object}  map[string]string  "Пользователь не найден"
// @Failure      409  {object}  map[string]string  "2FA уже подтвержден и включен"
// @Failure      412  {object}  map[string]string  "Сначала настройте 2FA"
// @Failure      500  {object}  map[string]string  "Внутренняя ошибка сервера"
// @Router       /me/otp/verify [post]
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

// DisableTotp godoc
// @Summary      Отключить 2FA
// @Description  Отключает двухфакторную аутентификацию. Требует код TOTP и текущий пароль.
// @Tags         totp
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      entity.DisableTotpRequest  true  "Код TOTP и пароль"
// @Success      200  {object}  map[string]string  "2FA отключён"
// @Failure      400  {object}  map[string]string  "Неверный код / пароль"
// @Failure      401  {object}  map[string]string  "Требуется авторизация"
// @Failure      404  {object}  map[string]string  "Пользователь не найден"
// @Failure      409  {object}  map[string]string  "2FA уже отключён"
// @Failure      412  {object}  map[string]string  "Сначала настройте 2FA"
// @Failure      500  {object}  map[string]string  "Внутренняя ошибка сервера"
// @Router       /me/otp/disable [post]
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
