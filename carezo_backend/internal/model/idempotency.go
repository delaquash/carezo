type IdempotencyRecord struct {
	ID             string    `db:"id"`
	IdempotencyKey string    `db:"idempotency_key"`
	UserID         string    `db:"user_id"`
	RequestPath    string    `db:"request_path"`
	RequestHash    string    `db:"request_hash"`
	Status         string    `db:"status"`
	ResponseStatus *int      `db:"response_status"` 
	ResponseBody   *string   `db:"response_body"`   
	CreatedAt      time.Time `db:"created_at"`
	ExpiresAt      time.Time `db:"expires_at"`
}