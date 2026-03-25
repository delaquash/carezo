-- Drop indexes first
DROP INDEX IF EXISTS idx_sessions_expires;
DROP INDEX IF EXISTS idx_sessions_token;
DROP INDEX IF EXISTS idx_sessions_user;

-- Drop the helper function
DROP FUNCTION IF EXISTS delete_expired_sessions();

-- Drop the table
DROP TABLE IF EXISTS sessions CASCADE;