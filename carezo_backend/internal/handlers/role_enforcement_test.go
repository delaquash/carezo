
package handlers_test

import (
    "net/http"
    "testing"

    "github.com/delaquash/carezo/internal/testhelpers"
    "github.com/stretchr/testify/assert"
)

func TestRoleEnforcement(t *testing.T) {
    app := testhelpers.SetUpTestApp(t)
    t.Cleanup(func() { app.CleanUpDB(t) })

    // generate tokens for different roles WITHOUT logging in
    // GenerateTestToken creates a valid JWT with any role we specify
    userToken  := testhelpers.GenerateTestToken(
        "user-uuid-001", "user", app.Config.JWTSecret,
    )
    adminToken := testhelpers.GenerateTestToken(
        "admin-uuid-001", "admin", app.Config.JWTSecret,
    )

    // a valid car body for create car tests
    carBody := map[string]interface{}{
        "model":            "Camry",
        "year":             2022,
        "color":            "Black",
        "license_plate":    "TEST-ROLE-001",
        "engine_output":    "200HP",
        "transmission":     "automatic",
        "fuel_type":        "petrol",
        "seating_capacity": 5,
        "maximum_speed":    180,
        "mileage":          10000,
        "driver_name":      "Test Driver",
        "driver_number":    "+2348012345678",
        "driver_miles":     5000,
        "hourly_rate": map[string]interface{}{
            "weekday": 5000, "weekend": 7000, "holiday": 10000,
        },
        "caution_fee":      50000,
        "current_location": "Lagos",
    }

    // ── Test table — defines all role enforcement scenarios ────────
    // Table-driven tests are idiomatic Go — define inputs/expectations as data
    // then loop through them instead of copy-pasting the same test logic
    tests := []struct {
        name           string
        method         string
        url            string
        body           interface{}
        token          string
        expectedStatus int
        description    string
    }{
        // ── Admin endpoint tests ────────────────────────────────────
        {
            name:           "user_cannot_create_car",
            method:         "POST",
            url:            "/api/admin/cars",
            body:           carBody,
            token:          userToken,
            expectedStatus: http.StatusForbidden,  // must be 403
            description:    "regular user should NOT be able to create a car",
        },
        {
            name:           "admin_can_create_car",
            method:         "POST",
            url:            "/api/admin/cars",
            body:           carBody,
            token:          adminToken,
            expectedStatus: http.StatusCreated,    // must be 201
            description:    "admin should be able to create a car",
        },
        {
            name:           "user_cannot_delete_car",
            method:         "DELETE",
            url:            "/api/admin/cars/some-car-id",
            body:           nil,
            token:          userToken,
            expectedStatus: http.StatusForbidden,
            description:    "regular user should NOT be able to delete a car",
        },
        {
            name:           "user_cannot_create_driver",
            method:         "POST",
            url:            "/api/admin/drivers",
            body:           map[string]interface{}{"first_name": "Test"},
            token:          userToken,
            expectedStatus: http.StatusForbidden,
            description:    "regular user should NOT be able to create a driver",
        },
        // ── CRITICAL: confirm c.Abort() is working ───────────────────
        // Before the Abort() fix, this would return 403 AND create the car
        // After the fix, it returns 403 AND the car is NOT created
        {
            name:           "forbidden_request_does_not_execute_handler",
            method:         "POST",
            url:            "/api/admin/cars",
            body:           carBody,
            token:          userToken,
            expectedStatus: http.StatusForbidden,
            description:    "403 must mean the handler was NOT called (Abort check)",
        },
        // ── Unauthenticated access tests ─────────────────────────────
        {
            name:           "no_token_cannot_access_bookings",
            method:         "GET",
            url:            "/api/bookings",
            body:           nil,
            token:          "",   // no token at all
            expectedStatus: http.StatusUnauthorized,  // must be 401
            description:    "unauthenticated user should get 401",
        },
        {
            name:           "no_token_cannot_create_booking",
            method:         "POST",
            url:            "/api/bookings",
            body:           map[string]interface{}{"car_id": "some-id"},
            token:          "",
            expectedStatus: http.StatusUnauthorized,
            description:    "unauthenticated booking attempt should get 401",
        },
        // ── Public endpoint tests (should work without token) ─────────
        {
            name:           "public_can_list_cars",
            method:         "GET",
            url:            "/api/cars",
            body:           nil,
            token:          "",   // no token needed
            expectedStatus: http.StatusOK,
            description:    "listing cars is public — no token required",
        },
        {
            name:           "public_can_list_drivers",
            method:         "GET",
            url:            "/api/drivers",
            body:           nil,
            token:          "",
            expectedStatus: http.StatusOK,
            description:    "listing drivers is public — no token required",
        },
    }

    // loop through every test case
    for _, tc := range tests {
        tc := tc // capture range variable (important for goroutines, good habit)

        t.Run(tc.name, func(t *testing.T) {
            w := app.MakeRequest(tc.method, tc.url, tc.body, tc.token)

            assert.Equal(t, tc.expectedStatus, w.Code,
                "FAILED: %s\nExpected status %d, got %d\nResponse: %s",
                tc.description,
                tc.expectedStatus,
                w.Code,
                w.Body.String(),
            )

            t.Logf("✅ %s", tc.description)
        })
    }

    // ── Extra check: confirm Abort() actually stopped handler execution ──
    t.Run("abort_prevents_car_creation", func(t *testing.T) {
        // count cars BEFORE the forbidden request
        var countBefore int
        app.DB.Get(&countBefore, "SELECT COUNT(*) FROM cars")

        // fire the forbidden request (user token on admin endpoint)
        w := app.MakeRequest("POST", "/api/admin/cars", carBody, userToken)
        assert.Equal(t, http.StatusForbidden, w.Code)

        // count cars AFTER — must be the same
        var countAfter int
        app.DB.Get(&countAfter, "SELECT COUNT(*) FROM cars")

        assert.Equal(t, countBefore, countAfter,
            "car count should not change after forbidden request — Abort() must be working. "+
            "Before: %d, After: %d", countBefore, countAfter)

        t.Logf("c.Abort() confirmed working — car count unchanged: %d", countAfter)
    })
}