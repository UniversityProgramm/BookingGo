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

type NotificationController struct {
	notificationUseCase *usecase.NotificationUseCase
}

func NewNotificationController(notificationUseCase *usecase.NotificationUseCase) *NotificationController {
	return &NotificationController{notificationUseCase: notificationUseCase}
}

// GetMyNotifications godoc
// @Summary      Получить уведомления
// @Description  Возвращает список уведомлений текущего пользователя и помечает их как прочитанные
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   entity.Notification      "Список уведомлений"
// @Failure      401  {object}  map[string]string        "Требуется авторизация"
// @Failure      500  {object}  map[string]string        "Внутренняя ошибка сервера"
// @Router       /me/notifications [get]
func (nc *NotificationController) GetMyNotifications(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	notifications, err := nc.notificationUseCase.GetMyNotifications(currentUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки уведомлений"})
		return
	}

	if len(notifications) > 0 {
		err = nc.notificationUseCase.MarkAllAsRead(currentUser.UserID)
		if err != nil {
			if errors.Is(err, domain.ErrNotificationsAlreadyRead) {
				c.JSON(http.StatusOK, gin.H{
					"notifications": notifications,
					"warning":       "Все уведомления уже прочитаны"},
				)
			} else {
				c.JSON(http.StatusOK, gin.H{
					"notifications": notifications,
					"warning":       "Уведомления получены, но не удалось пометить их как прочитанные"},
				)
			}
			return
		}
	}

	c.JSON(http.StatusOK, notifications)
}

// UpdateNotificationSettings godoc
// @Summary      Обновить настройки уведомлений
// @Description  Включает/отключает email и SMS уведомления
// @Tags         notifications
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      entity.NotificationSettings  true  "Настройки уведомлений"
// @Success      200  {object}  map[string]interface{}  "Настройки обновлены"
// @Failure      400  {object}  map[string]string       "Неверный формат"
// @Failure      401  {object}  map[string]string       "Требуется авторизация"
// @Failure      404  {object}  map[string]string       "Пользователь не найден"
// @Failure      500  {object}  map[string]string       "Внутренняя ошибка сервера"
// @Router       /me/notifications/settings [patch]
func (nc *NotificationController) UpdateNotificationSettings(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	var settingsReq entity.NotificationSettings
	if err := c.ShouldBind(&settingsReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	err := nc.notificationUseCase.UpdateNotificationSettings(currentUser.UserID, &settingsReq)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка изменения настроек уведомлений"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Настройки уведомлений успешно обновлены",
		"settings": settingsReq,
	})
}
