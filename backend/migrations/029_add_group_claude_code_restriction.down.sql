DROP INDEX IF EXISTS idx_groups_fallback_group_id;
DROP INDEX IF EXISTS idx_groups_claude_code_only;
ALTER TABLE groups DROP COLUMN IF EXISTS fallback_group_id;
ALTER TABLE groups DROP COLUMN IF EXISTS claude_code_only;
