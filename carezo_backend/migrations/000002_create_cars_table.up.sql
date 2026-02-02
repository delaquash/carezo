CREATE TABLE IF NOT EXISTS drivers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Personal Information
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    age INTEGER NOT NULL CHECK (age >= 21 AND age <= 70),
    gender VARCHAR(10) CHECK (gender IN ('male', 'female')),
    
    -- Additional Details (as requested)
    nationality VARCHAR(100) NOT NULL,
    religion VARCHAR(100),
    complexion VARCHAR(50) NOT NULL, -- e.g., "fair", "medium", "dark"
    height INTEGER NOT NULL CHECK (height >= 140 AND height <= 220), -- in centimeters
    
    -- Contact
    phone_number VARCHAR(20) NOT NULL,
    email VARCHAR(255),
    
    -- License Information
    license_number VARCHAR(50) UNIQUE NOT NULL,
    license_expiry_date DATE NOT NULL,
    years_of_experience INTEGER DEFAULT 0 CHECK (years_of_experience >= 0),
    
    -- Profile
    profile_image_url TEXT,
    bio TEXT,
    languages JSONB DEFAULT '[]'::jsonb, -- ["English", "Yoruba", "Hausa"]
    
    -- Ratings (calculated from reviews)
    average_rating DECIMAL(3, 2) DEFAULT 0.00 CHECK (average_rating >= 0 AND average_rating <= 5),
    total_reviews INTEGER DEFAULT 0,
    total_trips INTEGER DEFAULT 0,
    
    -- Availability
    is_available BOOLEAN DEFAULT TRUE,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'on_trip', 'unavailable', 'suspended')),
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Indexes
CREATE INDEX idx_drivers_status ON drivers(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_drivers_available ON drivers(is_available) WHERE deleted_at IS NULL;
CREATE INDEX idx_drivers_gender ON drivers(gender);
CREATE INDEX idx_drivers_nationality ON drivers(nationality);
CREATE INDEX idx_drivers_religion ON drivers(religion);
CREATE INDEX idx_drivers_complexion ON drivers(complexion);
CREATE INDEX idx_drivers_rating ON drivers(average_rating DESC);
CREATE INDEX idx_drivers_experience ON drivers(years_of_experience DESC);

-- Update timestamp trigger
CREATE TRIGGER update_drivers_updated_at BEFORE UPDATE ON drivers
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Sample drivers (for testing)
INSERT INTO drivers (
    first_name, last_name, age, gender, nationality, religion, complexion, height,
    phone_number, email, license_number, license_expiry_date, years_of_experience,
    bio, languages, is_available
) VALUES 
(
    'Ibrahim', 'Musa', 35, 'male', 'Nigerian', 'Muslim', 'dark', 175,
    '+2348012345001', 'ibrahim.musa@example.com', 'DL-NG-12345', '2026-12-31', 10,
    'Experienced driver with 10 years of safe driving. Familiar with Lagos routes.',
    '["English", "Hausa", "Yoruba"]'::jsonb, true
),
(
    'Fatima', 'Yusuf', 28, 'female', 'Nigerian', 'Muslim', 'medium', 165,
    '+2348012345002', 'fatima.yusuf@example.com', 'DL-NG-67890', '2027-06-30', 5,
    'Professional female driver. Specializes in airport transfers and long distance trips.',
    '["English", "Hausa"]'::jsonb, true
),
(
    'Chinedu', 'Okafor', 40, 'male', 'Nigerian', 'Christian', 'dark', 180,
    '+2348012345003', 'chinedu.okafor@example.com', 'DL-NG-11223', '2025-12-31', 15,
    'Very experienced driver. Over 15 years of professional driving. Excellent customer service.',
    '["English", "Igbo"]'::jsonb, true
),
(
    'Blessing', 'Adeyemi', 30, 'female', 'Nigerian', 'Christian', 'fair', 170,
    '+2348012345004', 'blessing.adeyemi@example.com', 'DL-NG-44556', '2026-03-15', 7,
    'Friendly and reliable. Specializes in family trips and airport pickups.',
    '["English", "Yoruba"]'::jsonb, true
);
