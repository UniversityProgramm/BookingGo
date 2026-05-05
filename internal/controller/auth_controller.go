package controller

import (
	"BookingGo/internal/entity"
	"BookingGo/internal/middleware"
	"BookingGo/internal/usecase"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authUsecase *usecase.AuthUsecase
}

func NewAuthController(authUsec *usecase.AuthUsecase) *AuthController {
	return &AuthController{authUsecase: authUsec}
}

func (ac *AuthController) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8,max=20"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	token, err := ac.authUsecase.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidEmail) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email"})
		} else if errors.Is(err, usecase.ErrInvalidPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный пароль"})
		} else {
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
	}

	token, err := ac.authUsecase.Register(&req)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailTaken) {
			c.JSON(http.StatusOK, gin.H{"error": "Этот Email занят"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка регистрации"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"token": token})
}

func (ac *AuthController) GetMe(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	user, err := ac.authUsecase.GetUserByID(currentUser.UserID)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки профиля"})
	}

	c.JSON(http.StatusOK, user)
}

func (ac *AuthController) UpdateMe(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	userID := currentUser.UserID
	var req entity.UpdateUserRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	updatedUser, err := ac.authUsecase.UpdateUser(userID, &req)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		} else if errors.Is(err, usecase.ErrEmailTaken) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email уже занят"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления данных профиля"})
		}
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

func (ac *AuthController) ChangePassword(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
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

	err := ac.authUsecase.ChangePassword(userID, &req)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		} else if errors.Is(err, usecase.ErrCurPasswordIsNotCorrect) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Введен неверный текущий пароль"})
		} else if errors.Is(err, usecase.ErrSamePassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Новый пароль совпадает с текущим"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка смены пароля"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пароль успешно изменен"})
}
