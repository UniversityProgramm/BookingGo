package postgres

import (
	"BookingGo/internal/domain"
	"context"
	"errors"

	"gorm.io/gorm"
)

type UserPostgres struct {
	db *gorm.DB
}

func NewUserPostgres(DB *gorm.DB) *UserPostgres {
	return &UserPostgres{db: DB}
}

func (r *UserPostgres) GetAll() ([]domain.User, error) {
	var users []domain.User
	result := r.db.Find(&users)
	return users, result.Error
}

func (r *UserPostgres) GetByID(id int) (*domain.User, error) {
	return r.GetByIDContext(context.Background(), id)
}

func (r *UserPostgres) GetByIDContext(ctx context.Context, id int) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).First(&user, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *UserPostgres) GetByEmail(email string) (*domain.User, error) {
	return r.GetByEmailContext(context.Background(), email)
}

func (r *UserPostgres) GetByEmailContext(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *UserPostgres) Create(user *domain.User) error {
	result := r.db.Create(user)
	return result.Error
}

func (r *UserPostgres) Update(id int, requestUser map[string]any) (*domain.User, error) {
	result := r.db.Model(&domain.User{}).Where("id = ?", id).Updates(requestUser)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrUserNotFound
	}

	return r.GetByID(id)
}

func (r *UserPostgres) Delete(id int) error {
	err := r.db.Delete(&domain.User{}, id)
	if err.Error != nil {
		return err.Error
	}
	if err.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserPostgres) EmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
