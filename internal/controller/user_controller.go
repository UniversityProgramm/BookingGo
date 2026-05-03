package controller

import (
	"BookingGo/internal/entity"
	"BookingGo/internal/usecase"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userUsecase *usecase.UserUsecase
}

func NewUserController(userUsecase *usecase.UserUsecase) *UserController {
	return &UserController{userUsecase: userUsecase}
}

func (u UserController) GetAllUsers(c *gin.Context) {
	users, err := u.userUsecase.GetAllUsers()
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

	user, err := u.userUsecase.GetUserByID(userId)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
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
	user, err := u.userUsecase.GetUserByEmail(c.Param("email"))
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
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

	user, err := u.userUsecase.CreateUser(&createRequest)
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

	var updateRequest entity.UpdateUserRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprint("Неверный формат данных", err.Error()),
		})
		return
	}

	updatedUser, err := u.userUsecase.UpdateUser(userId, &updateRequest)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailTaken) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Этот Email занят",
			})
		} else if errors.Is(err, usecase.ErrUserNotFound) {
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

// Сделать проверку роли с помощью JWT(только RoleAdmin может удалять)
func (u UserController) DeleteUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID должен быть числом",
		})
		return
	}

	err = u.userUsecase.DeleteUser(userID)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
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
