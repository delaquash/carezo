package services

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
	"github.com/google/uuid"
)


type ReviewService struct{}

func NewReviewService() *ReviewService {
	return &ReviewService{}
}

// User create review after booking
func (r *ReviewService)  CreateReview(userID string, req *models.CreateReviewRequest)(*models.Review, error) {
	// check if booking exist and belongs to a driver
	var booking struct {
		DriverID string `db:"driver_id"`
		Status   string `db:"status"`
	}

	query := `SELECT driver_id, status FROM bookings WHERE id = $1 AND user_id = $2`
	err := database.DB.Get(&booking, query, userID, req.BookingID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Booking not found or doesnt belong to you")
		}
		return nil, fmt.Errorf("Database error: %w", err)
	}

	// check if booking is completed 
	if booking.Status != "completed" {
		return nil, errors.New("Can only review completed booking")
	}

	// check if review is exist for this booking before
	var exist bool
	query = `SELECT EXISTS(SELECT 1 FROM reviews WHERE booking_id = $1)`
	err = database.DB.Get(&exist, query, req.BookingID)
	if err != nil {
		return nil, fmt.Errorf("Database error: %w", err)
	}

	if exist {
		return nil, errors.New("You have already reviewed this booking")
	}

	// create review
	reviewID := uuid.New().String()
	query = `
		INSERT INTO reviews (
			id, booking_id, user_id, driver_id, rating,
			punctuality_rating, professionalism_rating, vehicle_condition_rating,
			title, comment, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'published'
		)
			RETURNING *
	`

	var review models.Review

	err = database.DB.Get(&review, query, reviewID, req.BookingID, userID, booking.DriverID, req.Rating, req.PunctualityRating,req.ProfessionalismRating, req.VehicleConditionRating, req.Title, req.Comment)

	if err != nil {
		return nil, fmt.Errorf("Failed to create reviews: %w", err)
	}
	// automatically update drivers average rating
	return  &review, nil
}

// Get single review by ID
func (r *ReviewService) GetReviewByID(reviewID string)(*models.Review, error) {
	var review models.Review
	query := `SELECT * FROM reviews WHERE id = $1`
	err := database.DB.Get(&review, reviewID, query)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Review not found")
		}
		return nil, fmt.Errorf("Database error: %w", err)
	}
	return &review, nil
}