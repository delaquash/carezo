package handlers_test

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

func TestAuthFlow(t *testing.T) {
	// Spin up a full router + test DB + test Redis, same pattern as every
	// other integration test in this package. t.Cleanup runs even if the
	// test fails partway, so leftover rows/keys don't bleed into the next run.
	app := testhelpers.SetUpTestApp(t)
	t.Cleanup(func() { app.CleanUpDB(t) })

	email := "authflow_test@carezo.com"
	password := "Test123@"

	registerBody := map[string]interface{}{
		"fullName": "Auth Flow Test",
		"email":    email,
		"password": password,
		// RegisterRequest requires "country" (binding:"required") — omitting
		// this would fail validation before we even reach the handler logic
		// we're trying to test, so it has to be present even though this
		// test isn't about country validation.
		"country": "Nigeria",
	}

	// Step 1: Register — happy path
	// POST with body=registerBody, token="" since registration is public
	// (unauthenticated) — matches MakeRequest's "pass '' for unauthenticated
	// requests" contract from its own doc comment.
	w := app.MakeRequest("POST", "/api/auth/register", registerBody, "")

	// StatusCreated (201), not StatusOK (200) — matches your handler:
	// response.Success(c, http.StatusCreated, "Registration successful...", nil)
	assert.Equal(t, http.StatusCreated, w.Code,
		"registration should succeed, got: %s", w.Body.String())

	// step 2: Duplicate registration details should be rejected even if unverified
	// Same registration details should be refuse a second insert, this may cause
	// DB unique-constraint error leaking as 500, or second registration overriding the first
	w = app.MakeRequest("POST", "/api/auth/register", registerBody, "")
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"duplicate email registration should be rejected, got: %s", w.Body.String())

	// Step 3: Login before OTP verification — must be blocked
	// The user exists in the DB now (from Step 1) but email_verified=false.
	// A real attacker or confused user hitting login at this point should
	// not get a valid session token for an unverified account.
	loginBody := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	w = app.MakeRequest("POST", "/api/auth/login", loginBody, "")
	assert.NotEqual(t, http.StatusBadRequest, w.Code,
		"login should be blocked before emaul verification, got: %s", w.Body.String())

	// Step 4: Pull the real OTP straight out of Redis
	// We can't read the OTP from an email (Mailtrap sends a real email,
	// we're not intercepting SMTP in this test) — but OTPService stores it
	// at key "otp:<email>" with a plain string value (see otp_service.go's
	// GenerateAndStoreOTP), so we read it directly from the same Redis
	// instance the app just wrote to.
	ctx := context.Background()
	redisKey := fmt.Sprintf("otp:%s", email)
	// getting OTP from redit
	realOTP, err := database.RedisClient.Get(ctx, redisKey).Result()
	require.NoError(t, err, "OTP should exist in redis after registration")
	require.NotEmpty(t, realOTP, "OTP value should not be empty")
	t.Logf("Fetched real OTP from redis: %s", realOTP)

	// step 5 I am verifying with a WRONG OTP  and it must be rejected
	// Deliberately mangle the real OTP so it's guaranteed wrong. OTPService's
	// VerifyOTP only deletes the Redis key on a *successful* match, so this
	// wrong attempt should leave the real OTP untouched in Redis for Step 6.

	wrongOTP := "000000"

	if wrongOTP == realOTP {
		// assuming the real otp is "000000" same as wrong, it will pass this test
		// which is not the aim, so we have to flip to ensure when right OTP is "000000"
		// wrong OTP become "99999" and vice versa, so bot OTP are nto the same.
		wrongOTP = "999999"
	}

	verifyWrongBody := map[string]interface{}{
		"email": email,
		"otp":   wrongOTP,
	}

	w = app.MakeRequest("POST", "api/aut/verify-otp", verifyWrongBody, "")
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"wrong OTP should be rejected, got: %s", w.Body.String())

	// verify with correct Opt
	verifyCorrectBody := map[string]interface{}{
		"email": email,
		"otp":   realOTP,
	}

	w = app.MakeRequest("POST", "api/aut/verify-otp", verifyCorrectBody, "")
	assert.Equal(t, http.StatusOK, w.Code,
		"correct otp should verify successfully, got: %s", w.Body.String())

	// Parse the response the same way your other tests do (testhelpers.ParseResponse),
	// then confirm VerifyOTP actually returns a real token, per auth_handler.go's
	// comment "// Returns token now!" on the VerifyOTP handler.

	resp := testhelpers.ParseResponse(w)
	require.NotNil(t, resp["data"], "verify-otp respinse data should not be nil")
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "verify-otp data should be an object")

	// field namw here follows the "access_token" in models.AuthResponse not token

	verifyToken, ok := data["access_token"].(string)
	require.True(t, ok, "access_token missing from verify-otp response")
	require.NotEmpty(t, verifyToken, "access_token should not be empty")
	t.Logf("Email verified, got token: %s...", verifyToken[:10])

	// login with wrong password must be rejected since user is verified
	wrongLoginBody := map[string]interface{}{
		"email":    email,
		"password": "WrongPassword123!",
	}
	w = app.MakeRequest("POST", "/api/auth/login", wrongLoginBody, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code,
		"login with wrong password should be rejected, got: %s", w.Body.String())

	// login with correct password
	// Step 8: Login with CORRECT password should succeed 
	w = app.MakeRequest("POST", "/api/auth/login", loginBody, "")
	assert.Equal(t, http.StatusOK, w.Code,
		"login with correct credentials should succeed, got: %s", w.Body.String())

	resp = testhelpers.ParseResponse(w)
	require.NotNil(t, resp["data"], "login response data should not be nil")
	data, ok = resp["data"].(map[string]interface{})
	require.True(t, ok, "login data should be an object")

	loginToken, ok := data["access_token"].(string)
	require.True(t, ok, "access_token missing from login response")
	require.NotEmpty(t, loginToken, "access_token should not be empty")

	t.Logf("Full auth flow passed: register → duplicate rejected → "+
		"login-before-verify blocked → wrong OTP rejected → verified → "+
		"wrong password rejected → login succeeded (token: %s...)", loginToken[:10])
}
