package repositories

import (
	"BookingGo/internal/entity"
	"BookingGo/pkg/db"
	"context"
	"log"
)

// Репозиторий, в будущем можно добавить логгер, кэширование...
type UserRepository struct{}

// Конструктор
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) GetAll() ([]entity.User, error) {
	query := `
		SELECT id, email, password, fio, phone, role, created_at
		FROM users
		ORDER BY id`
	rows, err := db.DB.Query(context.Background(), query)
	if err != nil {
		log.Printf("Ошибка при выполнении SQL запроса: %v", err)
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	// Итерация по строкам
	for rows.Next() {
		var u entity.User
		// Копируем данный из sql строки
		err := rows.Scan(&u.ID, &u.Email, &u.Password, &u.FIO, &u.Phone, &u.Role, &u.CreatedAt)
		if err != nil {
			log.Printf("Ошибка при сканировании строки: %v", err)
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка в результатах запроса: %v", err)
		return nil, err
	}
	return users, nil
}
