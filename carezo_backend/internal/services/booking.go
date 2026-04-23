package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
	"github.com/google/uuid"
)

type BookingService struct{}

func NewBookingService() *BookingService {
	return &BookingService{}
}

const refCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
const refLength = 8

func generateBookingReference() (string, error) {
	b := make([]byte, refLength)

	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(refCharset))))

		if err != nil {
			return "", fmt.Errorf("Failed to generate booking reference: %w", err)
		}
		b[i] = refCharset[n.Int64()]
	}
	return "BK-" + string(b), nil
}

// CreateBooking by user
func (s *BookingService) CreateBooking(userID string, req *models.CreateBookingRequest) (*models.Booking, error) {
	// validate date
	now := time.Now()
	if req.PickupDate.Before(now) {
		return nil, errors.New("Pickup date must be in the future or future date")
	}
	if !req.ReturnDate.After(req.PickupDate) {
		return nil, errors.New("Return date must be after pickup date")
	}
	// fetch car after confirming car exist and pricing
	var car models.Car

	query := `SELECT * FROM cars WHERE id = $1 AND deleted_at IS NULL`

	err := database.DB.Get(&car, query, req.CarID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Car not found")
		}
		return nil, fmt.Errorf("Database error: %w", err)
	}

	if !car.IsAvailable {
		return nil, errors.New("Car is currently marked as unavailable")
	}

	// fetch driver

	var driver models.Driver
	query = `SELECT * FROM drivers WHERE id = $1 AND deleted_at IS NULL`
	err = database.DB.Get(&driver, query, req.DriverID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Driver not found")
		}
		return nil, fmt.Errorf("Database error: %w", err)
	}

	if !driver.IsAvailable {
		return nil, errors.New("Driver is currently maarked as unavailable")
	}

	// check for car overlapping
	var carBookingCount int

	query = `
		SELECT COUNT(*) FROM bookings
		WHERE car_id        = $1
		  AND status        NOT IN("cancelled", "completed")
		  AND pickup_date < $3
		  AND return_date > $2	
	`
	err = database.DB.Get(&carBookingCount, query, req.CarID, req.PickupDate, req.ReturnDate)

	if err != nil {
		return nil, fmt.Errorf("Database error checking car availability: %w", err)
	}
	if carBookingCount > 0 {
		return nil, errors.New("Car is not available for the selected dates")
	}

	// check driver availability
	var driverBookingCount int
	query = `
		SELECT COUNT(*) FROM bookings
		WHERE driver_id        = $1
		  AND status        NOT IN("cancelled", "completed")
		  AND pickup_date < $3
		  AND return_date > $2	
	`
	err = database.DB.Get(&driverBookingCount, query, req.DriverID, req.PickupDate, req.ReturnDate)

	if err != nil {
		return nil, fmt.Errorf("Database error checking driver availability: %w", err)
	}
	if driverBookingCount > 0 {
		return nil, errors.New("Driver is not available for the selected dates")
	}
	var rateToUse float64

	// unmarshal JSONB into a real Go map BEFORE the switch
	var rates map[string]float64
	if err := json.Unmarshal([]byte(car.HourlyRate), &rates); err != nil {
		return nil, fmt.Errorf("failed to parse hourly rate: %w", err)
	}

	// using the map in your switch
	switch time.Now().Weekday() {
	case time.Saturday, time.Sunday:
		// weekend rate
		if val, ok := rates["weekend"]; ok {
			rateToUse = val
		}
	default:
		// weekday rate
		if val, ok := rates["weekday"]; ok {
			rateToUse = val
		}
	}

	// safety check
	if rateToUse == 0 {
		return nil, errors.New("Car hourly rate is not configured correctly")
	}

	if rateToUse == 0 {
		return nil, fmt.Errorf("no rate configured for this day")
	}
	// calculate pricing
	duration := req.ReturnDate.Sub(req.PickupDate)
	totalHours := math.Ceil(duration.Hours()) // partial hours round up
	tripCost := rateToUse * totalHours
	totalAmount := tripCost + car.CautionFee
	refundableAmount := car.CautionFee // caution fee is refundable on completion

	// generate booking reference
	ref, err := generateBookingReference()
	if err != nil {
		return nil, err
	}
	// create booking
	bookingID := uuid.New().String()

	query = `
        INSERT INTO bookings (
            id,
            booking_reference,
            user_id,
            car_id,
            driver_id,
            pickup_date,
            return_date,
            destination,
            pickup_location,
            hourly_rate,
            caution_fee,
            total_amount,
            refundable_amount,
            payment_status,
            status,
            special_requests 
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
        )
        RETURNING *
    `

	var booking models.Booking
	err = database.DB.Get(&booking, query,
		bookingID,
		ref,
		userID,
		req.CarID,
		req.DriverID,
		req.PickupDate,
		req.ReturnDate,
		req.Destination,
		req.PickupLocation,
		rateToUse,
		car.CautionFee,
		totalAmount,
		refundableAmount,
		models.PaymentStatusPending,
		models.BookingStatusPending,
		req.SpecialRequests,
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to create booking: %w", err)
	}

	return &booking, nil
}

func (s *BookingService) GetBookingByID(bookingID string) (*models.Booking, error) {
	var booking models.Booking

	query := `SELECT * FROM bookings WHERE id = $1`
	err := database.DB.Get(&booking, query, bookingID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Booking not found")
		}
		return nil, fmt.Errorf("Databse error: %w", err)
	}
	return &booking, nil
}

// GetUserBooking
func (s *BookingService) GetUserBookings(userID string, status string, page, limit int) ([]models.Booking, int, error) {
	// default pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var conditions []string
	var args []interface{}
	argCount := 1

	conditions = append(conditions, fmt.Sprintf("user_id = $%d", argCount))
	args = append(args, userID)
	argCount++

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, status)
		argCount++
	}

	// ── count total (for pagination meta) ───────────────────────────────
	countQuery := "SELECT COUNT(*) FROM bookings WHERE "
	for i, cond := range conditions {
		if i > 0 {
			countQuery += " AND "
		}
		countQuery += cond
	}

	var total int
	err := database.DB.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("Database error counting bookings: %w", err)
	}

	// ── fetch the page ──────────────────────────────────────────────────
	offset := (page - 1) * limit

	dataQuery := "SELECT * FROM bookings WHERE "
	for i, cond := range conditions {
		if i > 0 {
			dataQuery += " AND "
		}
		dataQuery += cond
	}
	dataQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	var bookings []models.Booking
	err = database.DB.Select(&bookings, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("Database error fetching bookings: %w", err)
	}

	return bookings, total, nil
}

// Cancel Booking
func (s *BookingService) CancelBooking(bookingID string, userID string, reason string) error {
	// fetch the booking

	booking, err := s.GetBookingByID(bookingID)
	if err != nil {
		return err
	}

	// only pending booking or confirmed booking can be cancelled
	if booking.Status != models.BookingStatusPending && booking.Status != models.BookingStatusConfirmed {
		return fmt.Errorf("Only pending or confirmed bookings can be cancelled (current status: %s)", booking.Status)
	}

	// update status and set cancellation
	query := `
		UPDATE bookings
		SET    status              = $1,
		       cancellation_reason = $2
		WHERE  id                  = $3
	`

	result, err := database.DB.Exec(query, models.BookingStatusCancelled, reason, bookingID)

	if err != nil {
		return fmt.Errorf("Failed to cancel booking: %w", err)
	}

	rows, _ := result.RowsAffected()

	if rows == 0 {
		return errors.New("Booking not found or already deleted")
	}

	// new feature:- send cancellation email

	return nil
}

// Called during payment initialization to save the Paystack
// reference string into the booking row so we can find it later.

func (s *BookingService) StorePaymentReference(bookingID string, reference string) error {
	query := `
		UPDATE bookings
		SET	   payment_reference = $1
		WHERE  id 				 = $2	
	`

	result, err := database.DB.Exec(query, reference, bookingID)
	if err != nil {
		return fmt.Errorf("Failed to store payment reference: %w", err)
	}

	rows, _ := result.RowsAffected()

	if rows == 0 {
		return errors.New("Booking not found")
	}
	return nil
}

// Called during payment verification to find the booking
// using the Paystack reference string.

func (s *BookingService) GetBookingByPaymentReference(reference string) (*models.Booking, error) {
	var booking models.Booking

	query := `SELECT * FROM bookings WHERE payment_reference = $1`
	err := database.DB.Get(&booking, query, reference)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Booking not found for this payment reference")
		}
		return nil, fmt.Errorf("Database error: %w", err)
	}
	return &booking, nil
}
