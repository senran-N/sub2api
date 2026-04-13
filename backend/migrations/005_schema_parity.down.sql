DROP INDEX IF EXISTS idx_usage_logs_billing_type;
ALTER TABLE usage_logs DROP COLUMN IF EXISTS billing_type;

DROP INDEX IF EXISTS idx_redeem_codes_group_id;
ALTER TABLE redeem_codes DROP COLUMN IF EXISTS validity_days;
ALTER TABLE redeem_codes DROP COLUMN IF EXISTS group_id;

DROP INDEX IF EXISTS idx_accounts_overload_until;
DROP INDEX IF EXISTS idx_accounts_rate_limit_reset_at;
DROP INDEX IF EXISTS idx_accounts_rate_limited_at;
DROP INDEX IF EXISTS idx_accounts_schedulable;
ALTER TABLE accounts DROP COLUMN IF EXISTS session_window_status;
ALTER TABLE accounts DROP COLUMN IF EXISTS session_window_end;
ALTER TABLE accounts DROP COLUMN IF EXISTS session_window_start;
ALTER TABLE accounts DROP COLUMN IF EXISTS overload_until;
ALTER TABLE accounts DROP COLUMN IF EXISTS rate_limit_reset_at;
ALTER TABLE accounts DROP COLUMN IF EXISTS rate_limited_at;
ALTER TABLE accounts DROP COLUMN IF EXISTS schedulable;

ALTER TABLE api_keys ALTER COLUMN key TYPE VARCHAR(64);

ALTER TABLE users DROP COLUMN IF EXISTS notes;
ALTER TABLE users DROP COLUMN IF EXISTS wechat;
ALTER TABLE users DROP COLUMN IF EXISTS username;

DROP TABLE IF EXISTS settings;
