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


// // UpdateUserProfile
// func (s *UserService) UpdateProfile(userID string, updates map[string]interface{})(*models.User, error){
// 	// fetch user first
// 	_,err := s.GetUserByID(userID)

// 	if err != nil {
// 		return nil, err
// 	}
// }