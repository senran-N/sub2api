DROP INDEX IF EXISTS idx_api_keys_last_used_at;
ALTER TABLE api_keys DROP COLUMN IF EXISTS last_used_at;
