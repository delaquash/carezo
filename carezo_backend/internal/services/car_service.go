package services

import (
	// "database/sql"
	"encoding/json"
	"errors"
	"fmt"

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
