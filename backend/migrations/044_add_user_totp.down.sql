DROP INDEX IF EXISTS idx_users_totp_enabled;
ALTER TABLE users
  DROP COLUMN IF EXISTS totp_enabled_at,
  DROP COLUMN IF EXISTS totp_enabled,
  DROP COLUMN IF EXISTS totp_secret_encrypted;
