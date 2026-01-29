package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	// check if car exist using licence plate
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM cars WHERE licence_plate = $1 AND deleted_at IS NULL)`
	err := database.DB.Get(&exists, query, req.LicencePlate)

	if err != nil {
		return nil, fmt.Errorf("Database error: %w", err)
	}

	if exists {
		return nil, errors.New("Car with this licence plate exist already")
	}

	// converts features to JSOn
	convertFeaturesToJson, err := json.Marshal(req.Features)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert features: %w", err)
	}

	// convert hourly rate to JSON
	hourlyRateJSON, err := json.Marshal(req.HourlyRate)

	if err != nil {
		return nil, fmt.Errorf("Failed to marshal hourly)rate: %w", err)
	}

	// create car in database
	carID := uuid.New().String()
	query = `
			INSERT INTO cars (
			id, model, brand, year, color, licence_plate, engine_output, transmission, fuel_type
			seating_capacity, maximum_speed, mileage, driver_name,driver_number, driver_miles, hourly_rate,
			caution_fee, car_features, images, is_available, status, current_location
			) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, '[]'::jsonb, true, "active", $19
			)
			RETURNING *
		`

	var car models.Car

	err = database.DB.Get(&car, query,
		carID, req.Model, req.Brand, req.Year, req.Color, req.LicencePlate,
		req.EngineOutput, req.Transmission, req.FuelType, req.SeatingCapacity,
		req.MaximumSpeed, req.Mileage, req.DriverName, req.DriverNumber,
		req.DriverMiles, hourlyRateJSON, req.CautionFee, convertFeaturesToJson,
		req.CurrentLocation,
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to create car: %w", err)
	}
	return &car, nil
}


// get a single car by ID
func(s *CarService) GetCarByID(carID string)(*models.Car, error) {
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


func(s *CarService) UpdateCar(carID string, req *models.UpdateCarRequest) (*models.Car, error) {
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

	if req.Brand != nil {
		updates = append(updates, fmt.Sprintf("brand = $%d", argCount))
		args = append(args, *req.Brand)
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
	if req.CarFeatures != nil {
		featuresJSON, _ := json.Marshal(req.CarFeatures)
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
		`, strings.Join(updates, ", ")argCount
	)
	var car models.Car
	err = database.DB.Get(&car, query, args...)

	if err != nil {
		return nil, fmt.Errorf("Failed to update car: %w", err)
	}
	return &car, nil
}