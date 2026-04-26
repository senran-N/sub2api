DROP INDEX IF EXISTS idx_user_affiliates_admin_settings;

ALTER TABLE user_affiliates
    DROP COLUMN IF EXISTS aff_code_custom;

ALTER TABLE user_affiliates
    DROP COLUMN IF EXISTS aff_rebate_rate_percent;
