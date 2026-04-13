DROP INDEX IF EXISTS idx_usage_cleanup_tasks_canceled_at;
ALTER TABLE usage_cleanup_tasks
    DROP COLUMN IF EXISTS canceled_at,
    DROP COLUMN IF EXISTS canceled_by;
