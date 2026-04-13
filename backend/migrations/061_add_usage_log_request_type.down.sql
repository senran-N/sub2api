DROP INDEX IF EXISTS idx_usage_logs_request_type_created_at;
ALTER TABLE usage_logs DROP CONSTRAINT IF EXISTS usage_logs_request_type_check;
ALTER TABLE usage_logs DROP COLUMN IF EXISTS request_type;
