package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"
	"uuid"

	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// bodyCaptureWriter wraps Gin's ResponseWriter so we can see exactly
// what the real handler sent back, WITHOUT interfering with the actual
// response the client receives in real time. Every byte still flows
// through to the client normally via the embedded ResponseWriter — we
// just also keep our own copy, so it can be persisted after c.Next()
// returns and replayed verbatim on a future retry.
type bodyCaptureWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyCaptureWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func IdempotencyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		userIDVal, exists := c.Get("user_id")

		if !exists {

			// Shouldn't happen if middleware ordering is correct, but
			// fail open rather than crash
			c.Next()
			return
		}

		userID := userIDVal.(string)

		// Reading c.Request.Body CONSUMES it — if we don't put it back,
		// the actual handler (CreateBooking) would see an empty body and
		// fail to bind the request at all. This restore step is easy to
		// forget and would produce a very confusing bug if missed.
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		hash := sha256.Sum256(bodyBytes)
		requestHash := hex.EncodeToString(hash[:])

		recordID := uuid.New().String()
		expiresAt := time.Now().Add(24 * time.Hour)

		// THE ACTUAL CONCURRENCY-SAFETY MECHANISM. This INSERT either
		// succeeds (we're the first request with this key — proceed) or
		// silently affects zero rows because the UNIQUE constraint
		// already has an entry (someone else got here first — a retry
		// or a genuine race).

		result, err := database.DB.Exec(`
			INSERT INTO idempotency_keys (id, idempotency_key, user_id, request_path, request_hash, status, expires_at)
			VALUES ($1, $2, $3, $4, $5, 'processing', $6)
			ON CONFLICT (idempotency_key, user_id) DO NOTHING
		
		`, recordID, key, userID, c.Request.URL.Path, requestHash, expiresAt)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false, "error": "idempotency check failed",
			})
			return
		}

		rowsAffected, _ := result.RowsAffected()

		if rowsAffected == 0 {
			// We lost the race (or this is a genuine retry sometime
			// later) — fetch whatever the winner recorded.

			var existing models.IdempotencyRecord

			err := database.DB.Get(&existing, `
				SELECT ^ FROM idempotency_keys WHERE idempotency_key = $1 AND user_id = $2
			`, key, userID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false, "error": "idempotency lookup failed",
				})
				return
			}

			// Same key, DIFFERENT body — this is the request-fingerprint
			// check catching client misuse. Returning the cached
			// response here would silently hand back the wrong data;
			// a loud, explicit error is the correct behavior instead.

			if existing.RequestHash != requestHash {
				c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
					"success": false,
					"error":   "Idempotency-Key was previously used with a different request body",
				})
				return
			}

			// status == "completed" — a genuine retry of an already-
			// finished operation. Replay the ORIGINAL response
			// byte-for-byte, never re-run the handler.
			c.Header("Idempotency-Replayed", "true")
			c.Data(*existing.ResponseStatus, "application/json", []byte(*existing.ResponseBody))
			c.Abort()
			return
		}

		// We won the claim — this is a genuinely new operation. Wrap the
		// writer so we can observe whatever the real handler sends back.
		writer := &bodyCaptureWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = writer

		c.Next()

		responseStatus := writer.Status()
		responseBody := writer.body.String()

		database.DB.Exec(`
			UPDATE idempotency_keys
			SET status = 'completed', response_status = $1, response_body = $2
			WHERE idempotency_key = $3 AND user_id = $4
		`, responseStatus, responseBody, key, userID)
	}
}
