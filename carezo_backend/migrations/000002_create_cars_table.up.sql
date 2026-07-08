
CREATE TABLE IF NOT EXISTS cars (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic Information
    model VARCHAR(100) NOT NULL,

    year INTEGER NOT NULL
        CHECK (
            year >= 1990
            AND year <= EXTRACT(YEAR FROM CURRENT_DATE) + 1
        ),

    color VARCHAR(50) NOT NULL,

    license_plate VARCHAR(20) NOT NULL,

    -- Vehicle Specifications
    engine_output VARCHAR(50),

    transmission VARCHAR(20) NOT NULL
        CHECK (transmission IN ('automatic', 'manual')),

    fuel_type VARCHAR(20) NOT NULL
        CHECK (fuel_type IN ('petrol', 'diesel', 'electric', 'hybrid')),

    seating_capacity INTEGER NOT NULL
        CHECK (seating_capacity BETWEEN 2 AND 15),

    maximum_speed INTEGER,

    mileage INTEGER NOT NULL DEFAULT 0
        CHECK (mileage >= 0),

    -- Optional Driver Information
    driver_name VARCHAR(100),

    driver_number VARCHAR(20),

    driver_miles INTEGER,

    -- Pricing
    hourly_rate JSONB NOT NULL,

    caution_fee NUMERIC(12,2) NOT NULL
        CHECK (caution_fee >= 0),

    -- Description
    description TEXT,

    -- Extra Features
    features JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- Images
    images JSONB NOT NULL DEFAULT '[]'::jsonb,

    image_public_ids JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- Location
    current_location TEXT,

    -- Availability
    is_available BOOLEAN NOT NULL DEFAULT TRUE,

    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (
            status IN (
                'active',
                'on_trip',
                'maintenance',
                'retired'
            )
        ),

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    deleted_at TIMESTAMP
);


CREATE UNIQUE INDEX idx_cars_license_plate
ON cars (license_plate)
WHERE deleted_at IS NULL;


CREATE INDEX idx_cars_status
ON cars(status)
WHERE deleted_at IS NULL;

CREATE INDEX idx_cars_available
ON cars(is_available)
WHERE deleted_at IS NULL;

CREATE INDEX idx_cars_transmission
ON cars(transmission);

CREATE INDEX idx_cars_fuel_type
ON cars(fuel_type);

CREATE INDEX idx_cars_seating_capacity
ON cars(seating_capacity);


CREATE TRIGGER update_cars_updated_at
BEFORE UPDATE ON cars
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();