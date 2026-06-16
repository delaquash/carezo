ALTER TABLE users
    ADD COLUMN IF NOT EXISTS profile_image_url       TEXT,
    ADD COLUMN IF NOT EXISTS profile_image_public_id TEXT;
 

 ALTER TABLE cars
    ADD COLUMN IF NOT EXISTS images          JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS image_public_ids JSONB DEFAULT '[]'::jsonb;
 
 ALTER TABLE drivers
    ADD COLUMN IF NOT EXISTS profile_image_url       TEXT,
    ADD COLUMN IF NOT EXISTS profile_image_public_id TEXT;
 

 ALTER TABLE reviews
    ADD COLUMN IF NOT EXISTS images           JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS image_public_ids JSONB DEFAULT '[]'::jsonb;
 

--  Will run this when i install docker 
-- docker exec -it carezo-postgres psql -U carezo_user -d carezo_db -f migration_add_image_columns.sql
