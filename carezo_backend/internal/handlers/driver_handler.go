package handlers

import (
	"net/http"

	models "github.com/delaquash/carezo/internal/model"
	"github.com/delaquash/carezo/internal/services"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)

type DriverHandler struct {
	driverService *services.DriverService
	reviewServices *services.ReviewService
}


func NewDriverHandler() *DriverHandler {
	return &DriverHandler{
		driverService: services.NewDriverService(),
		reviewServices: services.NewReviewService(),
	}
}

// Admin creates new driver
// POST /api/admin/drivers
func(h *DriverHandler) CreateDriver(c *gin.Context) {
	var req models.CreateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	driver, err := h.driverService.CreateDriver(&req)

	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Driver created successfully", driver)
}

// Get single driver details
// GET /api/driver/:id\
