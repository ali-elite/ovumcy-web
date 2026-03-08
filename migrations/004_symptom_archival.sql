ALTER TABLE symptom_types ADD COLUMN archived_at DATETIME;
CREATE INDEX IF NOT EXISTS idx_symptom_types_user_archived_at ON symptom_types(user_id, archived_at);
