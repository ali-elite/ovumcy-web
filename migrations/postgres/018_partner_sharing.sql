CREATE TABLE IF NOT EXISTS partner_invitations (
  id SERIAL PRIMARY KEY,
  owner_user_id INTEGER NOT NULL,
  code_hash TEXT NOT NULL UNIQUE,
  code_hint TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'redeemed', 'revoked', 'expired')),
  expires_at TIMESTAMPTZ NOT NULL,
  redeemed_at TIMESTAMPTZ,
  redeemed_by_user_id INTEGER,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_partner_invitations_owner FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_partner_invitations_redeemed_by FOREIGN KEY (redeemed_by_user_id) REFERENCES users(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_partner_invitations_owner_user_id ON partner_invitations(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_partner_invitations_status_expires_at ON partner_invitations(status, expires_at);

CREATE TABLE IF NOT EXISTS partner_links (
  id SERIAL PRIMARY KEY,
  owner_user_id INTEGER NOT NULL,
  partner_user_id INTEGER NOT NULL,
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked')),
  invited_by_user_id INTEGER NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  revoked_at TIMESTAMPTZ,
  CONSTRAINT fk_partner_links_owner FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_partner_links_partner FOREIGN KEY (partner_user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_partner_links_invited_by FOREIGN KEY (invited_by_user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT ck_partner_links_owner_not_partner CHECK (owner_user_id <> partner_user_id),
  CONSTRAINT uq_partner_links_owner_partner UNIQUE (owner_user_id, partner_user_id),
  CONSTRAINT uq_partner_links_partner_user UNIQUE (partner_user_id)
);
CREATE INDEX IF NOT EXISTS idx_partner_links_owner_user_id ON partner_links(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_partner_links_partner_user_id ON partner_links(partner_user_id);
