package handlers

import (
	"fmt"
	"net/http"

	models "github.com/delaquash/carezo/internal/model"
	"github.com/delaquash/carezo/internal/services"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)

// DriverHandler needs THREE services: drivers, reviews (drivers expose a review
// creation endpoint), and Cloudinary for profile photo cleanup
type DriverHandler struct {
	driverService      *services.DriverService
	reviewServices     *services.ReviewService
	cloudinaryServices *services.CloudinaryService
}

func NewDriverHandler(cloudinaryServices *services.CloudinaryService) *DriverHandler {

	return &DriverHandler{
		driverService:      services.NewDriverService(),
		reviewServices:     services.NewReviewService(),
		cloudinaryServices: cloudinaryServices,
	}
}

// Admin creates new driver, photo is optional
// POST /api/admin/drivers
func (h *DriverHandler) CreateDriver(c *gin.Context) {
	var req models.CreateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	// no length-check needed here — a driver has at most ONE photo, not an array

	driver, err := h.driverService.CreateDriver(&req)

	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Driver created successfully", driver)
}

// Get single driver details  or profile and it is prublic
// GET /api/driver/:id\
func (h *DriverHandler) GetDriver(c *gin.Context) {
	driverID := c.Param("id") //get driver ID

	driver, err := h.driverService.GetDriverByID(driverID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Driver retrieved successfully", driver)
}

// Admin updates driver details
// PUT /api/admin/drivers/:id

func (h *DriverHandler) UpdateDriver(c *gin.Context) {
	driverID := c.Param("id")

	var req models.UpdateDriverRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	// write new driver to postgresql first
	driver, err := h.driverService.UpdateDriver(driverID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// only delete the OLD photo if BOTH a new photo was sent AND an old
	// public_id was provided to tell us what to remove — otherwise we'd
	// have nothing to delete (e.g. driver never had a photo before)
	if req.ProfileImageURL != nil && req.OldProfileImagePublicID != nil {
		go func(oldID string) {
			// gorountine background doesnt block the response
			if err := h.cloudinaryServices.DeleteImage(oldID); err != nil {
				fmt.Printf("[Cloudinary] failed to delete old driver photos  %s: %v\n", oldID, err)
			}
		}(*req.OldProfileImagePublicID) // dereference the pointer to get the string value
	}
	response.Success(c, http.StatusOK, "Driver updated successfully", driver)
}

// Delete driver by admin — soft-delete + remove their profile photo
// DELETE /api/admin/drivers/:id

func (h *DriverHandler) DeleteDriver(c *gin.Context) {
	driverID := c.Param("id")
	// fetch BEFORE deleting so we still have their profile_image_public_id
	driver, err := h.driverService.GetDriverByID(driverID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	// soft-delete in PostgreSQL — driver disappears from app instantly
	if err := h.driverService.DeleteDriver(driverID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// only attempt cloudinary cleanuo if the driver had a photo
	if driver.ProfileImagePublicID != nil && *driver.ProfileImagePublicID != "" {
		go func(publicID string) {
			if err := h.cloudinaryServices.DeleteImage(publicID); err != nil {
				fmt.Printf("[Cloudinary] failed to delete driver photo %s: %v\n", publicID, err)
			}
		}(*driver.ProfileImagePublicID)
	}
	response.Success(c, http.StatusOK, "Driver deleted successfully", nil)
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

// GetDriverReviews for all drivers
// GET /api/drivers/:id/reviews
func (h *DriverHandler) GetDriverReviews(c *gin.Context) {
	driverID := c.Param("id")

	reviews, err := h.driverService.GetDriverReviews(driverID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Reviews retrieved successfully", gin.H{
		"driver_id": driverID,
		"reviews":   reviews,
		"total":     len(reviews),
	})
}

// CreateReview // POST /api/reviews  -  logged-in user reviews a driver after a completed booking
// (image handling for this lives inside ReviewService.CreateReview, called below)
func (h *DriverHandler) CreateReview(c *gin.Context) {
	// Get user ID from auth middleware / set earlier by AuthMiddleware from the JWT
	userID := c.GetString("user_id")

	var req models.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	review, err := h.reviewServices.CreateReview(userID, &req) // ownership + 3-image-max checks happen here
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Review created successfully", review)
}
