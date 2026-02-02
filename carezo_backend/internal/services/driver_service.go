package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
func (s *DriverService) CreateDriver(req *models.CreateDriverRequest) (*models.Driver, error) {
	// check if vehicle exist using license number and has not been soft-deleted (deleted_at IS NULL)
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM drivers WHERE license_number = $1 AND deleted_at is NULL)`

	err := database.DB.Get(&exists, query, req.LicenseNumber)

	if err != nil {
		return nil, fmt.Errorf("Data error: %w", err)
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
func (s *DriverService) GetDriverByID(driverID string) (*models.Driver, error) {
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

func (s *DriverService) UpdateDriver(driverID string, req *models.UpdateDriverRequest) (*models.Driver, error) {
	// check if driver exist
	_, err := s.GetDriverByID(driverID)

	if err != nil {
		return nil, err
	}

	// dynamic update query that only update provided fields
	var updates []string
	var args []interface{}

	argCount := 1

	if req.FirstName != nil {
		updates = append(updates, fmt.Sprintf("first_name = $%d", argCount))
		args = append(args, *req.FirstName)
		argCount++
	}
	if req.LastName != nil {
		updates = append(updates, fmt.Sprintf("last_name = $%d", argCount))
		args = append(args, *req.LastName)
		argCount++
	}
	if req.Age != nil {
		updates = append(updates, fmt.Sprintf("age = $%d", argCount))
		args = append(args, *req.Age)
		argCount++
	}
	if req.Gender != nil {
		updates = append(updates, fmt.Sprintf("gender = $%d", argCount))
		args = append(args, *req.Gender)
		argCount++
	}
	if req.State != nil {
		updates = append(updates, fmt.Sprintf("state = $%d", argCount))
		args = append(args, *req.State)
		argCount++
	}
	if req.Religion != nil {
		updates = append(updates, fmt.Sprintf("religion = $%d", argCount))
		args = append(args, *req.Religion)
		argCount++
	}
	if req.Complexion != nil {
		updates = append(updates, fmt.Sprintf("complexion = $%d", argCount))
		args = append(args, *req.Complexion)
		argCount++
	}
	if req.Height != nil {
		updates = append(updates, fmt.Sprintf("height = $%d", argCount))
		args = append(args, *req.Height)
		argCount++
	}
	if req.PhoneNumber != nil {
		updates = append(updates, fmt.Sprintf("phone_number = $%d", argCount))
		args = append(args, *req.PhoneNumber)
		argCount++
	}
	if req.Email != nil {
		updates = append(updates, fmt.Sprintf("email = $%d", argCount))
		args = append(args, *req.Email)
		argCount++
	}
	if req.LicenseNumber != nil {
		updates = append(updates, fmt.Sprintf("license_number = $%d", argCount))
		args = append(args, *req.LicenseNumber)
		argCount++
	}
	if req.LicenseExpiryDate != nil {
		expiryDate, err := time.Parse("1991-01-11", *req.LicenseExpiryDate)

		if err != nil {
			return nil, errors.New("Invalid license expiry date format")
		}
		updates = append(updates, fmt.Sprintf("Licence expiry date = $%d", argCount))
		args = append(args, expiryDate)
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
	if req.Languages != nil {
		languagesJSON, _ := json.Marshal(req.Languages)
		updates = append(updates, fmt.Sprintf("Languages = $%d", argCount))
		args = append(args, languagesJSON)
		argCount++
	}
	if len(updates) == 0 {
		return nil, errors.New("No fields to update")
	}

	// add updated_at
	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")

	// add driverID to args
	args = append(args, driverID)

	// execute update fields
	query := fmt.Sprintf(
		`
		UPDATE drivers
		SET %s
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING *
		`, strings.Join(updates, ", "), argCount)

	var driver models.Driver
	err = database.DB.Get(&driver, query, args...)

	if err != nil {
		return nil, fmt.Errorf("Failed to update driver: %w", err)
	}
	return &driver, nil
}

// Soft delete driver
func (s *DriverService) DeleteDriver(driverID string) error {

	// check if driver exist
	_, err := s.GetDriverByID(driverID)

	if err != nil {
		return err
	}

	// soft delete if driver exist
	query := `UPDATE drivers SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`

	result, err := database.DB.Exec(query, driverID)

	if err != nil {
		return fmt.Errorf("Failed to delete driver: %w", err)
	}

	rows, _ := result.RowsAffected()

	if rows == 0 {
		return errors.New("Driver not found")
	}
	return nil
}

// Search for drivers and filter with pagination
func (s *DriverService) SearchDrivers(req *models.SearchDriversRequest) (*models.DriverListResponse, error) {
	var conditions []string
	var args []interface{}
	argCount := 1

	// exclude deleted acrs
	conditions = append(conditions, "deleted_at IS NULL")

	if req.Gender != nil {
		conditions = append(conditions, fmt.Sprintf("gender = $%d", argCount))
		args = append(args, *req.Gender)
		argCount++
	}
	if req.State != nil {
		conditions = append(conditions, fmt.Sprintf("LOWER(state) = LOWER($%d)", argCount))
		args = append(args, *req.State)
		argCount++
	}
	if req.Religion != nil {
		conditions = append(conditions, fmt.Sprintf("LOWER(religion) = LOWER($%d)", argCount))
		args = append(args, *req.Religion)
		argCount++
	}
	if req.Complexion != nil {
		conditions = append(conditions, fmt.Sprintf("LOWER(complexion) = LOWER($%d)", argCount))
		args = append(args, *req.Complexion)
		argCount++
	}
	if req.MinAge != nil {
		conditions = append(conditions, fmt.Sprintf("age >= $%d", argCount))
		args = append(args, *req.MinAge)
		argCount++
	}
	if req.MaxAge != nil {
		conditions = append(conditions, fmt.Sprintf("age <= $%d", argCount))
		args = append(args, *req.MaxAge)
		argCount++
	}
	if req.MinHeight != nil {
		conditions = append(conditions, fmt.Sprintf("height >= $%d", argCount))
		args = append(args, *req.MinHeight)
		argCount++
	}
	if req.MaxHeight != nil {
		conditions = append(conditions, fmt.Sprintf("height <= $%d", argCount))
		args = append(args, *req.MaxHeight)
		argCount++
	}
	if req.MinExperience != nil {
		conditions = append(conditions, fmt.Sprintf("years_of_experience >= $%d", argCount))
		args = append(args, *req.MinExperience)
		argCount++
	}
	if req.MinRating != nil {
		conditions = append(conditions, fmt.Sprintf("average_rating >= $%d", argCount))
		args = append(args, *req.MinRating)
		argCount++
	}
	if req.IsAvailable != nil {
		conditions = append(conditions, fmt.Sprintf("is_available = $%d", argCount))
		args = append(args, *req.IsAvailable)
		argCount++
	}
	if req.Language != nil {
		conditions = append(conditions, fmt.Sprintf("languages::text ILIKE $%d", argCount))
		args = append(args, "%"+*req.Language+"%")
		argCount++
	}
	// count total
	whereClause := strings.Join(conditions, " AND ")
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM drivers WHERE %s", whereClause)

	var total int
	err := database.DB.Get(&total, countQuery, args...)

	if err != nil {
		return nil, fmt.Errorf("Failed to count drivers: %w", err)
	}

	// Build order by

	orderBy := "created_at DESC"

	if req.SortBy != "" {
		order := "ASC"
		if req.OrderBy == "desc" {
			order = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", req.SortBy, order)
	}

	// 4. Pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	// query drivers
	query := fmt.Sprintf(`
		SELECT * FROM drivers
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argCount, argCount+1)

	args = append(args, perPage, offset)

	var drivers []*models.Driver

	err = database.DB.Select(&drivers, query, args...)

	if err != nil {
		return nil, fmt.Errorf("Failed to fetch drivers: %w", err)
	}
	// calculate total pages
	totalPages := (total + perPage - 1) / perPage

	// response
	return &models.DriverListResponse{
		Drivers: drivers,
		Pagination: models.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
		Filters: map[string]interface{}{
			"gender":     req.Gender,
			"state":      req.State,
			"religion":   req.Religion,
			"complexion": req.Complexion,
		},
	}, nil
}

func (s *DriverService) GetDriverReviews(driverID string) ([]*models.Review, error) {
	var reviews []*models.Review
	query := `
		SELECT * FROM reviews
		WHERE driver_id = $1 AND status = "published"
		ORDER BY created_at DESC
	`

	err := database.DB.Select(&reviews, query, driverID)

	if err != nil {
		return nil, fmt.Errorf("Failed to fetch reviews: %w", err)
	}
	return reviews, nil
}
