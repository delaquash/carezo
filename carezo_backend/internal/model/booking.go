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
)

type Booking struct {
	ID        string `json:"id" db:"id"`

	// foreign keys
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
    CarID    uuid.UUID `json:"car_id" db:"car_id"`
	DriverID uuid.UUID `json:"driver_id" db:"driver_id"`


	// Booking details
	PickUpDate   time.Time `json:"pickup_date" db:"pickup_date"`
	ReturnDate   time.Time `json:"return_date" db:"return_date"`
	ActualReturnDate time.Time `json:"actual_return_date,omitempty" db:"actual_return_date"`


	TotalHours  int `json:"total_hours" db:"total_hours"`
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

	CancellationReason *string `json:"cancellation_reason" db:"cancellation_reason"`
	SpecialRequests    *string `json:"special_requests" db:"special_requests"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}