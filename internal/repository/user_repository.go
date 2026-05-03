package repository

import (
	"BookingGo/internal/entity"
	"BookingGo/pkg/db"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Репозиторий, в будущем можно добавить логгер, кэширование...
type UserRepository struct{}

// Конструктор
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) GetAll() ([]entity.User, error) {
	var users []entity.User
	result := db.DB.Find(&users)
	return users, result.Error
}

func (r *UserRepository) GetById(id int) (*entity.User, error) {
	var user entity.User
	result := db.DB.First(&user, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*entity.User, error) {
	var user entity.User
	result := db.DB.Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *UserRepository) Create(user *entity.User) error {
	result := db.DB.Create(user)
	return result.Error
}

func (r *UserRepository) Update(id int, requestUser *entity.UpdateUserRequest) (*entity.User, error) {
	_, err := r.GetById(id)
	if err != nil {
		log.Println("Ошибка при попытке обновить данные пользователя, пользователь не найден")
		return nil, err
	}

	updates := map[string]interface{}{}
	if requestUser.Email != nil {
		exists, err := r.EmailExists(*requestUser.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrEmailTaken
		}
		updates["email"] = *requestUser.Email
	}
	if requestUser.Password != nil {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(*requestUser.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		updates["password_hash"] = string(passwordHash)
	}
	if requestUser.FIO != nil {
		updates["fio"] = *requestUser.FIO
	}
	if requestUser.Phone != nil {
		updates["phone"] = *requestUser.Phone
	}

	result := db.DB.Model(&entity.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrUserNotFound
	}

	return r.GetById(id)
}

func (r *UserRepository) Delete(id int) error {
	err := db.DB.Delete(&entity.User{}, id)
	if err.Error != nil {
		return err.Error
	}
	if err.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int64
	err := db.DB.Model(&entity.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
