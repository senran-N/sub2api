-- Recreate legacy columns for schema compatibility only.
-- Historical values dropped by the forward migration are not recoverable here.
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS cache_creation5m_tokens INT NOT NULL DEFAULT 0;
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS cache_creation1h_tokens INT NOT NULL DEFAULT 0;
