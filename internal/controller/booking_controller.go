package controller

import (
	"BookingGo/internal/auth"
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
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

// CreateBooking godoc
// @Summary      Создать бронирование
// @Description  Создаёт новое бронирование на указанный слот (1 час). Автоматически отправляет уведомление.
// @Tags         bookings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      entity.CreateBookingRequest  true  "Данные бронирования"
// @Success      201  {object}  entity.Booking           "Созданное бронирование"
// @Failure      400  {object}  map[string]string        "Неверный формат / время в прошлом / слот занят"
// @Failure      401  {object}  map[string]string        "Требуется авторизация"
// @Failure      500  {object}  map[string]string        "Внутренняя ошибка сервера"
// @Router       /me/bookings [post]
func (bc *BookingController) CreateBooking(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
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

// CompleteBookingByID godoc
// @Summary      Завершить бронирование
// @Description  Помечает бронирование как завершённое. Доступно только staff.
// @Tags         management
// @Security     BearerAuth
// @Param        id  path      int  true  "ID бронирования"
// @Success      200  {object}  entity.Booking           "Обновлённое бронирование"
// @Failure      400  {object}  map[string]string        "ID должен быть числом / бронирование ещё не завершено"
// @Failure      401  {object}  map[string]string        "Требуется авторизация"
// @Failure      403  {object}  map[string]string        "Доступ запрещён"
// @Failure      404  {object}  map[string]string        "Бронирование не найдено"
// @Failure      500  {object}  map[string]string        "Внутренняя ошибка сервера"
// @Router       /staffPanel/bookings/completeBooking/{id} [post]
func (bc *BookingController) CompleteBookingByID(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
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

// GetAllBookings godoc
// @Summary      Получить все бронирования
// @Description  Возвращает список всех бронирований в системе. Доступно только staff и admin.
// @Tags         management
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   entity.Booking           "Список всех бронирований"
// @Failure      401  {object}  map[string]string        "Требуется авторизация"
// @Failure      403  {object}  map[string]string        "Доступ запрещён"
// @Failure      500  {object}  map[string]string        "Внутренняя ошибка сервера"
// @Router       /staffPanel/bookings [get]
// @Router       /adminPanel/bookings [get]
func (bc *BookingController) GetAllBookings(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
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

// GetMyBookings godoc
// @Summary      Получить мои бронирования
// @Description  Возвращает список всех бронирований текущего пользователя
// @Tags         bookings
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   entity.Booking           "Список бронирований"
// @Failure      401  {object}  map[string]string        "Требуется авторизация"
// @Failure      500  {object}  map[string]string        "Внутренняя ошибка сервера"
// @Router       /me/bookings [get]
func (bc *BookingController) GetMyBookings(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
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

// CancelMyBooking godoc
// @Summary      Отменить бронирование
// @Description  Отменяет бронирование, если оно ещё не началось
// @Tags         bookings
// @Security     BearerAuth
// @Param        id  path      int  true  "ID бронирования"
// @Success      204  {object}  nil                    "Бронирование отменено"
// @Failure      400  {object}  map[string]string        "ID должен быть числом / бронирование уже активно"
// @Failure      401  {object}  map[string]string        "Требуется авторизация"
// @Failure      404  {object}  map[string]string        "Бронирование не найдено"
// @Failure      500  {object}  map[string]string        "Внутренняя ошибка сервера"
// @Router       /me/bookings/cancelBooking/{id} [post]
func (bc *BookingController) CancelMyBooking(c *gin.Context) {
	currentUser := auth.GetCurrentUser(c)
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
