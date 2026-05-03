package repository

import (
	"BookingGo/internal/entity"
	"BookingGo/pkg/db"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrEmailTaken   = errors.New("user not found")
	ErrUserNotFound = errors.New("email is taken")
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) GetAll() ([]entity.User, error) {
	var users []entity.User
	result := db.DB.Find(&users)
	return users, result.Error
}

func (r *UserRepository) GetByID(id int) (*entity.User, error) {
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

func (r *UserRepository) Update(id int, requestUser map[string]interface{}) (*entity.User, error) {
	result := db.DB.Model(&entity.User{}).Where("id = ?", id).Updates(requestUser)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrUserNotFound
	}

	return r.GetByID(id)
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
