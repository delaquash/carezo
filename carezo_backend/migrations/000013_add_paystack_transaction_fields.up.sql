ALTER TABLE bookings ADD COLUMN IF NOT EXISTS paystack_transaction_id BIGINT;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS payment_channel VARCHAR(50);