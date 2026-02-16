package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
)

type PaymentService struct {
	paystackSecretKey string
	paystackBaseURL string
	bookingService *BookingService
}


func NewPaymentService(paystackSecretKey string) *PaymentService {
	return &PaymentService{
		paystackSecretKey: paystackSecretKey,
		paystackBaseURL: "https://api.paystack.co",
		bookingService: NewBookingService(),
	}
}

type PaystackInitializeResponse struct {
	Status bool `json:"status"`
	Message string `json:"message"`
	Data struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode string `json:"access_code"`
		Reference string `json:"reference"`
	} `json:"data"`
}


type PaystackVerifyResponse struct {
	Status bool `json:"status"`
	Message string `json:"message"`
	Data struct {
		ID       int64    `json:"id"`
		Domain   string   `json:"domain"`
		Status   string   `json:"status"`
		Reference string  `json:"reference"`
		Amount    int64    `json:"amount"`
		PaidAt    string   `json:"paid_at"`
		Channel   string   `json:"channel"`
		Currency  string   `json:"currency"`
		Customer  struct {
			Email string `json:"email"`
		} `json:"customer"`
	} `json:"data"`
}


func (s *PaymentService) InitializePayment(bookingID string, userEmail string)(*models.PaymentInitializeResponse, error){
	var booking models.Booking

	query := `SELECT * FROM bookings WHERE id =$1`

	err := database.DB.Get(&booking, query, bookingID)

	if err != nil {
		if err == sql.ErrNoRows{
			return nil, errors.New("Booking not found")
		}
		return nil, fmt.Errorf("Database error: %w", err)
	}

	// check if payment already exist
	if booking.PaymentStatus == models.PaymentStatusCompleted {
		return nil, errors.New("Payment already completed for this booking")
	}


	// convert amount from kobo to naira
	// because paystack charges in kobo
	amountInKobo := int(booking.TotalAmount * 100)

	// request payload
	payload := map[string]interface{} {
		"email": userEmail,
		"amount": amountInKobo,
		"metadata": map[string]string {
			"booking_id": bookingID,
		},
	}

	payloadBytes, err := json.Marshal(payload)

	if err != nil {
		return nil, fmt.Errorf("Failed to marshal payment payload: %w", err)
	}

	// calling paystack api
	url := s.paystackBaseURL + "/transaction/initialize"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))

	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.paystackSecretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to initialize payment: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("Failed to read response: %w", err)
	}

	var paystackResp PaystackInitializeResponse

	if err := json.Unmarshal(body, &paystackResp);  err != nil {
		return nil, fmt.Errorf("Failed to parse Paystack response: %w", err)
	}

	if !paystackResp.Status {
		return nil, fmt.Errorf("Paystack error: %s", paystackResp.Message)
	}


	// store payment reference in the booking so that it can be used during verification
	err = s.bookingService.StorePaymentReference(bookingID, paystackResp.Data.Reference)

	if err != nil {
		return nil, fmt.Errorf("Failed to store payment reference: %w", err)
	}


	// return payment url and reference

	return &models.PaymentInitializeResponse{	
		AuthorizationURL: paystackResp.Data.AuthorizationURL,
		AccessCode: paystackResp.Data.AccessCode,
		Reference: paystackResp.Data.Reference,
	}, nil
}
	

// This is to verify payment
func (s *PaymentService) VerifyPayment(reference string) error {
	// call paystack verify endpoint
	url := s.paystackBaseURL + "/transaction/verify/" + reference

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return fmt.Errorf("Failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.paystackSecretKey)
	client := &http.Client{Timeout: 30 * time.Second}

	response, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("Failed to verify payment: %w", err)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return fmt.Errorf("Failed to read response: %w", err)
	}

	var paystackResp PaystackVerifyResponse

	if err := json.Unmarshal(body, &paystackResp); err != nil {
		return fmt.Errorf("Failed to parse Paystack response: %w", err)
	}

	if !paystackResp.Status {
		return fmt.Errorf("Paystack error: %w", paystackResp.Message)
	}

	// check if payment was successfull
	if paystackResp.Data.Status != "success" {
		return fmt.Errorf("Payment wasnt successfull. Status of payment: %w", paystackResp.Data.Status)
	}

	// find the booking
	booking, err := s.bookingService.GetBookingByPaymentReference(reference)

	if err != nil {
		return err
	}

	// verify that amount matches
	expectedAmountInKobo := int64(math.Round(booking.TotalAmount * 100))

	if paystackResp.Data.Amount != expectedAmountInKobo {
		return fmt.Errorf("Payment amount mismatch. Expected: %d kobo, Got: %d kobo", expectedAmountInKobo, paystackResp.Data.Amount)
	}

	// parse paid_at timestamp
	paidAt, err := time.Parse(time.RFC3339, paystackResp.Data.PaidAt)

	if err != nil {
		// use current time if parsing fails
		paidAt = time.Now()
	}

	// Update booking payment status\
	updateQuery := `
		UPDATE  bookings
		SET     payment_Status   = $1,
				paid_at			 = $2
				status			 = CASE
									  WHEN status = "pending" THEN "confirmed"
									  ELSE status
									END
		WHERE id 				=  $3
	`

		result, err := database.DB.Exec(updateQuery,
		models.PaymentStatusPaid,
		paidAt,
		booking.ID,

	)

	if err != nil {
		return fmt.Errorf("Failed to update booking payment status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return  errors.New("Booking not found")
	}
	return nil

}
