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
func (h *DriverHandler) GetDriver(c *gin.Context) {
	driverID := c.Param("id")

	driver, err := h.driverService.GetDriverByID(driverID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Driver retrieved successfully", driver)
}

// Admin updates driver details
// PUT /api/admin/drivers/:id

func(h *DriverHandler) UpdateDriver(c *gin.Context) {
	driverID := c.Param("id")

	var req models.UpdateDriverRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}
	driver, err := h.driverService.UpdateDriver(driverID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Driver updated successfully", driver)
}

// Delete driver by admin
// DELETE /api/admin/drivers/:id

func (h *DriverHandler) DeleteDriver(c *gin.Context) {
	driverID := c.Param("id")

	err := h.driverService.DeleteDriver(driverID)

	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusBadRequest, "Driver deleted successfully", nil)
}

// GET /api/drivers/search
func (h *DriverHandler) SearchDrivers(c *gin.Context) {
	var req models.SearchDriversRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PerPage == 0 {
		req.PerPage = 10
	}

	result, err := h.driverService.SearchDrivers(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Driver retrieved successfully", result)
}

// ListAllDrivers
// GET /api/drivers
func (h *DriverHandler) ListAllDrivers(c *gin.Context) {
	var req models.SearchDriversRequest
	req.Page = 1
	req.PerPage = 10

	// show only available drivers
	available := true
	req.IsAvailable = &available

	result, err := h.driverService.SearchDrivers(&req)

	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Driver retrieved successfully", result)
}

// // GetDriverReviews for all drivers
// // GET /api/drivers/:id/reviews
// func (h *DriverHandler) GetDriverReviews(c *gin.Context) {
// 	driverID := c.Param("id")

// 	reviews, err := h.driverService.GetDriverReviews(driverID)
// 	if err != nil {
// 		response.Error(c, http.StatusInternalServerError, err.Error())
// 	}
// }