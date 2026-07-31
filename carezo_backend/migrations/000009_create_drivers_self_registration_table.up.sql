-- Allow 'driver' as a valid role, alongside the existing 'user'/'admin'.
ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('user', 'admin', 'driver'));

-- Links a drivers row to its login account.
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id);
ALTER TABLE drivers ADD CONSTRAINT drivers_user_id_unique UNIQUE (user_id);

ALTER TABLE drivers ADD COLUMN IF NOT EXISTS verification_status VARCHAR(30)
    NOT NULL DEFAULT 'pending_profile';
ALTER TABLE drivers DROP CONSTRAINT IF EXISTS drivers_verification_status_check;
ALTER TABLE drivers ADD CONSTRAINT drivers_verification_status_check
    CHECK (verification_status IN (
        'pending_profile',
        'pending_documents',
        'pending_review',
        'approved',
        'rejected'
    ));

ALTER TABLE drivers ADD COLUMN IF NOT EXISTS nin VARCHAR(20);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS nin_document_url TEXT;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS license_document_url TEXT;

ALTER TABLE drivers ADD COLUMN IF NOT EXISTS rejection_reason TEXT;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS reviewed_by UUID REFERENCES users(id);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP;

ALTER TABLE drivers ADD COLUMN IF NOT EXISTS bank_account_name VARCHAR(255);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS bank_account_number VARCHAR(20);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS bank_name VARCHAR(100);