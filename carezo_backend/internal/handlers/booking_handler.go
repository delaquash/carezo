package handlers

import (
	"net/http"

	models "github.com/delaquash/carezo/internal/model"
	"github.com/delaquash/carezo/internal/services"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)


type BookingHandler struct {
	bookingService *services.BookingService 
}

func NewBookingHandler() *BookingHandler {
	return &BookingHandler{
		bookingService: services.NewBookingService(),
	}
}

// POST /api/bookings
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	// get auth user from jwt middleware
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid Request: " +err.Error())
		return
	}

	booking, err := h.bookingService.CreateBooking(userID.(string), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, "Booking created successfully", booking)
}