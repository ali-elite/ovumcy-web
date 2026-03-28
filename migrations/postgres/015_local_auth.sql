ALTER TABLE users ADD COLUMN local_auth_enabled BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE users
SET local_auth_enabled = TRUE
WHERE local_auth_enabled IS NULL;
