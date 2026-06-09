package controller

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/internal/middleware"
	"BookingGo/internal/usecase"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userUseCase *usecase.UserUseCase
}

func NewUserController(userUseCase *usecase.UserUseCase) *UserController {
	return &UserController{userUseCase: userUseCase}
}

func (u UserController) GetAllUsers(c *gin.Context) {
	users, err := u.userUseCase.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось получить пользователей",
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (u UserController) GetUserByID(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID должен быть числом",
		})
		return
	}

	user, err := u.userUseCase.GetUserByID(userId)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Пользователь с таким ID не найден",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Не удалось получить пользователя по ID",
			})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

func (u UserController) GetUserByEmail(c *gin.Context) {
	user, err := u.userUseCase.GetUserByEmail(c.Param("email"))
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Пользователь с таким Email не найден",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Не удалось получить пользователя по Email по ID",
			})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

func (u UserController) CreateUser(c *gin.Context) {
	var createRequest entity.CreateUserRequest
	if err := c.ShouldBindJSON(&createRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprint("Неверный формат данных", err.Error()),
		})
		return
	}

	user, err := u.userUseCase.CreateUser(&createRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось создать пользователя",
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (u UserController) UpdateUser(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID должен быть числом",
		})
		return
	}

	var updateRequest entity.UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprint("Неверный формат данных", err.Error()),
		})
		return
	}

	updatedUser, err := u.userUseCase.UpdateUserProfile(userId, &updateRequest)
	if err != nil {
		if errors.Is(err, domain.ErrEmailTaken) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Этот Email занят",
			})
		} else if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Пользователь с таким ID не найден",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления пользователя"})
		}
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

func (u UserController) DeleteUser(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	if currentUser.Role != enum.RoleAdmin {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
		return
	}

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID должен быть числом",
		})
		return
	}

	err = u.userUseCase.DeleteUser(userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Пользователь с таким ID не найден",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка при удалении пользователя",
			})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
