package repository

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/pkg/db"
	"errors"
	"time"

	"gorm.io/gorm"
)

type BookingRepository struct{}

func NewBookingRepository() *BookingRepository {
	return &BookingRepository{}
}

func (r *BookingRepository) Create(booking *entity.Booking) error {
	return db.DB.Create(booking).Error
}

func (r *BookingRepository) IsSlotAvailable(start, end time.Time) (bool, error) {
	var count int64
	err := db.DB.Model(&entity.Booking{}).
		Where("slot_start < ? AND ? < slot_end", end, start).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (r *BookingRepository) GetBookingsByUserID(id int) ([]entity.Booking, error) {
	var bookings []entity.Booking
	err := db.DB.Where("user_id = ?", id).Order("slot_start ASC").Find(&bookings).Error
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *BookingRepository) GetBookingByID(id int) (*entity.Booking, error) {
	var booking entity.Booking
	err := db.DB.First(&booking, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBookingNotFound
		}
		return nil, err
	}

	return &booking, nil
}

func (r *BookingRepository) DeleteBookingByID(bookingID, userID int) error {
	err := db.DB.Where("id = ? AND user_id = ?", bookingID, userID).Delete(&entity.Booking{})
	if err.Error != nil {
		return err.Error
	}
	if err.RowsAffected == 0 {
		return domain.ErrBookingNotFound
	}

	return nil
}
