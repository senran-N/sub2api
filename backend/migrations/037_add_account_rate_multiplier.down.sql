ALTER TABLE IF EXISTS usage_logs
  DROP COLUMN IF EXISTS account_rate_multiplier;

ALTER TABLE IF EXISTS accounts
  DROP COLUMN IF EXISTS rate_multiplier;
