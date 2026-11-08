package postgres

import (
	"BookingGo/internal/domain"

	"gorm.io/gorm"
)

type NotificationPostgres struct {
	db *gorm.DB
}

func NewNotificationPostgres(DB *gorm.DB) *NotificationPostgres {
	return &NotificationPostgres{db: DB}
}

func (r *NotificationPostgres) CreateNotif(notification *domain.Notification) error {
	return r.db.Create(notification).Error
}

func (r *NotificationPostgres) GetAllNotificationsByUserID(userID int) ([]domain.Notification, error) {
	var notifications []domain.Notification
	err := r.db.Where("recipient_id = ?", userID).Order("created_at ASC").Find(&notifications).Error
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationPostgres) UpdateSettings(userID int, req *domain.NotificationSettings) error {
	result := r.db.Model(&domain.User{}).Where("id = ?", userID).Update("notification_settings", req)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *NotificationPostgres) MarkAllAsRead(recipientID int) error {
	result := r.db.Model(&domain.Notification{}).
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
