package controller

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/middleware"
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

func (nc *NotificationController) GetMyNotifications(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
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
			}

			c.JSON(http.StatusOK, gin.H{
				"notifications": notifications,
				"warning":       "Уведомления получены, но не удалось пометить их как прочитанные"},
			)

		}
	}

	c.JSON(http.StatusOK, notifications)
}

func (nc *NotificationController) UpdateNotificationSettings(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
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
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка изменения настроек уведомлений"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Настройки уведомлений успешно обновлены",
		"settings": settingsReq,
	})
}
