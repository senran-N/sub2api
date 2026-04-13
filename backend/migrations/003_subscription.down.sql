DROP INDEX IF EXISTS idx_usage_logs_sub_created;
DROP INDEX IF EXISTS idx_usage_logs_subscription_id;
DROP INDEX IF EXISTS idx_usage_logs_group_id;

ALTER TABLE usage_logs DROP COLUMN IF EXISTS first_token_ms;
ALTER TABLE usage_logs DROP COLUMN IF EXISTS rate_multiplier;
ALTER TABLE usage_logs DROP COLUMN IF EXISTS subscription_id;
ALTER TABLE usage_logs DROP COLUMN IF EXISTS group_id;

DROP TABLE IF EXISTS user_subscriptions;

DROP INDEX IF EXISTS idx_groups_subscription_type;
DROP INDEX IF EXISTS idx_groups_platform;

ALTER TABLE groups DROP COLUMN IF EXISTS default_validity_days;
ALTER TABLE groups DROP COLUMN IF EXISTS monthly_limit_usd;
ALTER TABLE groups DROP COLUMN IF EXISTS weekly_limit_usd;
ALTER TABLE groups DROP COLUMN IF EXISTS daily_limit_usd;
ALTER TABLE groups DROP COLUMN IF EXISTS subscription_type;
ALTER TABLE groups DROP COLUMN IF EXISTS platform;
