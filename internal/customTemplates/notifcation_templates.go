package customTemplates

import "text/template"

var NotificationTemplates = map[string]*template.Template{
	"booking_created":   template.Must(template.New("booking_created").Parse("{{.UserName}}, вы записались на {{.BookingDate}} с проблемой: {{.Message}}")),
	"booking_completed": template.Must(template.New("booking_completed").Parse("{{.UserName}}, статус вашей записи на {{.BookingDate}} был изменен на «Завершена»")),
	"booking_canceled":  template.Must(template.New("booking_canceled").Parse("Запись на {{.BookingDate}} с проблемой: {{.Message}} отменена.")),
	"booking_auth":      template.Must(template.New("booking_auth").Parse("Вы вошли в аккаунт с IP: {{.IP}}, {{.DateNow}}")),
}
