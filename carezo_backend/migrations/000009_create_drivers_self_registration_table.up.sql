ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('users', 'admin', 'driver'));


ALTER TABLE drivers ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id);
ALTER TABLE drivers ADD CONSTRAINT drivers_user_id_unique UNIQUE (user_id)

ALTER TABLE drivers DROP CONSTRAINT IF EXISTS drivers_verification_status_check;
ALTER TABLE drivers ADD CONSTRAINT drivers_verification_status_check
    CHECK (verification_status IN (
        'pending_profile',
        'pending_documents',
        'pending_review',
        'approved',
        'rejected'
    ));

ALTER TABLE drivers ALTER COLUMN verification_status SET DEFAULT 'pending_profile';

-- The NIN itself (a number the driver types in) is distinct from the
-- uploaded document image proving it 

ALTER TABLE drivers ADD COLUMN IF NOT EXISTS nin VARCHAR(20);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS nin_document_url TEXT;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS license_document_url TEXT;


-- Rejection needs a reason (required by the PRD's P0 requirements), and an
-- audit trail of who reviewed it and when — useful the moment a driver
-- disputes a decision and you need to know which admin acted, and when.
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS rejection_reason TEXT;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS reviewed_by UUID REFERENCES users(id);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP;


-- Bank details — submission only, per your clarification (no payout logic
-- yet). Plain nullable text columns; populated only once
-- verification_status = 'approved'.
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS bank_account_name VARCHAR(255);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS bank_account_number VARCHAR(20);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS bank_name VARCHAR(100);