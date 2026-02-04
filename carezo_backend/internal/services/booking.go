package services

import (
	"crypto/rand"
	"database/sql"
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

func NewBookingService() * BookingService {
	return  &BookingService{}
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
	return "BK-" +string(b), nil
}


// CreateBooking
func (s *BookingService) CreateBooking(userID string, req *models.CreateBookingRequest)(*models.Booking, error) {
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

	// calculate pricing
	duration      := req.ReturnDate.Sub(req.PickupDate)
	totalHours    := math.Ceil(duration.Hours())
	tripCost      := car.HourlyRate  * totalHours
	totalAmount   := tripCost + car.CautionFee

	// generate booking reference
	ref, err := generateBookingReference()
	if err != nil {
		return nil, err
	}

		// NOTE: total_hours is a GENERATED column in postgres – we do NOT insert it.
		// We use RETURNING to get the full row back (including the computed total_hours).

		bookingID := uuid.New().String()

		query = `
			INSERT INTO bookings (
				id,
				booking_reference
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
			 $1, $2, $3, $4, $5,$6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
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
		car.HourlyRate,
		car.CautionFee,
		totalAmount,
		refundableAmount,
		models.PaymentStatusPending,
		models.BookingStatusPending,
		req.SpecialRequests,
		) 

}