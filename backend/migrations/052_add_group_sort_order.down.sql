DROP INDEX IF EXISTS idx_groups_sort_order;
ALTER TABLE groups DROP COLUMN IF EXISTS sort_order;
