package repository

import (
	"BookingGo/internal/entity"
	"BookingGo/pkg/db"
	"context"
	"database/sql"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// Репозиторий, в будущем можно добавить логгер, кэширование...
type UserRepository struct{}

// Конструктор
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) GetAll() ([]entity.User, error) {
	query := `
		SELECT id, email, password_hash, fio, phone, role, created_at, updated_at, is_active
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
		// Копируем данные из sql строки
		err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FIO, &u.Phone, &u.Role, &u.CreatedAt, &u.UpdatedAt, &u.IsActive)
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
		SELECT id, email, password_hash, fio, phone, role, created_at, updated_at, is_active
		FROM users
		WHERE id = $1`

	var user entity.User
	// Получаем строку по sql запросу и копируем данные из нее
	err := db.DB.QueryRow(context.Background(), query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FIO,
		&user.Phone, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
	)

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
		INSERT INTO users (email, password_hash, fio, phone, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at, is_active`

	exists, err_ex := r.EmailExists(user.Email)

	if err_ex != nil {
		return err_ex
	}
	if exists {
		return ErrEmailTaken
	}

	err := db.DB.QueryRow(context.Background(), query,
		user.Email,
		user.PasswordHash,
		user.FIO,
		user.Phone,
		user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.IsActive)

	if err != nil {
		log.Printf("Ошибка при добавлении пользователя в БД: %v", err)
		return err
	}
	return nil
}

func (r *UserRepository) Update(id int, requestUser *entity.UpdateUserRequest) (*entity.User, error) {
	current, err := r.GetById(id)

	if err != nil {
		log.Printf("Ошибка при попытке обновить данные пользователя с ID:%d Error:%v", id, err)
		return nil, err
	}

	newEmail := current.Email
	newPassword := current.PasswordHash
	newFIO := current.FIO
	newPhone := current.Phone

	if requestUser.Email != nil {
		exists, err := r.EmailExists(*requestUser.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrEmailTaken
		}
		newEmail = *requestUser.Email
	}
	if requestUser.Password != nil {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(*requestUser.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		newPassword = string(passwordHash)
	}
	if requestUser.FIO != nil {
		newFIO = *requestUser.FIO
	}
	if requestUser.Phone != nil {
		newPhone = *requestUser.Phone
	}

	query := `
		UPDATE users 
		SET email = $1, password_hash = $2, fio = $3, phone = $4
		WHERE id = $5
		RETURNING id, email, password_hash, fio, phone, role, created_at, updated_at, is_active`

	row := db.DB.QueryRow(context.Background(), query,
		newEmail,
		newPassword,
		newFIO,
		newPhone,
		id,
	)

	var updated entity.User

	err = row.Scan(&updated.ID, &updated.Email, &updated.PasswordHash, &updated.FIO,
		&updated.Phone, &updated.Role, &updated.CreatedAt, &updated.UpdatedAt, &updated.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &updated, nil
}

func (r *UserRepository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	commandTag, err := db.DB.Exec(context.Background(), query, id)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := db.DB.QueryRow(context.Background(), query, email).Scan(&exists)
	return exists, err
}
