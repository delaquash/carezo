-- Cars Table
CREATE TABLE IF NOT EXISTS cars (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Car Specifications
    make VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INTEGER NOT NULL CHECK (year >= 1990 AND year <= EXTRACT(YEAR FROM CURRENT_DATE) + 1),
    color VARCHAR(50) NOT NULL,
    plate_number VARCHAR(20) NOT NULL,

    -- Car Category
    category VARCHAR(20) NOT NULL DEFAULT 'regular'
        CHECK (category IN ('regular', 'luxury')),
    car_type VARCHAR(50) NOT NULL
        CHECK (car_type IN ('sedan', 'suv', 'truck', 'van', 'coupe', 'convertible', 'hatchback')),

    -- Capacity
    passenger_capacity INTEGER NOT NULL DEFAULT 5
        CHECK (passenger_capacity > 0),

    -- Pricing
    hourly_rate NUMERIC(12,2) NOT NULL CHECK (hourly_rate > 0),
    caution_fee NUMERIC(12,2) NOT NULL DEFAULT 0.00 CHECK (caution_fee >= 0),

    -- Availability & Status
    is_available BOOLEAN DEFAULT TRUE,
    status VARCHAR(20) DEFAULT 'active'
        CHECK (status IN ('active', 'on_trip', 'maintenance', 'retired')),

    -- Additional Info
    description TEXT,
    features JSONB DEFAULT '[]'::jsonb,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- ✅ Unique plate number (active cars only)
CREATE UNIQUE INDEX idx_cars_plate_number_unique
ON cars(plate_number)
WHERE deleted_at IS NULL;

-- ✅ Core indexes
CREATE INDEX idx_cars_status
ON cars(status) WHERE deleted_at IS NULL;

CREATE INDEX idx_cars_available
ON cars(is_available) WHERE deleted_at IS NULL;

CREATE INDEX idx_cars_category
ON cars(category) WHERE deleted_at IS NULL;

CREATE INDEX idx_cars_hourly_rate
ON cars(hourly_rate);

CREATE INDEX idx_cars_make_model
ON cars(make, model);

-- ✅ Trigger
CREATE TRIGGER update_cars_updated_at
BEFORE UPDATE ON cars
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();