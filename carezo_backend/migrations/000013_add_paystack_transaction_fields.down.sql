ALTER TABLE bookings DROP COLUMN IF EXISTS paystack_transaction_id;
ALTER TABLE bookings DROP COLUMN IF EXISTS payment_channel;
