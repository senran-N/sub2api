DROP INDEX IF EXISTS idx_ops_error_logs_is_count_tokens;
ALTER TABLE ops_error_logs DROP COLUMN IF EXISTS is_count_tokens;
