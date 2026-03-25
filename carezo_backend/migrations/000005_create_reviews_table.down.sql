-- Drop the trigger that updates driver rating first
DROP TRIGGER IF EXISTS trigger_update_driver_rating ON reviews;

-- Drop the function for updating driver rating
DROP FUNCTION IF EXISTS update_driver_rating();

-- Drop the trigger that updates review timestamps
DROP TRIGGER IF EXISTS update_reviews_updated_at ON reviews;

-- Finally, drop the reviews table along with dependencies
DROP TABLE IF EXISTS reviews CASCADE;