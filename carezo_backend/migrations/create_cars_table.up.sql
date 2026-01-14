CREATE TABLE IF NOT EXISTS (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),


    -- Car Details
    model VARCHAR(100) NOT NULL,
    brand VARCHAR(100) NOT NULL,
    year >= 1900 AND year <= EXTRACT(YEAR FROM CURRENT_DATE) + 1),
    color VARCHAR(50) NOT NULL,
    license_plate VARCHAR(20) UNIQUE NOT NULL,

    -- Specification
    engine_output VARCHAR(50),
    transmission VARCHAR(20) CHECK  (transmission IN ("automatic","manual")),
    fuel_type VARCHAR(20) CHECK (fuel_type IN ("petrol", "diesel", "electric", "hybrid")),
    seating_capacity INTEGER NOT NULL CHECK (seating_capacity >= 2 AND seating_capacity <= 17),
    maximum_speed INTEGER
    mileage INTEGER DEFAULT 0,


    -- pricing
    hourly_rate DECIMAL(10, 2) NOT NULL CHECK (hourly_rate > 0),
    caution_fee DECIMAL(10, 2) NOT NULL DEFAULT 200000 CHECK (caution >= 0),

    -- features(stored as JSON array)
    features JSONB DEFAULT '[]'::jsonb, -- ["GPS", "Bluetooth", "AC", "Leather Seats"]
    
    -- Images (stored as JSON array of URLs)
    images JSONB DEFAULT '[]'::jsonb, -- ["url1", "url2", "url3"]
    
    -- Availability
    is_available BOOLEAN DEFAULT TRUE,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'maintenance', 'retired')),
    
    -- Location
    current_location VARCHAR(255),
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP

);


-- Indexes
CREATE INDEX idx_cars_status ON cars(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_cars_available ON cars(is_available) WHERE deleted_at IS NULL;
CREATE INDEX idx_cars_brand_model ON cars(brand, model);
CREATE INDEX idx_cars_hourly_rate ON cars(hourly_rate);

-- Update timestamp trigger
CREATE TRIGGER update_cars_updated_at BEFORE UPDATE ON cars
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();