package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
	"github.com/google/uuid"
)


type DriverService struct{}

func NewDriverService() *DriverService {
	return &DriverService{}
}


// Admin to create driver
func (s *DriverService)CreateDriver(req *models.CreateDriverRequest) (*models.Driver, error) {
	// check if vehicle exist using license number
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM drivers WHERE license_number = $1 AND deleted_at is NULL)`

	err := database.DB.Get(&exists, query, req.LicenseNumber)

	if err != nil {
		return nil, fmt.Errorf("Data error: &w", err)
	}

	if exists {
		return nil, errors.New("Driver with this license number already exist")
	}

	// Parse license expiry date
	expiryDate, err := time.Parse("2006-01-02", req.LicenseExpiryDate)
	if err != nil {
		return nil, errors.New("invalid license_expiry_date format. Use YYYY-MM-DD")
	}

	// Check if license is expired
	if expiryDate.Before(time.Now()) {
		return nil, errors.New("driver license has expired")
	}

	// Convert languages to JSON
	languagesJSON, err := json.Marshal(req.Languages)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal languages: %w", err)
	}

	// create driver in DB
	driverID := uuid.New().String()
	query = `
		INSERT INTO drivers (
			id,first_name, last_name, age, gender, state, religion, complexion, height, phone_number, email, license_number, license_expiry_date, years_of_experience, bio, languages,
			is_available, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, true, 'active'
		)
		RETURNING *
	`

	var driver models.Driver

	err = database.DB.Get(&driver, query,
		driverID, req.FirstName, req.LastName, req.Age, req.Gender,
		req.State, req.Religion, req.Complexion, req.Height,
		req.PhoneNumber, req.Email, req.LicenseNumber, expiryDate,
		req.YearsOfExperience, req.Bio, languagesJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	return &driver, nil
}

// get single driver details by ID
func(s *DriverService) GetDriverByID(driverID string) (*models.Driver, error) {
	var driver models.Driver
	query := `SELECT * FROM driver WHERE id = $1 and deleted_at IS NULL`
	err := database.DB.Get(&driver, query, driverID)

	if err != nil {
		if err == sql.ErrNoRows {
		return nil, errors.New("Driver not found")
	}
		return nil, fmt.Errorf("Database error: %w", err)
	}
	return &driver, nil
	
}