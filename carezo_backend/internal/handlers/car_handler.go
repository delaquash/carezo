package handlers

import (
	"fmt"
	"net/http"
	"time"

	models "github.com/delaquash/carezo/internal/model"
	"github.com/delaquash/carezo/internal/services"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)

type CarHandler struct {
	carService *services.CarService
}

func NewCarHandler() *CarHandler {
	return &CarHandler{
		carService: services.NewCarService(),
	}
}

// admin create new car
// POST /api/admin/cars

func (h *CarHandler) CreateCar(c *gin.Context){
	var req models.CreateCarRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	car, err := h.carService.CreateCar(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, "Car created successfully", car)
}

// Get single car details
// GET /api/cars/:id

func (h *CarHandler) GetCar(c *gin.Context) {
	carID := c.Param("id")

	car, err := h.carService.GetCarByID(carID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Car retrieved successfully", car)
}

// PUT /api/admin/cars/id
func (h *CarHandler) UpdateCar(c *gin.Context) {
	carID := c.Param("id")

	var req models.UpdateCarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	car, err := h.carService.UpdateCar(carID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Car Updated Successfully", car)
}


// Delete /api/admin/car/:id
func(h *CarHandler) DeleteCar(c *gin.Context) {
	carID := c.Param("id")

	err := h.carService.DeleteCar(carID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Car deleted successfully", nil)
}

// GET /api/cars/search
// Query params: brand, model, min_year, max_year, color, transmission, fuel_type,
//               min_seats, max_seats, location, is_available, sort_by, order_by, page, per_page

func (h *CarHandler) SearchCars(c *gin.Context){
	var req models.SearchCarsRequest

	// bind every query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	// default if no query is provided
	if req.Page == 0 {
		req.Page = 1
	}

	if req.PerPage == 0 {
		req.PerPage = 20
	}

	result, err := h.carService.SearchCars(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
	}

	response.Success(c, http.StatusOK, "Cars retrieved successfull", result)
}

// Get  /api/cars
// query params: pag, per_page

func (h *CarHandler) ListAllCars(c *gin.Context) {
	// create search request with pagination
	var req models.SearchCarsRequest
	req.Page = 1
	req.PerPage = 20

	// Get page and per_page from query if provided
	if page, ok :=c.GetQuery("page"); ok {
		var p int
		if _, err := fmt.Sscanf(page, "%d", &p); err == nil && p > 0 {
			req.Page = p
		}
	}

	if perPage, ok := c.GetQuery("per_page"); ok {
		var pp int
		if _, err := fmt.Sscanf(perPage, "%d", &pp); err == nil && pp > 0 && pp <= 100 {
			req.PerPage = pp
		}
	}

	available := true
	req.IsAvailable = &available

	result, err := h.carService.SearchCars(&req)

	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Cars Retrieved Successfully", result)
}


// GET /api/cars/available
// Query params: pickup_date, return_date (ISO 8601 format)

func(h *CarHandler) GetAvailableCars(c *gin.Context) {
	pickupDateStr := c.Query("pickup_date")
	returnDateStr := c.Query("return_date")


	if pickupDateStr == "" || returnDateStr == "" {
		response.Error(c, http.StatusBadRequest, "Pick Up Date and ReturnDate are required")
	}


	// Parse dates
	pickupDate, err := time.Parse(time.RFC3339, pickupDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid pickup_date format. Use ISO 8601 (e.g., 2024-01-15T10:00:00Z)")
		return
	}

	returnDate, err := time.Parse(time.RFC3339, returnDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid return_date format. Use ISO 8601 (e.g., 2024-01-20T18:00:00Z)")
		return
	}

	// Validate dates
	if returnDate.Before(pickupDate) {
		response.Error(c, http.StatusBadRequest, "return_date must be after pickup_date")
		return
	}

	cars, err := h.carService.GetAvailableCars(pickupDate, returnDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Available cars retrieved successfully", gin.H{
		"cars":        cars,
		"pickup_date": pickupDate,
		"return_date": returnDate,
		"total":       len(cars),
	})
}