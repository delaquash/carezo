package handlers

import (
	"net/http"

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


