CREATE TABLE IF NOT EXISTS drivers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()

    -- drivers persaonl info
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    age INTEGER NOT NULL CHECK (age >= 21 AND <=50 ),
    gender VARCHAR(10) CHECK (gender IN ("male", "female")),
    nationality VARCHAR(100).
    religion VARCHAR(100),
    complexion VARCHAR(10) CHECK (complexion IN("fair", "medium", "dark")),
    height INTEGER,

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
    
    -- Ratings
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
)

- Indexes
CREATE INDEX idx_drivers_status ON drivers(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_drivers_available ON drivers(is_available) WHERE deleted_at IS NULL;
CREATE INDEX idx_drivers_gender ON drivers(gender);
CREATE INDEX idx_drivers_rating ON drivers(average_rating DESC);

-- Update timestamp trigger
CREATE TRIGGER update_drivers_updated_at BEFORE UPDATE ON drivers
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();