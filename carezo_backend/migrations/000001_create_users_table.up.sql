
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Authentication (Email-only now)
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    
    -- OAuth
    google_id VARCHAR(255) UNIQUE,
    oauth_provider VARCHAR(50) DEFAULT 'local', -- 'google', 'local'
    
    -- Profile Information
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone_number VARCHAR(20), -- Now optional, just for contact
    age INTEGER CHECK (age >= 18 AND age <= 120),
    age INTEGER CHECK (age >= 18 AND age <= 120),
    profession VARCHAR(100),
    location VARCHAR(255),
    profile_image_url TEXT,
    
    -- Verification (Email-only)
    email_verified BOOLEAN DEFAULT FALSE,
    email_verification_token VARCHAR(255),
    otp_expires_at TIMESTAMP,
    
    -- Account Status
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    
    -- Password Reset
    reset_token VARCHAR(255),
    reset_token_expires_at TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Indexes
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_google_id ON users(google_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON users(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users(role);

-- Update timestamp trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
