ALTER TABLE ops_error_logs
    DROP COLUMN IF EXISTS ws_healthcheck_ms,
    DROP COLUMN IF EXISTS ws_acquire_ms,
    DROP COLUMN IF EXISTS wait_account_ms,
    DROP COLUMN IF EXISTS wait_user_ms;
