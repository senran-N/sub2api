DROP INDEX IF EXISTS idx_ops_retry_attempts_success_time;
ALTER TABLE ops_retry_attempts DROP COLUMN IF EXISTS response_truncated;
ALTER TABLE ops_retry_attempts DROP COLUMN IF EXISTS response_preview;
ALTER TABLE ops_retry_attempts DROP COLUMN IF EXISTS used_account_id;
ALTER TABLE ops_retry_attempts DROP COLUMN IF EXISTS upstream_request_id;
ALTER TABLE ops_retry_attempts DROP COLUMN IF EXISTS http_status_code;
ALTER TABLE ops_retry_attempts DROP COLUMN IF EXISTS success;

DROP INDEX IF EXISTS idx_ops_error_logs_unresolved_time;
DROP INDEX IF EXISTS idx_ops_error_logs_resolved_time;
ALTER TABLE ops_error_logs DROP COLUMN IF EXISTS resolved_retry_id;
ALTER TABLE ops_error_logs DROP COLUMN IF EXISTS resolved_by_user_id;
ALTER TABLE ops_error_logs DROP COLUMN IF EXISTS resolved_at;
ALTER TABLE ops_error_logs DROP COLUMN IF EXISTS resolved;

DO $$
BEGIN
    RAISE NOTICE 'rollback for 038_ops_errors... drops added schema only; standardized historical classification values are not restored';
END
$$;
