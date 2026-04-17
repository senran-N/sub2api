ALTER TABLE ops_error_logs
    ADD COLUMN IF NOT EXISTS wait_user_ms BIGINT,
    ADD COLUMN IF NOT EXISTS wait_account_ms BIGINT,
    ADD COLUMN IF NOT EXISTS ws_acquire_ms BIGINT,
    ADD COLUMN IF NOT EXISTS ws_healthcheck_ms BIGINT;
