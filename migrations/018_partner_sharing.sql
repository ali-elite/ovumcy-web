PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS partner_invitations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  owner_user_id INTEGER NOT NULL,
  code_hash TEXT NOT NULL UNIQUE,
  code_hint TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'redeemed', 'revoked', 'expired')),
  expires_at DATETIME NOT NULL,
  redeemed_at DATETIME,
  redeemed_by_user_id INTEGER,
  revoked_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (redeemed_by_user_id) REFERENCES users(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_partner_invitations_owner_user_id ON partner_invitations(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_partner_invitations_status_expires_at ON partner_invitations(status, expires_at);

CREATE TABLE IF NOT EXISTS partner_links (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  owner_user_id INTEGER NOT NULL,
  partner_user_id INTEGER NOT NULL,
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked')),
  invited_by_user_id INTEGER NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  revoked_at DATETIME,
  FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (partner_user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (invited_by_user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT ck_partner_links_owner_not_partner CHECK (owner_user_id <> partner_user_id),
  UNIQUE (owner_user_id, partner_user_id),
  UNIQUE (partner_user_id)
);
CREATE INDEX IF NOT EXISTS idx_partner_links_owner_user_id ON partner_links(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_partner_links_partner_user_id ON partner_links(partner_user_id);
