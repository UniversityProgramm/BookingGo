package repository

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"context"
	"errors"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(DB *gorm.DB) *UserRepository {
	return &UserRepository{db: DB}
}

func (r *UserRepository) GetAll() ([]entity.User, error) {
	var users []entity.User
	result := r.db.Find(&users)
	return users, result.Error
}

func (r *UserRepository) GetByID(id int) (*entity.User, error) {
	return r.GetByIDContext(context.Background(), id)
}

func (r *UserRepository) GetByIDContext(ctx context.Context, id int) (*entity.User, error) {
	var user entity.User
	result := r.db.WithContext(ctx).First(&user, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*entity.User, error) {
	return r.GetByEmailContext(context.Background(), email)
}

func (r *UserRepository) GetByEmailContext(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *UserRepository) Create(user *entity.User) error {
	result := r.db.Create(user)
	return result.Error
}

func (r *UserRepository) Update(id int, requestUser map[string]any) (*entity.User, error) {
	result := r.db.Model(&entity.User{}).Where("id = ?", id).Updates(requestUser)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrUserNotFound
	}

	return r.GetByID(id)
}

func (r *UserRepository) Delete(id int) error {
	err := r.db.Delete(&entity.User{}, id)
	if err.Error != nil {
		return err.Error
	}
	if err.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&entity.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
