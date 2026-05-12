package usecase

type NotificationRepository interface {
}

type NotificationUseCase struct {
	notificationRepo NotificationRepository
}

func NewNotificationUseCase(notificationRepo NotificationRepository) *NotificationUseCase {
	return &NotificationUseCase{notificationRepo: notificationRepo}
}
