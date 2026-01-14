DROP TRIGGER IF EXISTS trigger_update_driver_rating ON reviews;
DROP FUNCTION IF EXISTS update_driver_rating();
DROP TRIGGER IF EXISTS update_reviews_updated_at ON reviews;
DROP TABLE IF EXISTS reviews CASCADE;