package handlers_test

import (
	"net/http"
	"testing"

	"github.com/delaquash/carezo/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRebookAfterCancellation(t *testing.T) {
    app := testhelpers.SetUpTestApp(t)
    t.Cleanup(func() { app.CleanUpDB(t) })

    // seed car and driver
    carID    := "22222222-2222-2222-2222-222222222222"
    driverID := "33333333-3333-3333-3333-333333333333"

    _, err := app.DB.Exec(`
        INSERT INTO cars (id, model, year, color, license_plate,
            transmission, fuel_type, seating_capacity,
            hourly_rate, caution_fee, is_available, status)
        VALUES ($1, 'Camry', 2022, 'Black', 'TEST-001',
            'automatic', 'petrol', 5,
            '{"weekday":5000,"weekend":7000,"holiday":10000}'::jsonb,
            50000, true, 'active')
        ON CONFLICT (id) DO NOTHING
    `, carID)
    require.NoError(t, err)

    _, err = app.DB.Exec(`
        INSERT INTO drivers (
            id, first_name, last_name, age, gender, state,
            phone_number, email, license_number, license_expiry_date,
            years_of_experience, height, is_available, status
        ) VALUES (
            $1, 'John', 'Doe', 35, 'male', 'Lagos',
            '+2348012345678', 'john.doe@test.com', 'TEST-LIC-001', '2030-01-01',
            5, 170, true, 'active'
        )
        ON CONFLICT (id) DO NOTHING
    `, driverID)
    require.NoError(t, err)

    // get a REAL token by logging in — no manual token generation
    userToken := app.LoginTestUser(t, "rebook_test@carezo.com", "Test123!@#")

    bookingBody := map[string]interface{}{
        "car_id":          carID,
        "driver_id":       driverID,
        "pickup_date":     "2026-07-20T09:00:00Z",
        "return_date":     "2026-07-25T09:00:00Z",
        "pickup_location": "Lagos Mainland",
        "destination":     "Lekki Phase 1",
    }

    // Step 1 — create booking
    w := app.MakeRequest("POST", "/api/bookings", bookingBody, userToken)
    assert.Equal(t, http.StatusCreated, w.Code,
        "booking creation failed: %s", w.Body.String())

    // safe extraction — no panic
    resp := testhelpers.ParseResponse(w)
    require.NotNil(t, resp["data"], "response data should not be nil")
    data := resp["data"].(map[string]interface{})
    bookingID := data["id"].(string)
    require.NotEmpty(t, bookingID)
    t.Logf("Booking created: %s", bookingID)

    // Step 2 — cancel it
    cancelURL := "/api/bookings/" + bookingID + "/cancel"
    w = app.MakeRequest("POST", cancelURL,
        map[string]interface{}{"reason": "Changed my mind"}, userToken)
    assert.Equal(t, http.StatusOK, w.Code,
        "cancel failed: %s", w.Body.String())
    t.Logf("Booking cancelled")

    // Step 3 — rebook same dates — must succeed
    w = app.MakeRequest("POST", "/api/bookings", bookingBody, userToken)
    assert.Equal(t, http.StatusCreated, w.Code,
        "rebook after cancel failed: %s", w.Body.String())

    resp2 := testhelpers.ParseResponse(w)
    require.NotNil(t, resp2["data"], "rebook response data should not be nil")
    data2 := resp2["data"].(map[string]interface{})
    newBookingID := data2["id"].(string)
    assert.NotEqual(t, bookingID, newBookingID, "should be a different booking ID")
    t.Logf("Rebook succeeded: %s", newBookingID)
}
func TestCannotDoubleBookSameDates(t *testing.T) {
	// Sibling test — confirms that WITHOUT cancellation, double booking is blocked
	app := testhelpers.SetUpTestApp(t)
	t.Cleanup(func() { app.CleanUpDB(t) })

	// seed + token (same as above — could extract to a helper function)
	userID := "44444444-4444-4444-4444-444444444444"
	carID := "55555555-5555-5555-5555-555555555555"
	driverID := "66666666-6666-6666-6666-666666666666"

	app.DB.Exec(`INSERT INTO users VALUES ($1, 'user2@test.com', 'h', 'A', 'B', 'user', 'active', true) ON CONFLICT DO NOTHING`, userID)
	app.DB.Exec(`INSERT INTO cars (id, model, year, color, license_plate, transmission, fuel_type, seating_capacity, hourly_rate, caution_fee, is_available, status) VALUES ($1, 'Corolla', 2021, 'White', 'TEST-002', 'automatic', 'petrol', 5, '{"weekday":5000,"weekend":7000,"holiday":10000}'::jsonb, 50000, true, 'active') ON CONFLICT DO NOTHING`, carID)
	app.DB.Exec(`INSERT INTO drivers (id, first_name, last_name, age, gender, license_number, license_expiry_date, years_of_experience, is_available, status) VALUES ($1, 'Jane', 'Doe', 30, 'female', 'TEST-LIC-002', '2030-01-01', 3, true, 'active') ON CONFLICT DO NOTHING`, driverID)

	adminToken := testhelpers.GenerateTestToken("admin-uuid-001", "admin", app.Config)
	body := map[string]interface{}{
		"car_id": carID, "driver_id": driverID,
		"pickup_date":     "2026-08-01T09:00:00Z",
		"return_date":     "2026-08-05T09:00:00Z",
		"pickup_location": "VI", "destination": "Ajah",
	}

	// first booking should succeed
	w := app.MakeRequest("POST", "/api/bookings", body, adminToken)
	assert.Equal(t, http.StatusCreated, w.Code, "first booking should succeed")

	// second booking for same car/dates should FAIL
	w = app.MakeRequest("POST", "/api/bookings", body, adminToken)
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"double booking same dates should fail")

	t.Logf("Double booking correctly blocked")
}
