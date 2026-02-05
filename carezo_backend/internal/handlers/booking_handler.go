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


func (h *BookingHandler) CreateBooking(c *gin.Context) {
	// get auth user from jwt middleware
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateBookingRequest

	
	
}