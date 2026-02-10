package models

import "time"


type PaymentInitializeRequest struct {
	BookingID string `json:"booking_id" binding:"required"`
}

type PaymentInitializeResponse struct {
	PaymentURL string `json:"payment_url"`
	AccessCode string `json:"access_code"`
	Reference  string `json:"reference"`
}


type PaymentVerifyRequeest struct {
	Reference string `json:"reference" binding:"required"`
}

type PaymentVerifyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	BookingID string `json:"booking_id"`
	Amount float64 `json:"amount"`
	PaidAt time.Time `json:"paid_at"`
}