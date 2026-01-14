CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),


    -- Authentication
    email VARCHAR(255) UNIQUE,
    phone_number VARCHAR(15) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,

    -- OAUTH
    google_id VARCHAR(255) UNIQUE,
    oauth_provider VARCHAR(100) NOT NULL,


    -- Profile Information
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    age INTEGER CHECK(age >= 18 AND age <=100),
    profession VARCHAR(100),
    location VARCHAR(100),
    profile_image_url TEXT,

    -- Verification
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    email_verification_token VARCHAR(255),
    phone_verification_token VARCHAR(6),
    otp_expires_at TIMESTAMP,


    -- Password Reset
    reset_token VARCHAR(255),
    reset_token_expires_at TIMESTAMP

    -- timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    deleted_at TIMESTAMP,
)
    -- indexes for faster queries
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_phone ON users(phone_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_google_id ON users(google_id) WHERE deleted_at is NULL;
CREATE INDEX idx_users_role ON users(role);

-- Update timestamp trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURN TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP
    RETURN NEW;
END;
$$ language "plpgsql";

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_updated_at_columnt();