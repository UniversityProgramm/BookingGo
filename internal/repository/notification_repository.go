package repository

import "gorm.io/gorm"

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(DB *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: DB}
}
