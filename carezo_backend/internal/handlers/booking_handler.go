package handlers

import (
	"fmt"
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

// route
// POST /api/bookings
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	// get auth user from jwt middleware
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	fmt.Printf("DEBUG: attempting booking for user_id: %s\n", userID.(string))
	var req models.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid Request: "+err.Error())
		return
	}

	booking, err := h.bookingService.CreateBooking(userID.(string), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, "Booking created successfully", booking)
}

func (h *BookingHandler) GetBooking(c *gin.Context) {
	bookingID := c.Param("id")

	booking, err := h.bookingService.GetBookingByID(bookingID)

	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Booking retrieved successfully", booking)
}

func (h *BookingHandler) UpdateBooking(c *gin.Context) {
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	bookingID := c.Param("id")

	var req models.UpdateBookingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	booking, err := h.bookingService.UpdateBooking(bookingID, userID.(string), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "booking updated successfully", booking)
}

// GET /api/bookings
//    ?status=pending&page=1&limit=10

func (h *BookingHandler) ListUserBooking(c *gin.Context) {
	// get authenticated user ID from jwt middleware
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// bind query params
	var req models.ListBookingsRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query params: "+err.Error())
		return
	}

	// apply defaults here so we can use them in the meta
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}
	if req.Limit > 50 {
		req.Limit = 50
	}

	bookings, total, err := h.bookingService.GetUserBookings(userID.(string), req.Status, req.Page, req.Limit)

	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// calcualte total pages
	totalPages := total / req.Limit

	if total%req.Limit != 0 {
		totalPages++
	}

	response.Success(c, http.StatusOK, "Booking retrieved successfully", gin.H{
		"bookings": bookings,
		"meta": gin.H{
			"total":       total,
			"page":        req.Page,
			"limit":       req.Limit,
			"total_pages": totalPages,
		},
	})
}

//  POST /api/bookings/:id/cancel

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	// get authenticated user ID from JWT middleware
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	bookingID := c.Param("id")

	var req models.CancelBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	err := h.bookingService.CancelBooking(bookingID, userID.(string), req.Reason)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Booking cancelled successfully", nil)
}
