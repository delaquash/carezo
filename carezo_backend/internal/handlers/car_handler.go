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