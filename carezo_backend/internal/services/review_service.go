package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
	"github.com/google/uuid"
)


type ReviewService struct{} // stateless — no fields needed, all data comes from the DB each call

func NewReviewService() *ReviewService {
	return &ReviewService{}
}

// User create review after booking
func (r *ReviewService)  CreateReview(userID string, req *models.CreateReviewRequest)(*models.Review, error) {
	// images and public images must be parallet
	if len(req.Images) != len(req.ImagePublicIDs) {
		return nil, errors.New("images and images_public_ids must have the same number of items")
	}

	// enforce limit of 3 photos for reviews
	if len(req.Images) > 3 {
		return nil, errors.New("minimum of 3 images allowed per review")
	}
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
		return nil, fmt.Errorf("Database error: %w", err) // %w preserves the original error for unwrapping
	}

	// check if booking is completed and only completed trips can be reviewed — stops reviews before the ride happened
	if booking.Status != "completed" {
		return nil, errors.New("Can only review completed booking")
	}

	// check if review is exist for this booking before to avoid duplicate review
	var exists bool
	query = `SELECT EXISTS(SELECT 1 FROM reviews WHERE booking_id = $1)`
	err = database.DB.Get(&exists, query, req.BookingID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if exists {
		return nil, errors.New("you have already reviewed this booking")
	}
	// convert the GO []string of image URLs into JSOn bytes for the JSONB column
	imagesJSON, err := json.Marshal(req.Images)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal image publics IDs: %w", err)
	}

	// same conversion for the public_ids array — stored alongside images
	imagePublicIDsJSON, err := json.Marshal(req.ImagePublicIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal image public IDs: %w", err)
	}

	// create review
	reviewID := uuid.New().String()  // generate a fresh UUID for the new review row
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

	var review models.Review  // will be filled by RETURNING * below

	// INSERT the review and immediately get the full row back via RETURNING —
	// columns are listed explicitly (not RETURNING *) so a schema mismatch
	// doesn't silently break the scan
	err = database.DB.Get(&review, `
		INSERT INTO reviews (
			id, booking_id, user_id, driver_id,
			rating, punctuality_rating, professionalism_rating,
			vehicle_condition_rating, title, comment,
			images, image_public_ids, status
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9, $10,
			$11, $12, 'published'
		)
		RETURNING
			id, booking_id, user_id, driver_id,
			rating, punctuality_rating, professionalism_rating,
			vehicle_condition_rating, title, comment,
			images, image_public_ids, status,
			created_at, updated_at
	`,
		reviewID, req.BookingID, userID, booking.DriverID,
		req.Rating, req.PunctualityRating, req.ProfessionalismRating,
		req.VehicleConditionRating, req.Title, req.Comment,
		imagesJSON, imagePublicIDsJSON, // the two JSONB arrays go in last
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	return &review, nil // hand back the fully-populated review struct
}

// EditReviewImages — add new photos and/or remove specific existing ones in one call
// returns the updated review AND the list of public_ids that were removed,
// so the HANDLER (not this service) can delete them from Cloudinary
func (r *ReviewService) EditReviewImages (reviewID string, requesterID string, isAdmin bool, newImages []string,newImagePublicIDs []string) (*models.Review, []string, error) {
	// same parallel-array guard as everywhere else in this codebase
	if len(newImages) != len(newImagePublicIDs) {
		return nil, nil, errors.New("new images and new_image_public_ids must have the same number of items")
	}

	// load the existing review row so we know its current images and owner
	var review models.Review
	query := `SELECT * FROM reviews WHERE id = $1`
	err := database.DB.Get(&review, query, reviewID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("review not found")
		}
		return nil, nil, fmt.Errorf("database error: %w", err)
	}

	// authorization: only the original authir or an admin can edit images
	if !isAdmin && review.UserID != requesterID {
		return nil, nil, errors.New("you can only edit your own review")
	}

	var currentImages []string    // will hold the review's current photo URLs
	var currentPublicIDs []string // and their matching public_ids

	// only unmarshal if the column actually has data — avoids errors on empty JSONB
	if len(review.Images) > 0 {
		json.Unmarshal([]byte(review.Images),  &currentImages)
	}
	if len(review.ImagePublicIDs) > 0 {
		json.Unmarshal([]byte(review.ImagePublicIDs), &currentPublicIDs)
	}

	// build a lookup SET of public_ids to remove — map gives O(1) "is this in
	// the remove list?" checks instead of looping through removePublicIDs every time
	removeSet := make(map[string]bool, len(removePublicIDs))
	for _, id := range removePublicIDs {
		removeSet[id] = true
	}

	filteredImages := make([]string, 0, len(currentImages))       // survives after filtering out removed ones
	filteredPublicIDs := make([]string, 0, len(currentPublicIDs))
	var actuallyRemoved []string // tracks what we genuinely deleted, for the caller to clean up in Cloudinary
	// walk through current images, keep the ones NOT marked for removal
	for i, pubID := range currentPublicIDs {
		if removeSet[pubID] {
			actuallyRemoved = append(actuallyRemoved, pubID) // mark this one as gone
		} else {
			if i < len(currentImages) { // defensive bounds check in case arrays got out of sync somehow
				filteredImages = append(filteredImages, currentImages[i])
			}
			filteredPublicIDs = append(filteredPublicIDs, pubID)
		}
	}

	// now append the brand new images onto whatever survived the filter
	filteredImages = append(filteredImages, newImages...)
	filteredPublicIDs = append(filteredPublicIDs, newImagePublicIDs...)

	// enforce the 3-image cap AFTER combining old + new — this is the real check
	if len(filteredImages) > 3 {
		return nil, nil, fmt.Errorf(
			"review cannot have more than 3 images (currently %d, adding %d)",
			len(currentImages)-len(actuallyRemoved), len(newImages),
		)
	}

	// convert the final arrays back to JSON bytes for storage
	updatedImagesJSON, _ := json.Marshal(filteredImages)
	updatedPublicIDsJSON, _ := json.Marshal(filteredPublicIDs)

	var updatedReview models.Review
	err = database.DB.Get(&updatedReview, `
		UPDATE reviews
		SET images           = $1,
		    image_public_ids = $2,
		    updated_at       = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING
			id, booking_id, user_id, driver_id,
			rating, title, comment,
			images, image_public_ids, status,
			created_at, updated_at
	`, updatedImagesJSON, updatedPublicIDsJSON, reviewID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update review images: %w", err)
	}

	// hand back BOTH the new review state AND what was removed —
	// the handler uses the second value to clean up Cloudinary
	return &updatedReview, actuallyRemoved, nil

}
// Get single review by ID
func (r *ReviewService) GetReviewByID(reviewID string)(*models.Review, error) {
	var review models.Review
	query := `SELECT * FROM reviews WHERE id = $1`
	err := database.DB.Get(&review,query, reviewID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("review not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &review, nil
}