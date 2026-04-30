package repositories

import (
	"BookingGo/internal/entity"
	"BookingGo/pkg/db"
	"context"
	"database/sql"
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

func (r *UserRepository) GetById(id int) (*entity.User, error) {
	query := `
		SELECT id, email, password, fio, phone, role, created_at
		FROM users
		WHERE id = $1`

	var user entity.User
	// Получаем строку по sql запросу и копируем данные из нее
	err := db.DB.QueryRow(context.Background(), query, id).Scan(&user.ID, &user.Email, &user.Password, &user.FIO, &user.Phone, &user.Role, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		log.Printf("Ошибка при получении данных пользователя с id: %d  Error:%v", id, err)
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(user *entity.User) error {
	query := `
		INSERT INTO users (email, password, fio, phone, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := db.DB.QueryRow(context.Background(), query,
		user.Email,
		user.Password,
		user.FIO,
		user.Phone,
		user.Role,
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		log.Printf("Ошибка при добавлении пользователя в БД: %v", err)
	}
	return err
}
