package usecase

import (
	"BookingGo/internal/cache"
	"BookingGo/internal/customTemplates"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/pkg/logger"
	"context"
	"fmt"
	"strings"
	"time"
)

//go:generate mockgen -source=notification_usecase.go -destination=mocks/notification_mocks.go -package=mocks
type NotificationRepository interface {
	UpdateSettings(userID int, req *entity.NotificationSettings) error
	GetAllNotificationsByUserID(userID int) ([]entity.Notification, error)
	MarkAllAsRead(recipientID int) error
	CreateNotif(notification *entity.Notification) error
}

type NotificationUseCase struct {
	notificationRepo NotificationRepository
	bookingRepo      BookingRepository
	userRepo         UserRepository
	cache            cache.Cache
}

func NewNotificationUseCase(notificationRepo NotificationRepository, bookingRepo BookingRepository, userRepo UserRepository) *NotificationUseCase {
	return &NotificationUseCase{notificationRepo: notificationRepo, bookingRepo: bookingRepo, userRepo: userRepo}
}

func (n *NotificationUseCase) invalidateUserCache(ctx context.Context, userID int) {
	prefix := fmt.Sprintf("user:%d", userID)
	err := n.cache.Delete(ctx, prefix)
	if err != nil {
		logger.Log.Error("[AuthUseCase] Failed to invalidate user cache", "error", err.Error())
	}
}

func (n *NotificationUseCase) CreateNotification(userID int, notificationType enum.TypeOfNotification, params entity.NotificationParams) error {
	user, err := n.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	var booking *entity.Booking
	if params.BookingID > 0 {
		booking, err = n.bookingRepo.GetBookingByID(params.BookingID)
		if err != nil {
			booking = nil
		}
	}

	text, title := buildTextAndTitle(notificationType, user, booking, params)
	newNotification := &entity.Notification{
		RecipientID: userID,
		Title:       title,
		Body:        text,
		IsRead:      false,
	}
	err = n.notificationRepo.CreateNotif(newNotification)
	if err != nil {
		return err
	}

	if user.UserNotificationSettings.IsEmailSend {
		if err := sendToEmail(user.Email, title, text); err != nil {
			return err
		}
	}
	if user.UserNotificationSettings.IsPhoneSend {
		if err := sendToPhone(user.Phone, title, text); err != nil {
			return err
		}
	}

	return nil
}

func buildTextAndTitle(notificationType enum.TypeOfNotification, user *entity.User, booking *entity.Booking, params entity.NotificationParams) (string, string) {
	bookingDate := ""
	message := ""
	if booking != nil {
		bookingDate = booking.SlotStart.Format("02.01.2006 15:04")
		message = strings.TrimSpace(booking.ProblemDescription)
	}

	tmplData := map[string]interface{}{
		"UserName":    strings.TrimSpace(user.FIO),
		"BookingDate": bookingDate,
		"Message":     message,
		"IP":          params.IP,
		"DateNow":     time.Now().Format("02.01.2006 15:04"),
	}

	switch notificationType {
	case enum.NewBookingType:
		return renderNotificationTemplate("booking_created", tmplData), "Вы совершили запись"
	case enum.CancelBookingType:
		return renderNotificationTemplate("booking_canceled", tmplData), "Вы отменили запись"
	case enum.AuthType:
		return renderNotificationTemplate("auth", tmplData), "Вы вошли в аккаунт"
	case enum.CompleteBookingType:
		return renderNotificationTemplate("booking_completed", tmplData), "Статус вашей записи был изменен"
	case enum.ChangeEmailType:
		return renderNotificationTemplate("change_email", tmplData), "Вы сменили email"
	case enum.ChangePasswordType:
		return renderNotificationTemplate("change_password", tmplData), "Вы сменили пароль"
	default:
		logger.Log.Error("[NotificationUseCase] Unknown notification type", "notificationType", notificationType)
		return "Неизвестный тип уведомления", "Получено новое уведомление"
	}

}

func renderNotificationTemplate(name string, data map[string]interface{}) string {
	tmpl, ok := customTemplates.NotificationTemplates[name]
	if !ok {
		return "Уведомления с таким ключом шаблона не существует"
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("Ошибка генерации текста: %v", err)
	}

	return strings.TrimSpace(buf.String())
}

func sendToEmail(email string, title string, text string) error {
	logger.Log.Info("[NotificationUseCase] Sent email")
	return nil
}

func sendToPhone(phone string, title string, text string) error {
	logger.Log.Info("[NotificationUseCase] Sent SMS")
	return nil
}

func (n *NotificationUseCase) GetMyNotifications(userID int) ([]entity.Notification, error) {
	return n.notificationRepo.GetAllNotificationsByUserID(userID)
}

func (n *NotificationUseCase) UpdateNotificationSettings(userID int, req *entity.NotificationSettings) error {
	err := n.notificationRepo.UpdateSettings(userID, req)
	if err != nil {
		return err
	}

	n.invalidateUserCache(context.Background(), userID)

	return nil
}

func (n *NotificationUseCase) MarkAllAsRead(userID int) error {
	return n.notificationRepo.MarkAllAsRead(userID)
}
