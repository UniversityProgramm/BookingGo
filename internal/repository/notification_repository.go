package repository

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"

	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(DB *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: DB}
}

func (r *NotificationRepository) CreateNotif(notification *entity.Notification) error {
	return r.db.Create(notification).Error
}

func (r *NotificationRepository) GetAllNotificationsByUserID(userID int) ([]entity.Notification, error) {
	var notifications []entity.Notification
	err := r.db.Where("recipient_id = ?", userID).Find(&notifications).Error
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepository) UpdateSettings(userID int, req *entity.NotificationSettings) error {
	result := r.db.Model(&entity.User{}).Where("id = ?", userID).Update("notification_settings", req)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *NotificationRepository) MarkAllAsRead(recipientID int) error {
	result := r.db.Model(&entity.Notification{}).
		Where("recipient_id = ? AND is_read = ?", recipientID, false).
		Update("is_read", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotificationsAlreadyRead
	}

	return nil
}
