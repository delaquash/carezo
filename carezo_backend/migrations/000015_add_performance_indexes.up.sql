CREATE INDEX idx_cars_available_status ON cars(is_available, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_cars_location ON cars(current_location);
CREATE INDEX idx_bookings_car_id_status ON bookings(car_id, status);