package repository

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"errors"
	"time"

	"gorm.io/gorm"
)

type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(DB *gorm.DB) *BookingRepository {
	return &BookingRepository{db: DB}
}

func (r *BookingRepository) Create(booking *entity.Booking) error {
	return r.db.Create(booking).Error
}

func (r *BookingRepository) IsSlotAvailable(start, end time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&entity.Booking{}).
		Where("slot_start < ? AND ? < slot_end", end, start).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (r *BookingRepository) GetAllBookingsByUserID(id int) ([]entity.Booking, error) {
	var bookings []entity.Booking
	err := r.db.Where("user_id = ?", id).Order("slot_start ASC").Find(&bookings).Error
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *BookingRepository) GetBookingByID(id int) (*entity.Booking, error) {
	var booking entity.Booking
	err := r.db.First(&booking, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBookingNotFound
		}
		return nil, err
	}

	return &booking, nil
}

func (r *BookingRepository) SetBookingComplete(bookingID int) (*entity.Booking, error) {
	result := r.db.Model(&entity.Booking{}).Where("id = ?", bookingID).Update("status", enum.StatusCompleted)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrBookingNotFound
	}

	return r.GetBookingByID(bookingID)
}

func (r *BookingRepository) SetBookingCancel(bookingID, userID int) (*entity.Booking, error) {
	result := r.db.Model(&entity.Booking{}).Where("id = ? AND user_id = ?", bookingID, userID).Update("status", enum.StatusCancelled)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrBookingNotFound
	}

	return r.GetBookingByID(bookingID)
}
