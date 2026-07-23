package models

import "time"

const (
	NotificationTypeBookingCreated       = "booking_created"
	NotificationTypePaymentSuccess       = "payment_success"
	NotificationTypeBookingCancelled     = "booking_cancelled"
	NotificationtypePickup               = "vehicle_picked_up"
	NotificationTypeDropoff              = "vehicle_dropped_off"
	NotificationTypeReturned             = "vehicle_returned"
	NotificationTypeLateCarReturnWarning = "late_return_warning"
	NotificationTypeLateCarReturn        = "late_return"
)

type Notification struct {
	ID      string `json:"id" db:"id"`
	UserID  string `json:"user_id" db:"user_id"`
	Title   string `json:"title" db:"title"`
	Message string `json:"message" db:"message"`
	Type    string `json:"type" db:"type"`
	Data    JSONB  `json:"data" db:"data"` // booking details data
	IsRead  bool   `json:"is_read" db:"is_read"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateNotificationRequest struct {
	UserID  string                 `json:"user_id"`
	Title   string                 `json:"title"`
	Message string                 `json:"message"`
	Type    string                 `json:"type"`
	Data    map[string]interface{} `json:"data"`
}
