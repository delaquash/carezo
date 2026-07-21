package handlers_test

import (
	"net/http"
	"testing"

	// "uuid"

	"github.com/delaquash/carezo/internal/testhelpers"
	"github.com/delaquash/carezo/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loginAsTestAdmin(t *testing.T, app *testhelpers.TestApp, email, password string) string {
	t.Helper()

	userID := uuid.New().String()
	hashedPassword, err := utils.HashPassword(password)
	require.NoError(t, err)

	_, err = app.DB.Exec(`
		INSERT INTO users (
			id, email, password_hash, first_name, last_name,
			role, status, email_verified
		) VALUES ($1, $2, $3, 'Test', 'Admin', 'admin', 'active', true)
		ON CONFLICT (email) DO UPDATE SET password_hash = $3
	`, userID, email, hashedPassword)
	require.NoError(t, err, "failed to insert test admin")

	w := app.MakeRequest("POST", "/api/auth/login", map[string]interface{}{
		"email":    email,
		"password": password,
	}, "")
	require.Equal(t, http.StatusOK, w.Code, "admin login failed: %s", w.Body.String())

	resp := testhelpers.ParseResponse(w)
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "admin login response data is missing")

	token, ok := data["access_token"].(string)
	require.True(t, ok && token != "", "admin access_token missing from login response")

	return token
}

func validCarPayload(licensePlate string) map[string]interface{} {
	return map[string]interface{}{
		"model":            "Camry",
		"year":             2022,
		"color":            "Black",
		"license_plate":    licensePlate,
		"transmission":     "automatic", // binding: oneof=automatic manual
		"fuel_type":        "petrol",    // binding: oneof=petrol diesel electric hybrid
		"seating_capacity": 5,           // binding: min=2,max=15
		"mileage":          0,
		"caution_fee":      50000,
		// "hourly_rate": map[string]interface{}{
		// 	"weekday": 5000,
		// 	"weekend": 7000,
		// 	"holiday": 10000,
		// },
	}
}


// Create car

func TestCreateCar(t *testing.T) {
	app := testhelpers.SetUpTestApp(t)

	t.Cleanup(func() { app.CleanUpDB(t) })

	adminToken := loginAsTestAdmin(t, app, "car_admin_create@carezo.com", "Test123!@#")

	w := app.MakeRequest("POST", "/api/admin/cars", validCarPayload("CREATE-001"), adminToken)
	assert.Equal(t, http.StatusCreated, w.Code, "car creation failed: %s", w.Body.String())

	resp := testhelpers.ParseResponse(w)
	require.NotNil(t, resp["data"], "response data should not be nil")
	data := resp["data"].(map[string]interface{})

	// Confirm what actually got PERSISTED matches what we sent — not just
	// that the endpoint returned 201. A handler can return 201 while
	// silently dropping/mangling a field (exactly what the licence_plate/
	// brand mismatches earlier would have caused).
	assert.Equal(t, "CREATE-001", data["license_plate"])
	assert.Equal(t, "Camry", data["model"])
	assert.NotEmpty(t, data["id"])

	t.Logf("Car created: %s", data["id"])
}


// createcar should not allow duplicate license plate, it should be rejected
func TestCreateCar_DuplicatedLicensePlate(t *testing.T) {
	app := testhelpers.SetUpTestApp(t)
	t.Cleanup(func() { app.CleanUpDB(t) })

	adminToken := loginAsTestAdmin(t, app, "car_admin_dup@carezo.com", "Test123!@#")

	payload := validCarPayload("DUP-001")

	// first creation should suceed
	w := app.MakeRequest("POST", "/api/admin/cars", payload, adminToken)
	require.Equal(t, http.StatusCreated, w.Code, "first car creation should succeed: %s", w.Body.String())


	// Second, identical creation should be rejected by the existence
	// check at the top of CreateCar — this is the check that was broken
	// by the licence_plate/license_plate spelling mismatch, so this test
	// specifically proves that fix is correct.

	w = app.MakeRequest("POST", "/api/admin/cars", payload, adminToken)
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"duplicate license plate should be rejected, got: %s", w.Body.String())

	// ── Full CRUD happy path: Get → Update → Delete ─────────────────────────
// Chained in one test rather than three separate ones because each step
// genuinely depends on the previous step's result (need a real car ID to
// update, need that same ID to delete) — same reasoning as
// TestRebookAfterCancellation being one flow instead of split tests.

func TestCarCRUDLifecycle(t *testing.T) {
	app := testhelpers.SetUpTestApp(t)
	t.Cleanup(func() { app.CleanUpDB(t) })

	adminToken := loginAsTestAdmin(t, app, "car_admin_crud@carezo.com", "Test123!@#")

	// create
	w := app.MakeRequest("POST", "/api/admin/cars", validCarPayload("CRUD-001"), adminToken)
	require.Equal(t, http.StatusCreated, w.Code, "create failed: %s", w.Body.String())
	created := testhelpers.ParseResponse(w)["data"].(map[string]interface{})
	carID := created["id"].(string)
	require.NotEmpty(t, carID)


	// get, this is public route and does not need token
	w = app.MakeRequest("GET", "/api/cars/"+carID, nil, "" )
	assert.Equal(t, http.StatusOK, w.Code, "get failed: %s", w.Body.String())
	fetched := testhelpers.ParseResponse(w)["data"].(map[string]interface{})
	assert.Equal(t, "CRUD-001", fetched["license_plate"])

	// Update
	// UpdateCarRequest fields are all pointers/optional — sending only
	// the fields we're actually changing exercises the dynamic
	// SET-clause-building logic (only changed fields get a $N placeholder).
	updateBody := map[string]interface{} {
		"color": "Blue",
		"caution_fee": 75000,

	}

	w = app.MakeRequest("PUT", "/api/admin/cars/"+carID, updateBody, adminToken)
	assert.Equal(t, http.StatusOK, w.Code, "update failed: %s", w.Body.String())
	updated := testhelpers.ParseResponse(w)["data"].map([string]interface{})
	assert.Equal(t, float6475000), updated["caution_fee"]

	// Field we did NOT touch should be unaffected — proves the dynamic
	// update only applies fields explicitly present in the request body,
	// not silently resetting everything else.
	assert.Equal(t, "CRUD-001", updated["license_plate"])

	// ── Delete (soft delete) ──
	w = app.MakeRequest("DELETE", "/api/admin/cars/"+carID, nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code, "delete failed: %s", w.Body.String())

	// Confirm the soft delete actually took — GetCarByID filters on
	// "deleted_at IS NULL", so a deleted car should now 404, not just
	// silently still show up.
	w = app.MakeRequest("GET", "/api/cars/"+carID, nil, "")
	assert.Equal(t, http.StatusNotFound, w.Code,
		"deleted car should return 404, got: %s", w.Body.String())
}
