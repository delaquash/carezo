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
		"suspended": true
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


