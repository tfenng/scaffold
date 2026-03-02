ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

UPDATE users
SET email = NULL
WHERE email IS NOT NULL AND btrim(email) = '';

UPDATE users
SET email = NULL
WHERE email IS NOT NULL
  AND email !~* '^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$';

ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique_not_null
ON users (email)
WHERE email IS NOT NULL;
