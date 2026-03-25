-- Drop the update trigger first
DROP TRIGGER IF EXISTS update_drivers_updated_at ON drivers;

-- Drop the table along with all dependent objects
DROP TABLE IF EXISTS drivers CASCADE;