DROP INDEX IF EXISTS idx_accounts_temp_unschedulable_until;
ALTER TABLE accounts DROP COLUMN IF EXISTS temp_unschedulable_reason;
ALTER TABLE accounts DROP COLUMN IF EXISTS temp_unschedulable_until;
