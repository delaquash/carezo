-- Drop the update trigger first
DROP TRIGGER IF EXISTS update_bookings_updated_at ON bookings;

-- Drop the table along with all dependent objects
DROP TABLE IF EXISTS bookings CASCADE;