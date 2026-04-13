DROP INDEX IF EXISTS idx_usage_logs_ip_address;
ALTER TABLE usage_logs DROP COLUMN IF EXISTS ip_address;
