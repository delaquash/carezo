package handlers_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/delaquash/carezo/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRebookAfterCancellation(t *testing.T) {
	// set up the test router and DB
	app := testhelpers.SetupTestApp(t)

	// clean up after this test finishes so other tests start fresh
	t.Cleanup(func() { app.CleanupDB(t) })

	// ── Step 1: Seed test data directly into the test DB ──────────
	// We insert a test user, car, and driver so we have IDs to work with
	// This is faster than calling the API for each piece of setup data

	userID := "11111111-1111-1111-1111-111111111111"
	carID := "22222222-2222-2222-2222-222222222222"
	driverID := "33333333-3333-3333-3333-333333333333"

	_, err := app.DB.Exec(`
        INSERT INTO users (id, email, password_hash, first_name, last_name,
            role, status, email_verified)
        VALUES ($1, 'test@carezo.com', 'hash', 'Test', 'User', 'user', 'active', true)
        ON CONFLICT (id) DO NOTHING
    `, userID)
	require.NoError(t, err, "failed to seed test user")

	_, err = app.DB.Exec(`
        INSERT INTO cars (id, model, year, color, license_plate,
            transmission, fuel_type, seating_capacity,
            hourly_rate, caution_fee, is_available, status)
        VALUES ($1, 'Camry', 2022, 'Black', 'TEST-001',
            'automatic', 'petrol', 5,
            '{"weekday":5000,"weekend":7000,"holiday":10000}'::jsonb,
            50000, true, 'active')
        ON CONFLICT (id) DO NOTHING
    `, carID)
	require.NoError(t, err, "failed to seed test car")

	_, err = app.DB.Exec(`
        INSERT INTO drivers (id, first_name, last_name, age, gender,
            license_number, license_expiry_date, years_of_experience,
            is_available, status)
        VALUES ($1, 'John', 'Doe', 35, 'male',
            'TEST-LIC-001', '2030-01-01', 5, true, 'active')
        ON CONFLICT (id) DO NOTHING
    `, driverID)
	require.NoError(t, err, "failed to seed test driver")

	// generate a JWT for our test user without calling /api/auth/login
	userToken := testhelpers.GenerateTestToken(userID, "user", app.Config.JWTSecret)

	// ── Step 2: Book the car for July 20-25 ───────────────────────
	bookingBody := map[string]interface{}{
		"car_id":          carID,
		"driver_id":       driverID,
		"pickup_date":     "2026-07-20T09:00:00Z",
		"return_date":     "2026-07-25T09:00:00Z",
		"pickup_location": "Lagos Mainland",
		"destination":     "Lekki Phase 1",
	}

	w := app.MakeRequest("POST", "/api/bookings", bookingBody, userToken)

	// confirm the booking was created
	assert.Equal(t, http.StatusCreated, w.Code,
		"expected 201 when creating booking, got: %s", w.Body.String())

	// extract the bookingID from the response to use in the cancel step
	resp := testhelpers.ParseResponse(w)
	data := resp["data"].(map[string]interface{})
	bookingID := data["id"].(string)
	assert.NotEmpty(t, bookingID, "booking ID should not be empty")

	t.Logf("✅ Booking created: %s", bookingID)

	// ── Step 3: Cancel that booking ───────────────────────────────
	cancelBody := map[string]interface{}{
		"reason": "Changed my mind",
	}

	cancelURL := fmt.Sprintf("/api/bookings/%s/cancel", bookingID)
	w = app.MakeRequest("POST", cancelURL, cancelBody, userToken)

	assert.Equal(t, http.StatusOK, w.Code,
		"expected 200 when cancelling booking, got: %s", w.Body.String())

	// verify the booking status is now cancelled in the DB
	var status string
	err = app.DB.Get(&status, "SELECT status FROM bookings WHERE id = $1", bookingID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", status, "booking status should be 'cancelled' after cancel")

	t.Logf("✅ Booking cancelled: %s", bookingID)

	// ── Step 4: Try to rebook for the SAME dates ──────────────────
	// THIS IS THE KEY TEST — before the fix this would return 400
	// because cancelled bookings were blocking the dates
	w = app.MakeRequest("POST", "/api/bookings", bookingBody, userToken)

	// after the fix this MUST return 201 Created
	assert.Equal(t, http.StatusCreated, w.Code,
		"should be able to rebook same dates after cancellation, got: %s", w.Body.String())

	resp2 := testhelpers.ParseResponse(w)
	data2 := resp2["data"].(map[string]interface{})
	newBookingID := data2["id"].(string)
	assert.NotEmpty(t, newBookingID)
	assert.NotEqual(t, bookingID, newBookingID, "should be a new booking ID")

	t.Logf("Rebook succeeded: %s (new ID)", newBookingID)
}

func TestCannotDoubleBookSameDates(t *testing.T) {
	// Sibling test — confirms that WITHOUT cancellation, double booking is blocked
	app := testhelpers.SetupTestApp(t)
	t.Cleanup(func() { app.CleanupDB(t) })

	// seed + token (same as above — could extract to a helper function)
	userID := "44444444-4444-4444-4444-444444444444"
	carID := "55555555-5555-5555-5555-555555555555"
	driverID := "66666666-6666-6666-6666-666666666666"

	app.DB.Exec(`INSERT INTO users VALUES ($1, 'user2@test.com', 'h', 'A', 'B', 'user', 'active', true) ON CONFLICT DO NOTHING`, userID)
	app.DB.Exec(`INSERT INTO cars (id, model, year, color, license_plate, transmission, fuel_type, seating_capacity, hourly_rate, caution_fee, is_available, status) VALUES ($1, 'Corolla', 2021, 'White', 'TEST-002', 'automatic', 'petrol', 5, '{"weekday":5000,"weekend":7000,"holiday":10000}'::jsonb, 50000, true, 'active') ON CONFLICT DO NOTHING`, carID)
	app.DB.Exec(`INSERT INTO drivers (id, first_name, last_name, age, gender, license_number, license_expiry_date, years_of_experience, is_available, status) VALUES ($1, 'Jane', 'Doe', 30, 'female', 'TEST-LIC-002', '2030-01-01', 3, true, 'active') ON CONFLICT DO NOTHING`, driverID)

	token := testhelpers.GenerateTestToken(userID, "user", app.Config.JWTSecret)

	body := map[string]interface{}{
		"car_id": carID, "driver_id": driverID,
		"pickup_date":     "2026-08-01T09:00:00Z",
		"return_date":     "2026-08-05T09:00:00Z",
		"pickup_location": "VI", "destination": "Ajah",
	}

	// first booking should succeed
	w := app.MakeRequest("POST", "/api/bookings", body, token)
	assert.Equal(t, http.StatusCreated, w.Code, "first booking should succeed")

	// second booking for same car/dates should FAIL
	w = app.MakeRequest("POST", "/api/bookings", body, token)
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"double booking same dates should fail")

	t.Logf("Double booking correctly blocked")
}
