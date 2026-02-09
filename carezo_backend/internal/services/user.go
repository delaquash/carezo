package services

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}


// GetUserByID

func (s *UserService) GetUserByID(userID string)(*models.User, error) {
	var user models.User

	query := `SELECT * FROM users WHERE id =$1 AND deleted_at IS NULL`
	err := database.DB.Get(&user, query, userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("User not found")
		}
		return nil, fmt.Errorf("Database error: %w", err)
	}
	return &user, nil
}


// GetUserByEmail

func(s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User

	query := `SELECT * FROM users WHERE email =$1 AND deleted_at IS NULL`
	err := database.DB.Get(&query, email, user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("User not found")
		}
		return nil, fmt.Errorf("Database error: %w", err)
	}
	return &user, nil
} 

// CompleteProfile completes user registration after OTP verification
func (s *UserService) CompleteProfile(userID string, req *models.CompleteProfileRequest) (*models.User, error) {
	// this is to fetch the user

	user, err := s.GetUserByID(userID)

	if err != nil {
		return nil, err
	}

	// check if user email is verified
	if !user.EmailVerified {
		return nil, errors.New("Email must be verified before completing profile")
	}
	// check if profile already completed

	if user.FirstName != "" && user.LastName != "" {
		return nil, errors.New("Profile already completed")
	}

	// Update user profile with complete information
	query := `
		UPDATE users
		SET    first_name = $1,
			   last_name  = $2,
			   phone_number = $3,
			   age = $4,
			   profession = $5,
			   location = $6,
			   profile_image_url = $7
		WHERE  id         = $8
		  AND  deleted_at IS NULL
		RETURNING *
	`

	var completeUpdate models.User
	err = database.DB.Get(&completeUpdate, query,
		req.FirstName,
		req.LastName,
		req.PhoneNumber,
		req.Age,
		req.Profession,
		req.Location,
		req.ProfileImageURL,
		userID,
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to complete profile: %w", err)
	}

	return &completeUpdate, nil
}

 // UpdateUserProfile
func (s *UserService) UpdateProfile(userID string, updates map[string]interface{})(*models.User, error){
	// fetch user first
	_,err := s.GetUserByID(userID)

	if err != nil {
		return nil, err
	}

	// build update query
	var setClauses []string
	var args []interface{}
	argCount := 1

	// fields that can be updated
	allowUpdateField := map[string]bool{
		"first_name":   		true,
		"last_name":    		true,
		"phone_number": 		true,
		"age":          		true,
		"profession":   		true,
		"location":     		true,
		"profile_image_url":   	true,
	}

	for fields, value := range updates {
		if !allowUpdateField[fields] {
			// this skips fields not allowed to be updated
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", fields, argCount))
		args = append(args, value)
		argCount++
	}

	if len(setClauses) == 0 {
		return nil, errors.New("No valid fields to update")
	}
	// add userID as last arg
	args = append(args, userID)

	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id    = $%d
		AND deleted_at IS NULL
		RETURNING *
	`, fmt.Sprintf("%s", setClauses[0]), argCount)


		// rebuild query with all SET clauses
	if len(setClauses) > 1 {
		query = fmt.Sprintf(`
			UPDATE users
			SET    %s
			WHERE  id         = $%d
			  AND  deleted_at IS NULL
			RETURNING *
		`, joinClauses(setClauses), argCount)
	}

	var updated models.User
	err = database.DB.Get(&updated, query, args...)
	if err != nil {
		return  nil, fmt.Errorf("Failed to update profile: %w", err)
	}

	return &updated, nil
}

// helper function to join SET clauses

func joinClauses(clauses []string) string {
	result := ""
	for i, clause := range clauses {
		if i > 0 {
			result += ", "
		}
		result += clause
	}

	return result
}

// UpdateLastLogin at timestamp
func(s *UserService) UpdateLastLogin(userID string) error {
	query := `
		UPDATE users
		SET    last_login_at = CURRENT_TIMESTAMP
		WHERE  id           =$1
			AND deleted_at   IS NULL
	`

	result, err := database.DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("Failed to update last login: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("User not found")
	}
	return nil
}


// UpdateUserStatus by admin

func (s *UserService) UpdateUserStatus(userID string, status string) error {
	// validate status
	validateStatus := map[string]bool {
		"active": true,
		"inactive": true,
		"suspended": true,
	}

	if !validateStatus[status] {
		return errors.New("Invalid status. Must be: active, inactive, or suspended")
	}

	query := `
		UPDATE users
		SET      status  =$1
		WHERE id  =$2
			AND deleted_at IS NULL	
	`

	result, err := database.DB.Exec(query, status, userID)
	if err != nil {
		return fmt.Errorf("Failed to update user status: %w", err)
	}

	rows, _ := result.RowsAffected()

	if rows == 0 {
		return errors.New("User not found")
	}
	return nil
}

// DeactivateAccount - soft delete user account

func (s *UserService) DeactivateAccount(userID string) error {
	query := `
		UPDATE users
		SET    deleted_at = CURRENT_TIMESTAMP,
		       status     = 'inactive'
		WHERE  id         = $1
		  AND  deleted_at IS NULL
	`

	result, err := database.DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("Failed to deactivate account: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("User not found")
	}
	return nil
}


func (s *UserService) GetAllUsers(status string, role string, page, limit int) ([]models.User, int, error) {
	
	// pagination defaults
	if page < 1  { page = 1 }
	if limit < 1 { limit = 10 }
	if limit > 20 { limit = 20 }

	// build WHERE clause
	var conditions []string
	var args []interface{}
	argCount := 1

	// exclude deleted users
	conditions = append(conditions, "deleted_at IS NULL")

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, status)
		argCount++
	}

	if role != "" {
		conditions = append(conditions, fmt.Sprintf("role = $%d", argCount))
		args = append(args, role)
		argCount++
	}

	// count total
	countQuery := "SELECT COUNT(*) FROM users WHERE "
	for i, cond := range conditions {
		if i > 0 {
			countQuery += " AND "
		}
		countQuery += cond
	}

	var total int
	err := database.DB.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("Database error counting users: %w", err)
	}

	// fetch page
	offset := (page - 1) * limit

	dataQuery := "SELECT * FROM users WHERE "
	for i, cond := range conditions {
		if i > 0 {
			dataQuery += " AND "
		}
		dataQuery += cond
	}
	dataQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	var users []models.User
	err = database.DB.Select(&users, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("Database error fetching users: %w", err)
	}

	return users, total, nil
}


