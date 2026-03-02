DROP INDEX IF EXISTS users_email_unique_not_null;

UPDATE users
SET email = 'user_' || id || '@invalid.local'
WHERE email IS NULL;

ALTER TABLE users ALTER COLUMN email SET NOT NULL;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
