DROP INDEX IF EXISTS idx_api_keys_expires_at;
DROP INDEX IF EXISTS idx_api_keys_quota_quota_used;
ALTER TABLE api_keys DROP COLUMN IF EXISTS expires_at;
ALTER TABLE api_keys DROP COLUMN IF EXISTS quota_used;
ALTER TABLE api_keys DROP COLUMN IF EXISTS quota;
