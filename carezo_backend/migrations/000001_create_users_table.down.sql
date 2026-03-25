-- Drop trigger if exists
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function if exists
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop table and cascade dependent objects
DROP TABLE IF EXISTS users CASCADE;