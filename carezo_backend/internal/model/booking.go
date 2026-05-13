package models

import (
	"time"

	"github.com/google/uuid"
)


const (
	BookingStatusPending  = "pending"
	BookingStatusConfirmed = "confirmed"
	BookingStatusCancelled = "cancelled"
	BookingStatusCompleted = "completed"
	BookingStatusActive = "active"

	PaymentStatusPending  = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
	PaymentStatusPaid = "paid"
)

type Booking struct {
	ID        string `json:"id" db:"id"`
	BookingReference string `json:"booking_reference" db:"booking_reference"`
	// foreign keys
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
    CarID    uuid.UUID `json:"car_id" db:"car_id"`
	DriverID uuid.UUID `json:"driver_id" db:"driver_id"`


	// Booking details
	PickUpDate   time.Time `json:"pickup_date" db:"pickup_date"`
	ReturnDate   time.Time `json:"return_date" db:"return_date"`
	ActualReturnDate time.Time `json:"actual_return_date,omitempty" db:"actual_return_date"`

	Destination string `json:"destination" db:"destination"`

	PickUpLocation *string  `json:"pickup_location" db:"pickup_location"`

	HourlyRate    float64 `json:"hourly_rate" db:"hourly_rate"`
	CautionFee    float64 `json:"caution_fee" db:"caution_fee"`
	TotalAmount   float64 `json:"total_amount" db:"total_amount"`

	RefundableAmount float64 `json:"refundable_amount" db:"refundable_amount"`

	// payment
	PaymentStatus string `json:"payment_status" db:"payment_status"`
	PaymentReference *string `json:"payment_reference" db:"payment_reference"`
	PaidAt   *time.Time `json:"paid_at," db:"paid_at"`
	RefundedAt  *time.Time `json:"refunded_at" db:"refunded_at"`

	Status string `json:"status" db:"status"`

	CancellationReason *string `json:"cancellation_reason,omitempty" db:"cancellation_reason"`
	SpecialRequests    *string `json:"special_requests,omitempty" db:"special_requests"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}


type CreateBookingRequest struct {
	CarID           uuid.UUID `json:"car_id"            binding:"required"`
	DriverID        uuid.UUID `json:"driver_id"         binding:"required"`
	PickupDate      time.Time `json:"pickup_date"       binding:"required"`
	ReturnDate      time.Time `json:"return_date"       binding:"required"`
	Destination     string    `json:"destination"       binding:"required"`
	PickupLocation  *string   `json:"pickup_location"`  // optional
	SpecialRequests *string   `json:"special_requests"` // optional
}

type CancelBookingRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type ListBookingsRequest struct {
	Status string `form:"status"`  // optional – filter by booking status
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

type BookingResponse struct {
	Booking
	Car    *CarSummary    `json:"car,omitempty"`
	Driver *DriverSummary `json:"driver,omitempty"`
}

type CarSummary struct {
	ID           uuid.UUID `json:"id"`
	Model        string    `json:"model"`
	Brand        string    `json:"brand"`
	Color        string    `json:"color"`
	LicensePlate string    `json:"license_plate"`
}

type DriverSummary struct {
	ID              uuid.UUID `json:"id"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Gender          string    `json:"gender"`
	AverageRating   float64   `json:"average_rating"`
	ProfileImageURL *string   `json:"profile_image_url,omitempty"`
}

type PaginatedBookingsResponse struct {
	Bookings []Booking `json:"bookings"`
	// Meta     PaginationMeta `json:"meta"`
}

// type PaginationMeta struct {
// 	Total      int `json:"total"`
// 	Page       int `json:"page"`
// 	Limit      int `json:"limit"`
// 	TotalPages int `json:"total_pages"`
// }
