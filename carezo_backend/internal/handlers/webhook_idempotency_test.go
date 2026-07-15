package handlers_test

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/delaquash/carezo/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateWebhookSignature creates the HMAC-SHA512 signature
// that Paystack sends in the x-paystack-signature header
func generateWebhookSignature(payload []byte, secret string) string {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestWebhookIdempotency(t *testing.T) {
	app := testhelpers.SetUpTestApp(t)
	t.Cleanup(func() { app.CleanUpDB(t) })

	// seed a booking in pending state
	// this simulates a booking that was created but not yet paid
	userID := "aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	carID := "bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	bookingID := "cccc3333-cccc-cccc-cccc-cccccccccccc"
	paymentRef := "test_ref_idempotency_001"

	app.DB.Exec(`
		INSERT INTO users(id, email, password_hash, first_name, last_name,
		role, status, email_verified
		) VALUES ( $1, 'webhook@test.com', 'hash', 'W', 'user', 'active', true)
		 ON CONFLICT (id) DO NOTHING
	`, userID)

	app.DB.Exec(`
		INSERT INTO cars(id, model, year, color, license_plate, transmission,
		fuel_type, seating_capacity, caution_fee, is_available, status)
		VALUES ($1, 'Civic', 2023, 'Red', 'TEST-003', 'automatic', 'petrol, 50000, true, 'active')
        ON CONFLICT (id) DO NOTHING
    `, carID)

	// insert booking directly — simulates a booking that called /initialize
	// and got a payment_reference back from Paystack
	_, err := app.DB.Exec(`
        INSERT INTO bookings (
            id, booking_reference, user_id, car_id,
            pickup_date, return_date, pickup_location, destination,
            caution_fee, total_amount, refundable_amount,
            payment_status, payment_reference, status
        ) VALUES (
            $1, 'BK-WEBHOOK-TEST', $2, $3,
            '2026-08-10T09:00:00Z', '2026-08-15T09:00:00Z',
            'Ikeja', 'Lekki',
            50000, 300000, 50000,
            'pending', $4, 'pending'
        )
    `, bookingID, userID, carID, paymentRef)
	require.NoError(t, err, "failed to seed test booking")

	// 2) Paystack webhook payload
	// response from paystack when payment succeed
	webhookPayload := []byte(`{
		"event": "charge.success",
		"data": {
			"reference": "` + paymentRef + `",
			"status": "success",
			"amount": 30000000,
			"currency": "NGN",
			"paid_at": "2026-07-20T09:00:00.000Z",
			"customer": {
                "email": "webhook@test.com"
            }
		}
	}`)

	   // sign the payload with your webhook secret
    // (set PAYSTACK_WEBHOOK_SECRET in your test config)
    signature := generateWebhookSignature(
        webhookPayload,
        app.Config.PaystackWebhookSecret,
    )

	// send the webhook the first time
	req1 := testhelpers.NewRequest("POST", "/api/payment/webhook", webhookPayload)
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("x-paystack-signature", signature)

	 w1 := app.MakeRequest(req1)
    assert.Equal(t, http.StatusOK, w1.Code,
        "first webhook call should return 200, got: %s", w1.Body.String())

    // verify the booking is now confirmed after first webhook
    var bookingStatus, paymentStatus string
    err = app.DB.QueryRow(
        "SELECT status, payment_status FROM bookings WHERE id = $1",
        bookingID,
    ).Scan(&bookingStatus, &paymentStatus)
    require.NoError(t, err)

	 assert.Equal(t, "confirmed", bookingStatus,   "booking should be confirmed after webhook")
    assert.Equal(t, "paid",      paymentStatus,   "payment should be paid after webhook")

    t.Logf("First webhook processed: status=%s payment=%s", bookingStatus, paymentStatus)

	// Send the EXACT SAME webhook a second time ──────────
    // This simulates Paystack retrying delivery (they retry on non-200 responses)
    w2 := app.MakeRequest(
        testhelpers.NewRequest("POST", "/api/payments/webhook", webhookPayload),
    )

    // second call should also return 200 (acknowledge it) but do nothing
    assert.Equal(t, http.StatusOK, w2.Code,
        "second webhook call should also return 200")

		// CRITICAL CHECK: booking state must be unchanged after second webhook
    // If this check fails it means your webhook is NOT idempotent
    var statusAfterSecond, paymentAfterSecond string
    app.DB.QueryRow(
        "SELECT status, payment_status FROM bookings WHERE id = $1",
        bookingID,
    ).Scan(&statusAfterSecond, &paymentAfterSecond)

    assert.Equal(t, "confirmed", statusAfterSecond,
        "booking should still be confirmed after duplicate webhook")
    assert.Equal(t, "paid", paymentAfterSecond,
        "payment should still be 'paid' after duplicate webhook — not changed to 'pending'")

    t.Logf("✅ Duplicate webhook correctly ignored")

    // ── Step 5: Check total_amount was not doubled ─────────────────
    // A non-idempotent system might credit the amount twice
    var totalAmount int
    app.DB.Get(&totalAmount, "SELECT total_amount FROM bookings WHERE id = $1", bookingID)
    assert.Equal(t, 300000, totalAmount,
        "total_amount should still be 300000, not doubled to 600000")

    t.Logf("✅ Amount not duplicated: %d", totalAmount)
}

func TestWebhookRejectsInvalidSignature(t *testing.T) {
    // Security test — webhook with wrong signature must be rejected
    app := testhelpers.SetUpTestApp(t)

    payload := []byte(`{"event":"charge.success","data":{"reference":"fake_ref"}}`)

    req := testhelpers.NewRawRequest("POST", "/api/payments/webhook", payload)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-paystack-signature", "invalid_signature_here")

    w := app.MakeRawRequest(req)

    // must be rejected — 400 or 401
    assert.NotEqual(t, http.StatusOK, w.Code,
        "webhook with invalid signature should be rejected")

    t.Logf("✅ Invalid signature correctly rejected with status %d", w.Code)
}
