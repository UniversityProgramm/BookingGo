package controller

import (
	"BookingGo/internal/usecase"
)

type NotificationController struct {
	notificationUseCase *usecase.NotificationUseCase
}

func NewNotificationController(notificationUseCase *usecase.NotificationUseCase) *NotificationController {
	return &NotificationController{notificationUseCase: notificationUseCase}
}
