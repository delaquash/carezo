-- Drop trigger if it exists
DROP TRIGGER IF EXISTS update_cars_updated_at ON cars;

-- Drop the table and all dependent objects
DROP TABLE IF EXISTS cars CASCADE;