package controller

import (
	"BookingGo/internal/entity"
	"BookingGo/internal/middleware"
	"BookingGo/internal/usecase"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BookingController struct {
	bookingUsecase *usecase.BookingUsecase
}

func NewBookingController(bookingUsec *usecase.BookingUsecase) *BookingController {
	return &BookingController{bookingUsecase: bookingUsec}
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

	booking, err := bc.bookingUsecase.CreateBooking(currentUser.UserID, &req)
	if err != nil {
		if errors.Is(err, usecase.ErrOnlyForClient) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка, запись доступна только для клиента"})
		} else if errors.Is(err, usecase.ErrInvalidTimeRange) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Недействительное время записи"})
		} else if errors.Is(err, usecase.ErrSlotNotAvailable) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Эти дата и время недоступны для записи"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка в записи на услугу"})
		}
		return
	}

	c.JSON(http.StatusCreated, booking)
}

func (bc *BookingController) GetMyBookings(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	bookings, err := bc.bookingUsecase.GetMyBookings(currentUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки записей"})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

func (bc *BookingController) DeleteBooking(c *gin.Context) {
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

	err = bc.bookingUsecase.DeleteMyBooking(bookingID, currentUser.UserID)
	if err != nil {
		if errors.Is(err, usecase.ErrBookingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Запись с такими параметрами не найдена или вы не имеете к ней доступа",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления записи"})
		}
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
