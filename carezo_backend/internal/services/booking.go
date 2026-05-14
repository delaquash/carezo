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
	// now := time.Now()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if req.PickupDate.UTC().Before(today) {
		return nil, errors.New("Pickup date cannot be in the past")
	}

	if !req.ReturnDate.After(req.PickupDate) {
		return nil, errors.New("Return date must be after pickup date")
	}

	// wrap everything in DB transaction

	tx, err := database.DB.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	// roll back on error before commit()

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	// fetch car after confirming car exist and pricing
	var car models.Car

	err =tx.Get(&car, `SELECT * FROM cars WHERE id = $1 AND deleted_at IS NULL`, req.CarID)

	// err := database.DB.Get(&car, query, req.CarID)

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
	err =tx.Get(&driver, `SELECT * FROM drivers WHERE id = $1 AND deleted_at IS NULL`, string, req.DriverID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("driver not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if !driver.IsAvailable {
		return nil, errors.New("driver is currently unavailable")
	}

	// check for car overlapping
	var carBookingCount int

	err =tx.Get(&carBookingCount, `
		SELECT COUNT(*) FROM bookings
		WHERE car_id        = $1
		  AND status        NOT IN('cancelled', 'completed')
		  AND pickup_date < $3
		  AND return_date > $2	
	`, req.CarID, req.PickupDate, req.ReturnDate)

	if err != nil {
		return nil, fmt.Errorf("error checking car availability: %w", err)
	}
	if carBookingCount > 0 {
		return nil, errors.New("car is not available for the selected dates")
	}

	// check driver availability
	var driverBookingCount int
	err =tx.Get(&driverBookingCount, `
		SELECT COUNT(*) FROM bookings
		WHERE driver_id        = $1
		  AND status        NOT IN('cancelled', 'completed')
		  AND pickup_date < $3
		  AND return_date > $2	
	`, req.DriverID, req.PickupDate, req.ReturnDate)

	if err != nil {
		return nil, fmt.Errorf("error checking driver availability: %w", err)
	}
	if driverBookingCount > 0 {
		return nil, errors.New("Driver is not available for the selected dates")
	}

	// calculate rate

	// unmarshal JSONB into a real Go map BEFORE the switch
	var rates map[string]float64
	if err = json.Unmarshal([]byte(car.HourlyRate), &rates); err != nil {
		return nil, fmt.Errorf("failed to parse hourly rate: %w", err)
	}

	// using the map in your switch
	var rateToUse float64
	switch req.PickupDate.Weekday() {
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
		return nil, errors.New("car hourly rate is not configured correctly")
	}

	// calculate pricing
	duration := req.ReturnDate.Sub(req.PickupDate)
	totalHours := math.Ceil(duration.Hours())
	tripCost := rateToUse * totalHours
	totalAmount := tripCost + car.CautionFee
	refundableAmount := car.CautionFee 


	// generate booking reference
	ref, err := generateBookingReference()
	if err != nil {
		return nil, err
	}
	// create and insert booking
	bookingID := uuid.New().String()
	var booking models.Booking

	err =tx.Get(&booking, `
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
        RETURNING
		id,
			booking_reference,
			user_id,
			car_id,
			driver_id,
			pickup_date,
			return_date,
			actual_return_date,
			destination,
			pickup_location,
			hourly_rate,
			caution_fee,
			total_amount,
			refundable_amount,
			payment_status,
			payment_reference,
			paid_at,
			refunded_at,
			status,
			cancellation_reason,
			special_requests,
			created_at,
			updated_at
    `,
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
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	// commit, nothing is written to the db until commit()
	// if Commit() fails, defer Rollback() above cleans everything up
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &booking, nil
}

func (s *BookingService) GetBookingByID(bookingID string) (*models.Booking, error) {
	var booking models.Booking

	err := database.DB.Get(&booking, `SELECT * FROM bookings WHERE id = $1`, bookingID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("booking not found")
		}
		return nil, fmt.Errorf("databse error: %w", err)
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

	whereClause := conditions[0]
	for i := 1; i < len(conditions); i++ {
		whereClause += " AND " + conditions[i]
	}
	// ── count total (for pagination meta) ───────────────────────────────
	
	var total int
	if err := database.DB.Get(&total, "SELECT COUNT(*) FROM bookings WHERE "+whereClause, args...); err != nil {
		return nil, 0, fmt.Errorf(" error counting bookings: %w", err)
	}


	// ── fetch the page ──────────────────────────────────────────────────
	offset := (page - 1) * limit

	dataQuery := fmt.Sprintf(
		"SELECT * FROM bookings WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		whereClause, argCount, argCount+1,
	)
	
	args = append(args, limit, offset)

	var bookings []models.Booking
	if err := database.DB.Select(&bookings, dataQuery, args...); err != nil {
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
		return fmt.Errorf("only pending or confirmed bookings can be cancelled (current status: %s)", booking.Status)
	}

	// update status and set cancellation
	result, err := database.DB.Exec( `
		UPDATE bookings
		SET    status              = $1,
		       cancellation_reason = $2
		WHERE  id                  = $3
	`, models.BookingStatusCancelled, reason, bookingID)

	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
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
	result, err := database.DB.Exec(`
		UPDATE bookings
		SET	   payment_reference = $1
		WHERE  id 				 = $2	
	`, reference, bookingID)
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
