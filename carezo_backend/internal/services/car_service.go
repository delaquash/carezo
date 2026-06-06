package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	//
	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
	"github.com/google/uuid"
)

type CarService struct{}

func NewCarService() *CarService {
	return &CarService{}
}

// Create car by Admin
func (s *CarService) CreateCar(req *models.CreateCarRequest) (*models.Car, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM cars WHERE license_plate = $1 AND deleted_at IS NULL)`
	err := database.DB.Get(&exists, query, req.LicensePlate)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if exists {
		return nil, errors.New("car with this license plate already exists")
	}

	if req.HourlyRate.Weekday < 0 || req.HourlyRate.Weekend < 0 || req.HourlyRate.Holiday < 0 {
		return nil, errors.New("invalid hourly rate values")
	}

	featuresJSON, err := json.Marshal(req.Features)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal features: %w", err)
	}

	hourlyRateJSON, err := json.Marshal(req.HourlyRate)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hourly rate: %w", err)
	}

	carID := uuid.New().String()

	query = `
		INSERT INTO cars (
    		id, model, year, color, license_plate,
    		engine_output, transmission, fuel_type,
    		seating_capacity, maximum_speed, mileage,
    		driver_name, driver_number, driver_miles,
    		hourly_rate, caution_fee, features,
    		current_location
	)
		VALUES (
			$1,$2,$3,$4,$5,
			$6,$7,$8,
			$9,$10,$11,
			$12,$13,$14,
			$15,$16,$17,
			$18
		)
		RETURNING *
			`

	var car models.Car

	err = database.DB.Get(&car, query,
		carID,
		req.Model,
		req.Year,
		req.Color,
		req.LicensePlate,
		req.EngineOutput,
		req.Transmission,
		req.FuelType,
		req.SeatingCapacity,
		req.MaximumSpeed,
		req.Mileage,
		req.DriverName,
		req.DriverNumber,
		req.DriverMiles,
		hourlyRateJSON,
		req.CautionFee,
		featuresJSON,
		req.CurrentLocation,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create car: %w", err)
	}

	return &car, nil
}

// get a single car by ID
func (s *CarService) GetCarByID(carID string) (*models.Car, error) {
	var car models.Car

	query := `SELECT * FROM cars WHERE id = $1 and deleted_at IS NULL`
	err := database.DB.Get(&car, query, carID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Car not found!!!")
		}
		return nil, fmt.Errorf("Database error: %w", err)
	}
	return &car, nil
}

func (s *CarService) UpdateCar(carID string, req *models.UpdateCarRequest) (*models.Car, error) {
	// check if car exist
	_, err := s.GetCarByID(carID)

	if err != nil {
		return nil, err
	}

	// dynamic update query that only update provided fields
	var updates []string
	var args []interface{}

	argCount := 1

	if req.Model != nil {
		updates = append(updates, fmt.Sprintf("model = $%d", argCount))
		args = append(args, *req.Model)
		argCount++
	}

	if req.Year != nil {
		updates = append(updates, fmt.Sprintf("year = $%d", argCount))
		args = append(args, *req.Year)
		argCount++
	}

	if req.Color != nil {
		updates = append(updates, fmt.Sprintf("color = $%d", argCount))
		args = append(args, *req.Color)
		argCount++
	}

	if req.LicensePlate != nil {
		updates = append(updates, fmt.Sprintf("license_plate = $%d", argCount))
		args = append(args, *req.LicensePlate)
		argCount++
	}
	if req.Transmission != nil {
		updates = append(updates, fmt.Sprintf("transmission = $%d", argCount))
		args = append(args, *req.Transmission)
		argCount++
	}
	if req.FuelType != nil {
		updates = append(updates, fmt.Sprintf("fuel_type = $%d", argCount))
		args = append(args, *req.FuelType)
		argCount++
	}
	if req.SeatingCapacity != nil {
		updates = append(updates, fmt.Sprintf("seating_capacity = $%d", argCount))
		args = append(args, *req.SeatingCapacity)
		argCount++
	}
	if req.Mileage != nil {
		updates = append(updates, fmt.Sprintf("mileage = $%d", argCount))
		args = append(args, *req.Mileage)
		argCount++
	}
	if req.DriverName != nil {
		updates = append(updates, fmt.Sprintf("driver_name = $%d", argCount))
		args = append(args, *req.DriverName)
		argCount++
	}
	if req.DriverNumber != nil {
		updates = append(updates, fmt.Sprintf("driver_number = $%d", argCount))
		args = append(args, *req.DriverNumber)
		argCount++
	}
	if req.CautionFee != nil {
		updates = append(updates, fmt.Sprintf("caution_fee = $%d", argCount))
		args = append(args, *req.CautionFee)
		argCount++
	}
	if req.IsAvailable != nil {
		updates = append(updates, fmt.Sprintf("is_available = $%d", argCount))
		args = append(args, *req.IsAvailable)
		argCount++
	}
	if req.Status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", argCount))
		args = append(args, *req.Status)
		argCount++
	}
	if req.CurrentLocation != nil {
		updates = append(updates, fmt.Sprintf("current_location = $%d", argCount))
		args = append(args, *req.CurrentLocation)
		argCount++
	}
	if req.Features != nil {
		featuresJSON, _ := json.Marshal(req.Features)
		updates = append(updates, fmt.Sprintf("features = $%d", argCount))
		args = append(args, featuresJSON)
		argCount++
	}
	if req.HourlyRate != nil {
		hourlyRateJSON, _ := json.Marshal(req.HourlyRate)
		updates = append(updates, fmt.Sprintf("hourly_rate = $%d", argCount))
		args = append(args, hourlyRateJSON)
		argCount++
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	// add updated_at
	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")

	// add carID to args
	args = append(args, carID)

	// Execute update
	query := fmt.Sprintf(`
		UPDATE cars
		SET %s
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING * 
		`, strings.Join(updates, ", "), argCount)
	var car models.Car
	err = database.DB.Get(&car, query, args...)

	if err != nil {
		return nil, fmt.Errorf("Failed to update car: %w", err)
	}
	return &car, nil
}

// Delete car by ID (soft delete)

func (s *CarService) DeleteCar(carID string) error {
	// check if car exist
	_, err := s.GetCarByID(carID)

	if err != nil {
		return err
	}

	// soft delete car
	query := `UPDATE cars SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`

	result, err := database.DB.Exec(query, carID)

	if err != nil {
		return fmt.Errorf("Failed to delete car: %w", err)
	}
	rows, _ := result.RowsAffected()

	if rows == 0 {
		return errors.New("Car not found or already deleted")
	}

	return nil
}

// Search for cars and filter by pagination
func (s *CarService) SearchCars(req *models.SearchCarsRequest) (*models.CarListResponse, error) {
	// build WHERE clause dynamically based on filters

	var conditions []string
	var args []interface{}
	argCount := 1 

	// exclude deleted acrs
	conditions = append(conditions, "deleted_at IS NULL")

	// apply filters

	if req.Model != nil {
		conditions = append(conditions, fmt.Sprintf("LOWER(model) LIKE LOWER($%d)", argCount))
		args = append(args, "%"+*req.Model+"%")
		argCount++
	}
	if req.MinYear != nil {
		conditions = append(conditions, fmt.Sprintf("year >= $%d", argCount))
		args = append(args, *req.MinYear)
		argCount++
	}
	if req.MaxYear != nil {
		conditions = append(conditions, fmt.Sprintf("year <= $%d", argCount))
		args = append(args, *req.MaxYear)
		argCount++
	}
	if req.Color != nil {
		conditions = append(conditions, fmt.Sprintf("LOWER(color) = LOWER($%d)", argCount))
		args = append(args, *req.Color)
		argCount++
	}
	if req.Transmission != nil {
		conditions = append(conditions, fmt.Sprintf("transmission = $%d", argCount))
		args = append(args, *req.Transmission)
		argCount++
	}
	if req.FuelType != nil {
		conditions = append(conditions, fmt.Sprintf("fuel_type = $%d", argCount))
		args = append(args, *req.FuelType)
		argCount++
	}
	if req.MinSeats != nil {
		conditions = append(conditions, fmt.Sprintf("seating_capacity >= $%d", argCount))
		args = append(args, *req.MinSeats)
		argCount++
	}
	if req.MaxSeats != nil {
		conditions = append(conditions, fmt.Sprintf("seating_capacity <= $%d", argCount))
		args = append(args, *req.MaxSeats)
		argCount++
	}
	if req.IsAvailable != nil {
		conditions = append(conditions, fmt.Sprintf("is_available = $%d", argCount))
		args = append(args, *req.IsAvailable)
		argCount++
	}
	if req.Location != nil {
		conditions = append(conditions, fmt.Sprintf("LOWER(current_location) LIKE LOWER($%d)", argCount))
		args = append(args, "%"+*req.Location+"%")
		argCount++
	}

	// 2. Count total matching cars (for pagination)
	whereClause := strings.Join(conditions, " AND ")
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM cars WHERE %s", whereClause)
	var total int
	err := database.DB.Get(&total, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to count cars: %w", err)
	}

	//  build ORDER by clause

	orderBy := "created_at DESC"

	if req.SortBy != "" {
		order := "ASC"

		if req.OrderBy != "" {
			order = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", req.SortBy, order)
	}

	// calculation for pagination

	page := req.Page

	if page < 1 {
		page = 1
	}

	perPage := req.PerPage

	if perPage < 1 {
		perPage = 10
	}

	offset := (page - 1) * perPage

	//  query cars with pagination
	query := fmt.Sprintf(`
		SELECT * FROM cars
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argCount, argCount+1)
	args = append(args, perPage, offset)

	var cars []*models.Car
	err = database.DB.Select(&cars, query, args...)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch cars: %w", err)
	}

	// claculate total pages
	totalPages := (total + perPage - 1) / perPage

	// response
	return &models.CarListResponse{
		Cars: cars,
		Pagination: models.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},

		Filters: map[string]interface{}{
			"model":        req.Model,
			"transmission": req.Transmission,
			"fuel_type":    req.FuelType,
			"is_available": req.IsAvailable,
		},
	}, nil
}

// Get available cars for given date range
func (s *CarService) GetAvailableCars(pickupDate, returnDate time.Time) ([]*models.Car, error) {
	// query cars not booked within requested period and date range
	query := `
		SELECT c.* FROM cars c
	    WHERE c.is_available = true 
		AND c.status = "active	
		AND c.deleted_at IS NULL
		AND c.id NOT IN (
			SELECT car_id FROM bookings
			WHERE status IN ("confirmed", "in_progress")
			AND (
				(pickup_date <= $1 AND return_date >= $1) OR
				(pickup_date <= $2 AND return_date >= $2) OR
				(pickup_date >= $1 AND return_date <= $1)

			)
        )
			ORDER BY created_at DESC
	`
	var cars []*models.Car
	err := database.DB.Select(&cars, query, pickupDate, returnDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch available cars: %w", err)
	}

	return cars, nil
}

func (s *CarService) GetNearbyCar(city string, page int, perPage int) ([]*models.Car, int, error) {
	if page < 1 {
		page = 0
	}

	if perPage < 1 {
		perPage = 10
	}
// Skip first 10 rows and start from row 11.
	offset := (page -1) * perPage

	
	// city = "Lagos" becomes "%Lagos%" Matches: Lagos, Lagos Island, Lagos Mainland

	searchCity := "%" + city + "%"

	var total int

	err := database.DB.Get(&total, `
		SELECT COUNT(*) FROM cars
		WHERE LOWER(current_location) ILIKE LOWER($1)
			AND is_available = true
			AND deleted_at IS NULL
	`, searchCity)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to count nearby cars: %w", err)
	}

	var cars []*models.Car

	err = database.DB.Select(&cars, `
		SELECT * FROM cars
		WHERE LOWER(current_location) ILIKE LOWER($1)
		  AND is_available = true
		  AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, searchCity, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch nearby cars: %w", err)
	}
 
	return cars, total, nil
 }
