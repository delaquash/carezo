package models

import (
	"time"
)

type Review struct {
	ID        string `json:"id" db:"id"`
	BookingID string `json:"booking_id" db:"booking_id"`
	UserID    string `json:"user_id" db:"user_id"`
	DriverID  string `json:"driver_id" db:"driver_id"`

	Rating int `json:"rating" db:"rating"`

	PunctualityRating      *int    `json:"punctuality_rating,omitempty" db:"punctuality_rating"`
	ProfessionalismRating  *int    `json:"professionalism_rating,omitempty" db:"professionalism_rating"`
	VehicleConditionRating *int    `json:"vehicle_condition_rating,omitempty" db:"vehicle_condition_rating"`
	Images                 JSONB   `json:"images" db:"images"`
	ImagePublicIDs         JSONB   `json:"image_public_ids" db:"image_public_ids"`
	Title                  *string `json:"title,omitempty" db:"title"`
	Comment                *string `json:"comment,omitempty" db:"comment"`

	Status string `json:"status" db:"status"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateReviewRequest struct {
	BookingID              string  `json:"booking_id" binding:"required"`
	Rating                 int     `json:"rating" binding:"required,min=1,max=5"`
	PunctualityRating      *int    `json:"punctuality_rating,omitempty"`
	ProfessionalismRating  *int    `json:"professionalism_rating,omitempty"`
	VehicleConditionRating *int    `json:"vehicle_condition_rating,omitempty"`
	Title                  *string `json:"title,omitempty"`
	Comment                *string `json:"comment,omitempty"`
	Images                 JSONB   `json:"images" db:"images"`
	ImagePublicIDs []string `json:"image_public_ids,omitempty"`
}

type GetReview struct {
	
}
