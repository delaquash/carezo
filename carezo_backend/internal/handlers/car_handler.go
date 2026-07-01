package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	models "github.com/delaquash/carezo/internal/model"
	"github.com/delaquash/carezo/internal/services"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)

// Carhandler bundles everything a car endpoint need to do it jobs
type CarHandler struct {
	carService        *services.CarService        //talks to Postgresql for car CRUD
	cloudinaryService *services.CloudinaryService //talks to cloudinary for image delete
}

// constructor -- main.go calls this one once and passes in the shared cloudinaryservice
func NewCarHandler(cloudinaryService *services.CloudinaryService) *CarHandler {
	return &CarHandler{
		carService:        services.NewCarService(), //create a fresh CarServices
		cloudinaryService: cloudinaryService,        //reuse the same cloudinary client everywhere
	}
}

// admin create new car
// POST /api/admin/cars

func (h *CarHandler) CreateCar(c *gin.Context) {
	var req models.CreateCarRequest

	// read the request body and fills req - fails if JSON is malformed
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error()) //400 + reason
		return
	}

	// images[] and image_public_ids[] MUST be the same length or we lose track
	// of which public_id belongs to which URL — breaks future deletion
	if len(req.Images) > 0  && len(req.Images) != len(req.ImagePublicIDs) {
		response.Error(c, http.StatusBadRequest, "images and image_public_ids must have the same number of items")
		return
	}

	// hard cap of 5 photos per care = enforced here
	if len(req.Images) > 5 {
		response.Error(c, http.StatusBadRequest, "maximum 5 images per car")
		return
	}

	// insert into the service layer
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
	carID := c.Param("id") //id of the car been we are getting

	car, err := h.carService.GetCarByID(carID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Car retrieved successfully", car)
}

// PUT /api/admin/cars/id
func (h *CarHandler) UpdateCar(c *gin.Context) {
	carID := c.Param("id") //id of the car been we are editing

	var req models.UpdateCarRequest //hold whatever field
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}
	// same parallel-array safety check as CreateCar, but for NEW images only
	if len(req.NewImages) != len(req.NewImagePublicIDs) {
		response.Error(c, http.StatusBadRequest,
			"new_images and new_image_public_ids must have the same number of items")
		return
	}

	// update the car row in PostgreSQL first — DB is the source of truth

	car, err := h.carService.UpdateCar(carID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	// ONLY after the DB write succeeds do we touch Cloudinary —
	// if DB had failed first, we'd be deleting images still referenced by the car

	if len(req.RemoveImagePublicIDs) > 0 {
		// go routine run in the background, HTTP response doesnt wait for this
		go func(ids []string) {
			for _, id := range ids { //loop through every public_id marked for the removal
				if err := h.cloudinaryService.DeleteImage(id); err != nil {
					// failure here is non-fatal, i jut want to log it for manual clean up latyer on
					fmt.Printf("[Cloudinary] failed to delete car %s: %v\n", id, err)
				}
			}
		}(req.RemoveImagePublicIDs) // pass this slice in so that goroutine has it own copy
	}
	response.Success(c, http.StatusOK, "Car Updated Successfully", car)
}

// Delete /api/admin/car/:id   soft-delete the car AND clean up its Cloudinary photos
func (h *CarHandler) DeleteCar(c *gin.Context) {
	carID := c.Param("id") //id of the car been we are deleting

	// fetch the car first
	car, err := h.carService.GetCarByID(carID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	// soft delete: sets deleted at timeStamp, row still exists for history/bookings
	if err := h.carService.DeleteCar(carID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var publicIDs []string //this will hold the parsed list of Cloudinary IDs to remove

	// car.ImagePublicIDs is raw JSONB bytes from Postgres — unmarshal into []string
	if len(car.ImagePublicIDs) > 0 {
		if err := json.Unmarshal([]byte(car.ImagePublicIDs), &publicIDs); err != nil {
			// parsing failed, it is logged but the car is already deleted, so response is not blocked
			fmt.Printf("[Cloudinary] failed to parse the public_ids for car %s: %v\n", carID, err)
		}

	}

	// background clean up  --- delete every photo this car had, one by one
	if len(car.ImagePublicIDs) > 0 {
		go func(ids []string) {
			for _, id := range ids {
				if err := h.cloudinaryService.DeleteImage(id); err != nil {
					fmt.Printf("[Cloudinary] failed to delete car image %s: %v\n", id, err)
				}
			}
		}(publicIDs)
	}

	response.Success(c, http.StatusOK, "Car deleted successfully", nil) //nil because no data is return
}

// GET /api/cars/search?brand=Toyota&transmission=automatic — public, full filter support
// Query params: brand, model, min_year, max_year, color, transmission, fuel_type,
//               min_seats, max_seats, location, is_available, sort_by, order_by, page, per_page

func (h *CarHandler) SearchCars(c *gin.Context) {
	var req models.SearchCarsRequest

	// bind every query parameters or 	// ShouldBindQuery maps ?key=value pairs onto the struct's `form:"..."` tags

	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	// default if no query is provided
	if req.Page == 0 {
		req.Page = 1
	}
	// guard against per_page=0 which would return nothing
	if req.PerPage == 0 {
		req.PerPage = 10
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
	req.Page = 1     //defult to page 1
	req.PerPage = 10 // olny show 20 card per page

	// Get page and per_page from query if provided
	// allow the caller to overrride via ?page=2
	if page, ok := c.GetQuery("page"); ok {
		var p int
		if _, err := fmt.Sscanf(page, "%d", &p); err == nil && p > 0 {
			req.Page = p //only accept positive integers
		}
	}
	// allow the caller to override page size via ?per_page=50, capped at 100

	if perPage, ok := c.GetQuery("per_page"); ok {
		var pp int
		if _, err := fmt.Sscanf(perPage, "%d", &pp); err == nil && pp > 0 && pp <= 100 {
			req.PerPage = pp
		}
	}

	available := true // only show available listing
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

func (h *CarHandler) GetAvailableCars(c *gin.Context) {
	pickupDateStr := c.Query("pickup_date") // raw string from the URL
	returnDateStr := c.Query("return_date")

	// both dates are required — without them we can't check overlap

	if pickupDateStr == "" || returnDateStr == "" {
		response.Error(c, http.StatusBadRequest, "Pick Up Date and ReturnDate are required")
		return
	}

	// parse using RFC3339 — the standard ISO 8601 format, e.g. 2024-01-15T10:00:00Z

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

	// sanity check — can't return a car before you picked it up
	if returnDate.Before(pickupDate) {
		response.Error(c, http.StatusBadRequest, "return_date must be after pickup_date")
		return
	}

	cars, err := h.carService.GetAvailableCars(pickupDate, returnDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	// success response and parameters to return with
	response.Success(c, http.StatusOK, "Available cars retrieved successfully", gin.H{
		"cars":        cars,
		"pickup_date": pickupDate,
		"return_date": returnDate,
		"total":       len(cars),
	})
}

// GET /api/cars/nearby?city=Lagos&page=1&per_page=10
// GET /api/cars/nearby?city=Lagos — matches cars whose current_location contains the city

func (h *CarHandler) GetNearByCars(c *gin.Context) {
	city := c.Query("city")

	if city == "" {
		response.Error(c, http.StatusBadRequest, "city query parameter is required")
		return
	}

	page := 1
	perPage := 10

	if p, ok := c.GetQuery("page"); ok {
		fmt.Sscanf(p, "%d", &page)
	}

	if pp, ok := c.GetQuery("per_page"); ok {
		fmt.Scanf(pp, "%d", &perPage)
	}

	cars, total, err := h.carService.GetNearbyCars(city, page, perPage)

	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
	}

	totalPages := (total + perPage - 1) / perPage

	response.Success(c, http.StatusOK, "nearby cars retrieved", gin.H{
		"cars":  cars,
		"city":  city,
		"total": total,
		// classic pagination math: round UP so a partial last page still counts

		"meta": gin.H{
			"page":        page,
			"per_page":    perPage,
			"total_pages": totalPages,
		},
	})
}


// GET /api/cars/popular — ranked by (rating × 0.6) + (bookings × weight)
// GET /api/cars/popular?page=1&per_page=10
func (h *CarHandler) GetPopularCars(c *gin.Context) {
	page := 1
	perPage := 10

	if p, ok := c.GetQuery("page"); ok {
		fmt.Sscanf(p, "%d", &page)
	}

	if pp, ok := c.GetQuery("per_page"); ok {
		fmt.Sscanf(pp, "%d", &perPage)
	}

	cars, total, err := h.carService.GetPopularCars(page, perPage)  // scoring happens inside the service

	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
	}

	totalPages := (total + perPage - 1) / perPage

	response.Success(c, http.StatusOK, "popular cars retrieved", gin.H{
		"cars":  cars,
		"total": total,
		"meta": gin.H{
			"page":        page,
			"per_page":    perPage,
			"total_pages": totalPages,
		},
	})
}
