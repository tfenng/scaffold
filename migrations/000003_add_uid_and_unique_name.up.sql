-- Add uid column as nullable first
ALTER TABLE users ADD COLUMN IF NOT EXISTS uid TEXT;
-- Update existing rows with generated uid
UPDATE users SET uid = 'user_' || id WHERE uid IS NULL OR uid = '';
-- Set NOT NULL constraint
ALTER TABLE users ALTER COLUMN uid SET NOT NULL;
-- Add unique constraints
ALTER TABLE users ADD CONSTRAINT users_uid_unique UNIQUE (uid);
ALTER TABLE users ADD CONSTRAINT users_name_unique UNIQUE (name);
