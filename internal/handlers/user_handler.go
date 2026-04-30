package handlers

import (
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/internal/repositories"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var userRep = repositories.NewUserRepository()

// Методы для эндпоинтов

func GetAllUsers(c *gin.Context) {
	users, err := userRep.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось получить пользователей",
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

func GetUserByID(c *gin.Context) {
	userId := c.Param("id")
	userIdInt, err := strconv.Atoi(userId)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID должен быть числом",
		})
		return
	}

	user, err := userRep.GetById(userIdInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprint("Не удалось получить пользователя с id:", userId),
		})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Пользователь с ID %d не найден", userIdInt),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

func CreateUser(c *gin.Context) {
	var createRequest entity.CreateUserRequest

	if err := c.ShouldBindJSON(&createRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат данных",
		})
		return
	}

	user := &entity.User{
		Email:    createRequest.Email,
		Password: createRequest.Password,
		FIO:      createRequest.FIO,
		Phone:    createRequest.Phone,
		Role:     enum.RoleAdmin,
	}

	if err := userRep.Create(user); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") ||
			strings.Contains(err.Error(), "Duplicate key") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Пользователь с таким email уже существует",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось создать пользователя",
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func UpdateUser(c *gin.Context) {
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
			"error": "Неверный формат данных",
		})
		return
	}

	updatedUser, err := userRep.Update(userIdInt, &updateRequest)
	if err != nil {
		if err.Error() == "email is taken" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Email %s занят", updateRequest.Email),
			})
		} else if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("Пользователь с ID:%d не найден", userIdInt),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления пользователя"})
		}
		return
	}
	c.JSON(http.StatusOK, updatedUser)
}

// Сделать проверку роли с помощью JWT(только RoleAdmin может удалять)
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	userIDInt, err := strconv.Atoi(userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID должен быть числом",
		})
		return
	}

	err = userRep.Delete(userIDInt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("Пользователь с ID:%d не существует", userIDInt),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Ошибка при удалении пользователя с ID:%d", userIDInt),
			})
		}
		return
	}
	c.Status(http.StatusNoContent)
}
