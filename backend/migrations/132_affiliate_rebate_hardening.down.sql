-- sub2api:irreversible
-- This rollback cannot restore de-duplicated payment_audit_logs rows or the
-- previous fractional affiliate_rebate_rate setting values. Restore those from
-- backup if this migration must be fully reversed.
DELETE FROM payment_audit_logs
WHERE action = 'AFFILIATE_REBATE_SKIPPED'
  AND detail = '{"reason":"baseline before affiliate rebate idempotency rollout"}';

DROP INDEX IF EXISTS idx_payment_audit_logs_order_action_uniq;
DROP INDEX IF EXISTS idx_user_affiliate_ledger_action;
DROP INDEX IF EXISTS idx_user_affiliate_ledger_user_id;
DROP TABLE IF EXISTS user_affiliate_ledger;
