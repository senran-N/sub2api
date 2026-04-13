DROP INDEX IF EXISTS idx_usage_logs_service_tier_created_at;
ALTER TABLE usage_logs DROP COLUMN IF EXISTS service_tier;
