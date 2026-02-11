package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

	var paystackRep PaystackInitializeResponse

	if err := json.Unmarshal(body, &paystackRep);  err != nil {
		return nil, fmt.Errorf("Failed to parse Paystack response: %w", err)
	}

	if !paystackRep.Status {
		return nil, fmt.Errorf("Paystack error: %s", paystackRep.Message)
	}


	// store payment reference in the booking so that it can be used during verification
	err = s.bookingService.StorePaymentReference(bookingID, paystackRep.Data.Reference)

	if err != nil {
		return nil, fmt.Errorf("Failed to store payment reference: %w", err)
	}


	// return payment url and reference

	return &models.PaymentInitializeResponse{	
		AuthorizationURL: paystackRep.Data.AuthorizationURL,
		AccessCode: paystackRep.Data.AccessCode,
		Reference: paystackRep.Data.Reference,
	}, nil
}
	
