DROP INDEX IF EXISTS idx_ual_frozen_thaw;

ALTER TABLE user_affiliate_ledger
    DROP COLUMN IF EXISTS frozen_until;

ALTER TABLE user_affiliates
    DROP COLUMN IF EXISTS aff_frozen_quota;
