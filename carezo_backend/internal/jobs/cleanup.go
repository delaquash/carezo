package jobs

import (
	"log"
	"time"

	"github.com/delaquash/carezo/internal/database"
	"github.com/robfig/cron/v3"
)

func StartCleanUpJobs() {
	// cron.New() creates a scheduler — think of it as a clock that watches
	// for specific times and runs a function when it hits them. Nothing
	// runs yet; this just creates the clock.
	c := cron.New()

	// "0 3 * * *" is cron syntax for "at 3:00 AM, every day." The five
	// slots are: minute, hour, day-of-month, month, day-of-week — a * means
	// "any." 3 AM is chosen because it's low-traffic, so a slow DELETE
	// query won't compete with real users for database time.
	c.AddFunc("0 3 * * *", cleanupExpiredIdempotencyKeys)

	// start() actually turns the clock on, from this point, it checks
	// the schedule continuously in the background, in its own goroutine,
	// without blocking anything else your app is doing.
	c.Start()

	log.Println("Cleanup jobs scheduled")
}

// This is the actual "throw away old tickets" function — one plain SQL
// DELETE,
func cleanupExpiredIdempotencyKeys() {
	result, err := database.DB.Exec(`
		DELETE FROM idempotency_keys WHERE expires_at < $1
	`, time.Now())

	if err != nil {
		// Logged, not fatal — a failed cleanup run shouldn't crash your
		// whole app. It just tries again tomorrow at 3 AM.
		log.Printf("failed to clean up expired idempotency keys: %v", err)
		return
	}

	rows, _ := result.RowsAffected()
	log.Printf("Cleanup: removed %d expired idempotency keys", rows)
}
