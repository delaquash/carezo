REATE TABLE IF NOT EXISTS cars (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Basic Information
    model VARCHAR(100) NOT NULL,
    brand VARCHAR(100) NOT NULL,
    year INTEGER NOT NULL CHECK (year >= 1900 AND year <= EXTRACT(YEAR FROM CURRENT_DATE) + 1),
    color VARCHAR(50) NOT NULL,
    license_plate VARCHAR(20) UNIQUE NOT NULL,
    
    -- Specifications
    engine_output VARCHAR(50),
    transmission VARCHAR(20) CHECK (transmission IN ('automatic', 'manual')),
    fuel_type VARCHAR(20) CHECK (fuel_type IN ('petrol', 'diesel', 'electric', 'hybrid')),
    seating_capacity INTEGER NOT NULL CHECK (seating_capacity >= 2 AND seating_capacity <= 15),
    maximum_speed INTEGER,
    mileage INTEGER DEFAULT 0,
    
    -- Driver Information (assigned driver for this car)
    driver_name VARCHAR(255),
    driver_number VARCHAR(100),
    driver_miles INTEGER,
    
    -- Pricing (hourly rate stored as JSON for flexibility)
    hourly_rate JSONB NOT NULL DEFAULT '{"standard": 50, "weekend": 60, "holiday": 70}'::jsonb,
    caution_fee DECIMAL(10, 2) NOT NULL DEFAULT 50.00 CHECK (caution_fee >= 0),
    
    -- Features (stored as JSON array)
    features JSONB DEFAULT '[]'::jsonb,
    
    -- Images (stored as JSON array of URLs)
    images JSONB DEFAULT '[]'::jsonb,
    
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

-- Indexes for better query performance
CREATE INDEX idx_cars_status ON cars(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_cars_available ON cars(is_available) WHERE deleted_at IS NULL;
CREATE INDEX idx_cars_brand_model ON cars(brand, model);
CREATE INDEX idx_cars_transmission ON cars(transmission);
CREATE INDEX idx_cars_fuel_type ON cars(fuel_type);
CREATE INDEX idx_cars_seating_capacity ON cars(seating_capacity);
CREATE INDEX idx_cars_location ON cars(current_location);

-- Update timestamp trigger
CREATE TRIGGER update_cars_updated_at BEFORE UPDATE ON cars
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Sample data (optional - for testing)
INSERT INTO cars (
    model, brand, year, color, license_plate, engine_output, transmission,
    fuel_type, seating_capacity, maximum_speed, mileage, driver_name, driver_number,
    driver_miles, hourly_rate, caution_fee, features, current_location
) VALUES 
(
    'Corolla', 'Toyota', 2023, 'Black', 'ABC-123-XY', '1.8L 4-cylinder', 'automatic',
    'petrol', 5, 180, 15000, 'Ibrahim Musa', 'DL-12345', 5000,
    '{"standard": 50, "weekend": 60, "holiday": 70}'::jsonb, 100.00,
    '["GPS", "Bluetooth", "AC", "Backup Camera"]'::jsonb, 'Lagos, Victoria Island'
),
(
    'Camry', 'Toyota', 2024, 'Silver', 'DEF-456-ZY', '2.5L 4-cylinder', 'automatic',
    'hybrid', 5, 200, 5000, 'Fatima Yusuf', 'DL-67890', 2000,
    '{"standard": 80, "weekend": 100, "holiday": 120}'::jsonb, 150.00,
    '["GPS", "Bluetooth", "AC", "Leather Seats", "Sunroof"]'::jsonb, 'Lagos, Lekki'
),
(
    'Hilux', 'Toyota', 2023, 'White', 'GHI-789-WY', '2.8L Diesel', 'manual',
    'diesel', 5, 160, 25000, 'Ahmed Bello', 'DL-11223', 10000,
    '{"standard": 70, "weekend": 85, "holiday": 100}'::jsonb, 120.00,
    '["GPS", "AC", "4WD"]'::jsonb, 'Abuja, Wuse'
);
