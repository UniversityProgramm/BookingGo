package controller

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
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

// GetAllUsers godoc
// @Summary      Получить всех пользователей
// @Description  Возвращает список всех пользователей. Доступно staff и admin.
// @Tags         management
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   entity.User              "Список пользователей"
// @Failure      401  {object}  map[string]string        "Требуется авторизация"
// @Failure      403  {object}  map[string]string        "Доступ запрещён"
// @Failure      500  {object}  map[string]string        "Внутренняя ошибка сервера"
// @Router       /staffPanel/users [get]
// @Router       /adminPanel/users [get]
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

// GetUserByID godoc
// @Summary      Получить пользователя по ID
// @Description  Возвращает данные пользователя по его ID
// @Tags         management
// @Security     BearerAuth
// @Param        id  path      int  true  "ID пользователя"
// @Success      200  {object}  entity.User              "Данные пользователя"
// @Failure      400  {object}  map[string]string        "ID должен быть числом"
// @Failure      401  {object}  map[string]string        "Требуется авторизация"
// @Failure      403  {object}  map[string]string        "Доступ запрещён"
// @Failure      404  {object}  map[string]string        "Пользователь не найден"
// @Failure      500  {object}  map[string]string        "Внутренняя ошибка сервера"
// @Router       /staffPanel/users/{id} [get]
// @Router       /adminPanel/users/{id} [get]
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

// GetUserByEmail godoc
// @Summary      Найти пользователя по email
// @Description  Возвращает данные пользователя по email
// @Tags         management
// @Security     BearerAuth
// @Param        email  path      string  true  "Email пользователя"
// @Success      200  {object}  entity.User              "Данные пользователя"
// @Failure      401  {object}  map[string]string        "Требуется авторизация"
// @Failure      403  {object}  map[string]string        "Доступ запрещён"
// @Failure      404  {object}  map[string]string        "Пользователь не найден"
// @Failure      500  {object}  map[string]string        "Внутренняя ошибка сервера"
// @Router       /staffPanel/users/email/{email} [get]
// @Router       /adminPanel/users/email/{email} [get]
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

// CreateUser godoc
// @Summary      Создать пользователя (admin)
// @Description  Создаёт нового пользователя с указанной ролью. Только для администраторов.
// @Tags         management
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      entity.CreateUserRequest  true  "Данные пользователя"
// @Success      201      {object}  entity.User               "Созданный пользователь"
// @Failure      400      {object}  map[string]string         "Неверный формат"
// @Failure      401      {object}  map[string]string         "Требуется авторизация"
// @Failure      403      {object}  map[string]string         "Доступ запрещён"
// @Failure      500      {object}  map[string]string         "Внутренняя ошибка сервера"
// @Router       /adminPanel/users [post]
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

// UpdateUser godoc
// @Summary      Обновить пользователя (admin)
// @Description  Обновляет данные пользователя по ID. Только для администраторов.
// @Tags         management
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int                          true  "ID пользователя"
// @Param        request  body      entity.UpdateUserProfileRequest  true  "Данные для обновления"
// @Success      200  {object}  entity.User              "Обновлённый пользователь"
// @Failure      400  {object}  map[string]string        "Неверный формат / ID должен быть числом"
// @Failure      401  {object}  map[string]string        "Требуется авторизация"
// @Failure      403  {object}  map[string]string        "Доступ запрещён"
// @Failure      404  {object}  map[string]string        "Пользователь не найден"
// @Failure      409  {object}  map[string]string        "Email уже занят"
// @Failure      500  {object}  map[string]string        "Внутренняя ошибка сервера"
// @Router       /adminPanel/users/{id} [put]
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

// DeleteUser godoc
// @Summary      Удалить пользователя (admin)
// @Description  Удаляет пользователя по ID. Только для администраторов.
// @Tags         management
// @Security     BearerAuth
// @Param        id  path      int  true  "ID пользователя"
// @Success      204  {object}  nil                    "Пользователь удалён"
// @Failure      400  {object}  map[string]string      "ID должен быть числом"
// @Failure      401  {object}  map[string]string      "Требуется авторизация"
// @Failure      403  {object}  map[string]string      "Доступ запрещён"
// @Failure      404  {object}  map[string]string      "Пользователь не найден"
// @Failure      500  {object}  map[string]string      "Внутренняя ошибка сервера"
// @Router       /adminPanel/users/{id} [delete]
func (u UserController) DeleteUser(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
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
