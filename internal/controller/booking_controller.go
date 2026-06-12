package controller

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/middleware"
	"BookingGo/internal/usecase"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BookingController struct {
	bookingUseCase *usecase.BookingUseCase
}

func NewBookingController(bookingUseCase *usecase.BookingUseCase) *BookingController {
	return &BookingController{bookingUseCase: bookingUseCase}
}

func (bc *BookingController) CreateBooking(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	var req entity.CreateBookingRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	ctx := c.Request.Context()
	booking, err := bc.bookingUseCase.CreateBooking(ctx, currentUser.UserID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTimeRange) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Недействительное время записи"})
		} else if errors.Is(err, domain.ErrSlotNotAvailable) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Эти дата и время недоступны для записи"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка в записи на услугу"})
		}
		return
	}

	c.JSON(http.StatusCreated, booking)
}

func (bc *BookingController) CompleteBookingByID(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID должен быть числом",
		})
		return
	}

	booking, err := bc.bookingUseCase.CompleteBookingByID(bookingID)
	if err != nil {
		if errors.Is(err, domain.ErrBookingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Такая запись не найдена"})
		} else if errors.Is(err, domain.ErrBookingNotFinished) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Время конца записи еще не прошло, нельзя поменять статус записи"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибки при изменении статуса записи"})
		}
		return
	}

	c.JSON(http.StatusOK, booking)
}

func (bc *BookingController) GetAllBookings(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	bookings, err := bc.bookingUseCase.GetAllBookings(currentUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки записей"})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

func (bc *BookingController) GetMyBookings(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	bookings, err := bc.bookingUseCase.GetMyBookings(currentUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки записей"})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

func (bc *BookingController) CancelMyBooking(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID должен быть числом",
		})
		return
	}

	err = bc.bookingUseCase.CancelMyBooking(bookingID, currentUser.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrBookingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Запись с такими параметрами не найдена или вы не имеете к ней доступа",
			})
		} else if errors.Is(err, domain.ErrBookingIsActive) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Время записи уже активно, запись нельзя отменить",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отмены записи"})
		}
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
