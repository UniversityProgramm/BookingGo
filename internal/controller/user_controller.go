package controller

import (
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/internal/repository"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	userRepository repository.UserRepository
}

func (u UserController) GetAllUsers(c *gin.Context) {
	users, err := u.userRepository.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось получить пользователей",
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (u UserController) GetUserByID(c *gin.Context) {
	userId := c.Param("id")
	userIdInt, err := strconv.Atoi(userId)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID должен быть числом",
		})
		return
	}

	user, err := u.userRepository.GetById(userIdInt)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Пользователь с таким ID не найден",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось получить пользователя",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (u UserController) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	user, err := u.userRepository.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Пользователь с таким Email не найден",
			})
			return
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

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(createRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось хешировать пароль",
		})
		return
	}

	user := &entity.User{
		Email:        createRequest.Email,
		PasswordHash: string(passwordHash),
		FIO:          createRequest.FIO,
		Phone:        createRequest.Phone,
		Role:         enum.RoleClient,
	}

	if err := u.userRepository.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось создать пользователя",
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (u UserController) UpdateUser(c *gin.Context) {
	userId := c.Param("id")
	userIdInt, err := strconv.Atoi(userId)

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

	updatedUser, err := u.userRepository.Update(userIdInt, &updateRequest)
	if err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Этот Email занят",
			})
		} else if errors.Is(err, repository.ErrUserNotFound) {
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
	userID := c.Param("id")
	userIDInt, err := strconv.Atoi(userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID должен быть числом",
		})
		return
	}

	err = u.userRepository.Delete(userIDInt)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
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
