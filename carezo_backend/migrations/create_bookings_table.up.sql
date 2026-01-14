CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_reference VARCHAR(20) UNIQUE NOT NULL,


    -- foreign keys
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    car_id UUID NOT NULL REFERENCES  cars(id) ON DELETE RESTRICT,
    driver_id UUID REFERENCES drivers(id) ON DELETE RESTRICT,


    -- Booking details
    pickup_date TIMESTAMP NOT NULL,
    return_date TIMESTAMP NOT NULL,
    actual_return_date TIMESTAMP,

    total_hours INTEGER GENERATED ALWAYS AS (
        EXTRACT(EPOCH FROM (return_date - pickup_date)) / 3600
    ) STORED


    -- location
    destination VARCHAR(255) NOT NULL,
    pickup_location VARCHAR(255),

    -- Pricing
    hourly_rate DECIMAL(10, 2) NOT NULL,
    caution_fee DECIMAL(10, 2) NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    refundable_amount DECIMAL(10, 2) DEFAULT 0.00,

    -- Payment
    payment_status VARCHAR(20) DEFAULT 'pending' CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded', 'partially_refunded')),
    payment_reference VARCHAR(255),
    paid_at TIMESTAMP,
    refunded_at TIMESTAMP,

     -- Status
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'in_progress', 'completed', 'cancelled')),
    
    -- Cancellation
    cancelled_at TIMESTAMP,
    cancellation_reason TEXT,
    
    -- Special Requests
    special_requests TEXT,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP


    -- Indexes
    CREATE INDEX idx_bookings_user ON bookings(user_id);
    CREATE INDEX idx_bookings_car ON bookings(car_id);
    CREATE INDEX idx_bookings_driver ON bookings(driver_id);
    CREATE INDEX idx_bookings_status ON bookings(status);
    CREATE INDEX idx_bookings_payment_status ON bookings(payment_status);
    CREATE INDEX idx_bookings_dates ON bookings(pickup_date, return_date);
    CREATE INDEX idx_bookings_reference ON bookings(booking_reference);

    -- Update timestamp trigger
    CREATE TRIGGER update_bookings_updated_at BEFORE UPDATE ON bookings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

    -- Check constraint: return date must be after pickup date
    ALTER TABLE bookings ADD CONSTRAINT check_return_after_pickup 
    CHECK (return_date > pickup_date);
    
)