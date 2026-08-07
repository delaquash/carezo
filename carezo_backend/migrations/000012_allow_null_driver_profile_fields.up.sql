-- These columns are only known once CompleteDriverProfile runs (step 2
-- of self-registration). RegisterDriver (step 1) intentionally leaves
-- them empty — enforcement that they're eventually filled in belongs to
-- CompleteDriverProfileRequest's `binding:"required"` tags at the
-- APPLICATION layer now, not the database's NOT NULL constraint, since
-- the database has to tolerate this field being empty for the period
-- between step 1 and step 2 (verification_status = 'pending_profile').
ALTER TABLE drivers ALTER COLUMN state DROP NOT NULL;
ALTER TABLE drivers ALTER COLUMN age DROP NOT NULL;
ALTER TABLE drivers ALTER COLUMN complexion DROP NOT NULL;
ALTER TABLE drivers ALTER COLUMN height DROP NOT NULL;
ALTER TABLE drivers ALTER COLUMN license_number DROP NOT NULL;
ALTER TABLE drivers ALTER COLUMN license_expiry_date DROP NOT NULL;
ALTER TABLE drivers ALTER COLUMN years_of_experience DROP NOT NULL;
ALTER TABLE drivers ALTER COLUMN nationality DROP NOT NULL;