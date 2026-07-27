ALTER TABLE drivers DROP COLUMN IF EXISTS bank_account_name;
ALTER TABLE drivers DROP COLUMN IF EXISTS bank_account_number;
ALTER TABLE drivers DROP COLUMN IF EXISTS bank_name;

ALTER TABLE drivers DROP COLUMN IF EXISTS rejection_reason;
ALTER TABLE drivers DROP COLUMN IF EXISTS reviewed_by;
ALTER TABLE drivers DROP COLUMN IF EXISTS reviewed_at;

ALTER TABLE drivers DROP COLUMN IF EXISTS nin;
ALTER TABLE drivers DROP COLUMN IF EXISTS nin_document_url;
ALTER TABLE drivers DROP COLUMN IF EXISTS license_document_url;

ALTER TABLE drivers DROP CONSTRAINT IF EXISTS drivers_verification_status_check;
ALTER TABLE drivers DROP COLUMN IF EXISTS verification_status;

ALTER TABLE drivers DROP CONSTRAINT IF EXISTS drivers_user_id_unique;
ALTER TABLE drivers DROP COLUMN IF EXISTS user_id;

ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('user', 'admin'));