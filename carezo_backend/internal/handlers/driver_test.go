package handlers

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/delaquash/carezo/internal/database"
	"github.com/delaquash/carezo/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	registerDriverURL       = "/api/drivers/register"
	completeProfileURL      = "/api/drivers/complete-profile"
	uploadDocumentURL       = "/api/drivers/documents"
	submitBankDetailsURL    = "/api/drivers/bank-details"
	reviewApplicationURLFmt = "/api/admin/drivers/%s/review"
)

func TestDriverOnboardingFlow_ApprovalPath(t *testing.T) {
	app := testhelpers.SetUpTestApp(t)
	t.Cleanup(func() { app.CleanUpDB(t) })

	driverEmail := "driver_approval_flow@carezo.com"
	driverPassword := "Test123!@#"

	// register driver
	registerBody := map[string]interface{}{
		"first_name":   "Emmanuel",
		"last_name":    "Olaide",
		"gender":       "male",
		"phone_number": "+2348064965574",
		"email":        driverEmail,
		"password":     driverPassword,
	}

	w := app.MakeRequest("POST", registerDriverURL, registerBody, "")
	require.Equal(t, http.StatusCreated, w.Code, "driver registration failed: %s", w.Body.String())

	resp := testhelpers.ParseResponse(w)
	require.NotNil(t, resp["data"], "register response should not be nil")
	data := resp["data"].(map[string]interface{})
	driverID := data["id"].(string)
	require.NotEmpty(t, driverID)

	assert.Equal(t, "pending_profile", data["verification_status"])

	// login before verifying must be blocked
	loginBody := map[string]interface{}{"email": driverEmail, "password": driverPassword}
	w = app.MakeRequest("POST", "/api/auth/login", loginBody, "")
	assert.NotEqual(t, http.StatusOK, w.Code,
		"login should be blocked before OTP verifiction, got: %s", w.Body.String())

	// Pulling the real OTP from REdis
	ctx := context.Background()
	redisKey := fmt.Sprintf("otp:%s", driverEmail)
	realOTP, err := database.RedisClient.Get(ctx, redisKey).Result()
	require.NoError(t, err, "OTP should exist in redis after driver registration")
	require.NotEmpty(t, realOTP)

	// Verify via shared endpoint
	verifyBody := map[string]interface{}{"email": driverEmail, "otp": realOTP}
	w = app.MakeRequest("POST", "/api/auth/verify-otp", verifyBody, "")
	require.Equal(t, http.StatusOK, w.Code, "OTP verification failed: %s", w.Body.String())

	verifyResp := testhelpers.ParseResponse(w)
	verifyData := verifyResp["data"].(map[string]interface{})
	driverToken := verifyData["access_token"].(string)
	require.NotEmpty(t, driverToken)

	// Login after verifying
	completeProfileBody := map[string]interface{}{
		"age":                 28,
		"gender":              "male",
		"state":               "Lagos",
		"nationality":         "Nigerian",
		"complexion":          "dark",
		"height":              175,
		"license_number":      "LASDRV-00123",
		"license_expiry_date": "2030-01-01",
		"years_of_experience": 5,
	}

	w = app.MakeRequest("PUT", completeProfileURL, completeProfileBody, driverToken)
	require.Equal(t, http.StatusOK, w.Code, "complete profile failed: %s", w.Body.String())

	completeResp := testhelpers.ParseResponse(w)
	completeData := completeResp["data"].(map[string]interface{})
	assert.Equal(t, "pending_documents", completeData["verification_status"])


	// Upload Document:- not real document
	w = app.MakeMultipartRequest(
		"POST", uploadDocumentURL,
		map[string]string{"nin": "1234567890"},
		map[string][]byte{
			"nin_document":  []byte("fake-nin-image-bytes"),
			"license_document": []byte("fake-license-image-bytes"),
		},
		driverToken,
	)

	require.Equal(t, http.StatusOK, w.Code, "document upload failed: %s", w.Body.String())
	uploadResp := testhelpers.ParseResponse(w)
	uploadData := uploadResp["data"].(map[string]interface{})
	assert.Equal(t, "pending_review", uploadData["verification_status"])


	// Admin approves
	adminToken := loginAsTestAdmin(t, app, "driver_review_admin@carezo.com", "Test123!@#")
	reviewBody := map[string]interface{}{"approved": true}
	reviewURL := fmt.Sprintf(reviewApplicationURLFmt, driverID)
	w = app.MakeRequest("POST", reviewURL, reviewBody, adminToken)
	require.Equal(t, http.StatusOK, w.Code, "admin approval failed: %s", w.Body.String())

	reviewResp := testhelpers.ParseResponse(w)
	reviewData := reviewResp["data"].(map[string]interface{})
	assert.Equal(t, "approved", reviewData["verification_status"])

	// Submit bank details
	bankBody := map[string]interface{}{
		"bank_account_name":   "Tunde Bakare",
		"bank_account_number": "0123456789",
		"bank_name":           "GTBank",
	}

	w = app.MakeRequest("POST", submitBankDetailsURL, bankBody, driverToken)
	assert.Equal(t, http.StatusOK, w.Code, "bank details submission failed: %s", w.Body.String())

	t.Logf("Full approved driver onboarding flow passed: register → login-blocked-pre-verify → "+
		"verify (shared endpoint) → login → complete profile → upload documents → "+
		"admin approved → bank details submitted")

}


func TestDriverOnboardingFlow_RejectedPath(t *testing.T) {
	app := testhelpers.SetUpTestApp(t)
	t.Cleanup(func() { app.CleanUpDB(t) })

	driverEmail := "driver_rejected_flow@carezo.com"
	driverPassword := "Test123!@#"

	// register, verify, complete profile and upload documents
	registerBody := map[string]interface{}{
		"first_name": "Ada", "last_name": "Obi", "gender": "female",
		"phone_number": "+2348099998888", "email": driverEmail, "password": driverPassword,
	}

	w := app.MakeRequest("POST", registerDriverURL, registerBody, "")
	require.Equal(t, http.StatusCreated, w.Code, "registration failed: %s", w.Body.String())
	driverID := testhelpers.ParseResponse(w)["data"].(map[string]interface{})["id"].(string)

	ctx := context.Background()
	realOTP, err := database.RedisClient.Get(ctx, fmt.Sprintf("otp:%s", driverEmail)).Result()
	require.NoError(t, err)

	w = app.MakeRequest("POST", "/api/auth/verify-otp",
		map[string]interface{}{"email": driverEmail, "otp": realOTP}, "")
	require.Equal(t, http.StatusOK, w.Code)
	driverToken := testhelpers.ParseResponse(w)["data"].(map[string]interface{})["access_token"].(string)

		w = app.MakeRequest("PUT", completeProfileURL, map[string]interface{}{
		"age": 30, "gender": "female", "state": "Abuja", "nationality": "Nigerian",
		"complexion": "fair", "height": 165, "license_number": "ABJDRV-00456",
		"license_expiry_date": "2029-06-01", "years_of_experience": 3,
	}, driverToken)
	require.Equal(t, http.StatusOK, w.Code, "complete profile failed: %s", w.Body.String())

	w = app.MakeMultipartRequest("POST", uploadDocumentURL,
		map[string]string{"nin": "98765432109"},
		map[string][]byte{"nin_document": []byte("fake"), "license_document": []byte("fake")},
		driverToken)
	require.Equal(t, http.StatusOK, w.Code, "upload failed: %s", w.Body.String())


	// reject without a reason must be blocked
	adminToken := loginAsTestAdmin(t, app, "driver_reject_admin@carezo.com", "Test123!@#")
	reviewURL := fmt.Sprintf(reviewApplicationURLFmt, driverID)
}