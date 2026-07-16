package handlers_test

import (
	"net/http"
	"testing"

	"github.com/delaquash/carezo/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)


func TestAuthFlow(t *testing.T) {
	// Spin up a full router + test DB + test Redis, same pattern as every
	// other integration test in this package. t.Cleanup runs even if the
	// test fails partway, so leftover rows/keys don't bleed into the next run.
	app := testhelpers.SetUpTestApp(t)
	t.Cleanup(func() { app.CleanUpDB(t) })


	email:= "authflow_test@carezo.com"
	password := "Test123@"

	registerBody := map[string]interface{} {
		"fullName": "Auth Flow Test",
		"email":  email,
		"password": password,
		// RegisterRequest requires "country" (binding:"required") — omitting
		// this would fail validation before we even reach the handler logic
		// we're trying to test, so it has to be present even though this
		// test isn't about country validation.
		"country": "Nigeria",
	
	}

	// ── Step 1: Register — happy path ───────────────────────────────
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
}