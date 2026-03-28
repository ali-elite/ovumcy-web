ALTER TABLE users ADD COLUMN local_auth_enabled BOOLEAN NOT NULL DEFAULT 1;

UPDATE users
SET local_auth_enabled = 1
WHERE local_auth_enabled IS NULL;
