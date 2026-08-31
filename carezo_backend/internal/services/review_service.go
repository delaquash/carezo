package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ReviewService struct{}

func NewReviewService() *ReviewService {
	return &ReviewService{}
}

// CreateReview — imageURLs/imagePublicIDs arrive as already-uploaded
// Cloudinary results, NOT as part of CreateReviewRequest. Same separation
// of concerns as UploadDriverDocuments: this service never touches
// multipart/HTTP mechanics, only final URLs the handler already uploaded.
func (r *ReviewService) CreateReview(userID string, req *models.CreateReviewRequest, imageURLs []string, imagePublicIDs []string) (*models.Review, error) {
	if len(imageURLs) != len(imagePublicIDs) {
		return nil, errors.New("images and image_public_ids must have the same number of items")
	}
	if len(imageURLs) > 3 {
		return nil, errors.New("a maximum of 3 images allowed per review")
	}

	tx, err := database.DB.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var booking struct {
		DriverID string `db:"driver_id"`
		CarID    string `db:"car_id"`
		Status   string `db:"status"`
	}
	// Fixed: missing comma before `status` — was silently aliasing car_id
	// AS status, discarding the real status column entirely.
	query := `SELECT driver_id, car_id, status FROM bookings WHERE id = $1 AND user_id = $2`
	err = tx.Get(&booking, query, req.BookingID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Booking not found or doesnt belong to you")
		}
		return nil, fmt.Errorf("Database error: %w", err)
	}

	if booking.Status != "completed" {
		return nil, errors.New("Can only review completed booking")
	}

	var exists bool
	err = tx.Get(&exists, `SELECT EXISTS(SELECT 1 FROM reviews WHERE booking_id = $1)`, req.BookingID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if exists {
		return nil, errors.New("you have already reviewed this booking")
	}

	imagesJSON, err := json.Marshal(imageURLs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal images: %w", err)
	}
	imagePublicIDsJSON, err := json.Marshal(imagePublicIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal image public ids: %w", err)
	}

	reviewID := uuid.New().String()
	query = `
		INSERT INTO reviews (
			id, booking_id, user_id, driver_id, car_id, rating,
			punctuality_rating, professionalism_rating, vehicle_condition_rating,
			title, comment, images, image_public_ids, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, 'published'
		)
		RETURNING *
	`
	var review models.Review
	err = tx.Get(&review, query,
		reviewID, req.BookingID, userID, booking.DriverID, booking.CarID,
		req.Rating, req.PunctualityRating, req.ProfessionalismRating, req.VehicleConditionRating,
		req.Title, req.Comment, imagesJSON, imagePublicIDsJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to create reviews: %w", err)
	}

	if err := recalculateDriverRating(tx, booking.DriverID); err != nil {
		return nil, err
	}
	if err := recalculateCarRating(tx, booking.CarID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit review: %w", err)
	}

	return &review, nil
}

// UpdateReview — full edit of the rating/text fields (NOT images, see
// EditReviewImages for that). Dynamic SET clause, same pattern as
// UpdateCar/UpdateDriver: only fields the caller actually provided get
// touched. Always recalculates BOTH driver and car ratings afterward,
// since we don't know in advance which rating field changed — cheap
// enough to just always recompute rather than track that conditionally.
func (r *ReviewService) UpdateReview(reviewID, requesterID string, isAdmin bool, req *models.UpdateReviewRequest) (*models.Review, error) {
	tx, err := database.DB.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var existing models.Review
	err = tx.Get(&existing, `SELECT * FROM reviews WHERE id = $1`, reviewID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("review not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if !isAdmin && existing.UserID != requesterID {
		return nil, errors.New("you can only edit your own review")
	}

	var updates []string
	var args []interface{}
	argCount := 1

	if req.Rating != nil {
		updates = append(updates, fmt.Sprintf("rating = $%d", argCount))
		args = append(args, *req.Rating)
		argCount++
	}
	if req.PunctualityRating != nil {
		updates = append(updates, fmt.Sprintf("punctuality_rating = $%d", argCount))
		args = append(args, *req.PunctualityRating)
		argCount++
	}
	if req.ProfessionalismRating != nil {
		updates = append(updates, fmt.Sprintf("professionalism_rating = $%d", argCount))
		args = append(args, *req.ProfessionalismRating)
		argCount++
	}
	if req.VehicleConditionRating != nil {
		updates = append(updates, fmt.Sprintf("vehicle_condition_rating = $%d", argCount))
		args = append(args, *req.VehicleConditionRating)
		argCount++
	}
	if req.Title != nil {
		updates = append(updates, fmt.Sprintf("title = $%d", argCount))
		args = append(args, *req.Title)
		argCount++
	}
	if req.Comment != nil {
		updates = append(updates, fmt.Sprintf("comment = $%d", argCount))
		args = append(args, *req.Comment)
		argCount++
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, reviewID)

	query := fmt.Sprintf(`
		UPDATE reviews SET %s WHERE id = $%d RETURNING *
	`, joinComma(updates), argCount)

	var updated models.Review
	err = tx.Get(&updated, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update review: %w", err)
	}

	if err := recalculateDriverRating(tx, updated.DriverID); err != nil {
		return nil, err
	}
	if err := recalculateDriverRating(tx, updated.DriverID); err != nil {
		return nil, err
	}
	if existing.CarID != nil {
		if err := recalculateCarRating(tx, *existing.CarID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit review update: %w", err)
	}

	return &updated, nil
}

// EditReviewImages — add new photos and/or remove specific existing ones.
func (r *ReviewService) EditReviewImages(reviewID string, requesterID string, isAdmin bool, newImages []string, newImagePublicIDs []string, removePublicIDs []string) (*models.Review, []string, error) {
	if len(newImages) != len(newImagePublicIDs) {
		return nil, nil, errors.New("new images and new_image_public_ids must have the same number of items")
	}

	var review models.Review
	query := `SELECT * FROM reviews WHERE id = $1`
	err := database.DB.Get(&review, query, reviewID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("review not found")
		}
		return nil, nil, fmt.Errorf("database error: %w", err)
	}

	if !isAdmin && review.UserID != requesterID {
		return nil, nil, errors.New("you can only edit your own review")
	}

	var currentImages []string
	var currentPublicIDs []string

	if len(review.Images) > 0 {
		json.Unmarshal([]byte(review.Images), &currentImages)
	}
	if len(review.ImagePublicIDs) > 0 {
		// Fixed — was corrupted with a stray comment fragment mid-token
		// (same corruption appeared in both uploads of this file; worth
		// double-checking whatever you're copying this through).
		json.Unmarshal([]byte(review.ImagePublicIDs), &currentPublicIDs)
	}

	removeSet := make(map[string]bool, len(removePublicIDs))
	for _, id := range removePublicIDs {
		removeSet[id] = true
	}

	filteredImages := make([]string, 0, len(currentImages))
	filteredPublicIDs := make([]string, 0, len(currentPublicIDs))
	var actuallyRemoved []string

	for i, pubID := range currentPublicIDs {
		if removeSet[pubID] {
			actuallyRemoved = append(actuallyRemoved, pubID)
		} else {
			if i < len(currentImages) {
				filteredImages = append(filteredImages, currentImages[i])
			}
			filteredPublicIDs = append(filteredPublicIDs, pubID)
		}
	}

	filteredImages = append(filteredImages, newImages...)
	filteredPublicIDs = append(filteredPublicIDs, newImagePublicIDs...)

	if len(filteredImages) > 3 {
		return nil, nil, fmt.Errorf(
			"review cannot have more than 3 images (currently %d, adding %d)",
			len(currentImages)-len(actuallyRemoved), len(newImages),
		)
	}

	updatedImagesJSON, _ := json.Marshal(filteredImages)
	updatedPublicIDsJSON, _ := json.Marshal(filteredPublicIDs)

	var updatedReview models.Review
	err = database.DB.Get(&updatedReview, `
		UPDATE reviews
		SET images = $1, image_public_ids = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING *
	`, updatedImagesJSON, updatedPublicIDsJSON, reviewID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update review images: %w", err)
	}

	return &updatedReview, actuallyRemoved, nil
}

func (r *ReviewService) GetReviewByID(reviewID string) (*models.Review, error) {
	var review models.Review
	query := `SELECT * FROM reviews WHERE id = $1`
	err := database.DB.Get(&review, query, reviewID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("review not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &review, nil
}


// review_service.go
func (r *ReviewService) GetCarReviews(carID string) ([]*models.Review, error) {
	var reviews []*models.Review
	query := `
		SELECT * FROM reviews
		WHERE car_id = $1 AND status = 'published'
		ORDER BY created_at DESC
	`
	err := database.DB.Select(&reviews, query, carID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch car reviews: %w", err)
	}
	return reviews, nil
}

// --- shared helpers, used by both CreateReview and UpdateReview so the
// recalculation logic exists in exactly one place, not duplicated twice
// with room to drift apart.

func recalculateDriverRating(tx *sqlx.Tx, driverID string) error {
	_, err := tx.Exec(`
		UPDATE drivers
		SET average_rating = (SELECT COALESCE(AVG(rating), 0) FROM reviews WHERE driver_id = $1 AND status = 'published'),
		    total_reviews  = (SELECT COUNT(*) FROM reviews WHERE driver_id = $1 AND status = 'published')
		WHERE id = $1
	`, driverID)
	if err != nil {
		return fmt.Errorf("failed to update driver rating: %w", err)
	}
	return nil
}

func recalculateCarRating(tx *sqlx.Tx, carID string) error {
	_, err := tx.Exec(`
		UPDATE cars
		SET average_rating = (SELECT COALESCE(AVG(vehicle_condition_rating), 0) FROM reviews WHERE car_id = $1 AND status = 'published'),
		    total_reviews  = (SELECT COUNT(*) FROM reviews WHERE car_id = $1 AND status = 'published')
		WHERE id = $1
	`, carID)
	if err != nil {
		return fmt.Errorf("failed to update car rating: %w", err)
	}
	return nil
}

func joinComma(items []string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += item
	}
	return result
}
