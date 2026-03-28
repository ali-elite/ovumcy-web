CREATE TABLE IF NOT EXISTS oidc_identities (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  issuer TEXT NOT NULL,
  subject TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_used_at TIMESTAMP,
  CONSTRAINT fk_oidc_identities_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_oidc_identities_issuer_subject ON oidc_identities(issuer, subject);
CREATE INDEX IF NOT EXISTS idx_oidc_identities_user_id ON oidc_identities(user_id);
