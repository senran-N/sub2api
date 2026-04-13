ALTER TABLE groups DROP COLUMN IF EXISTS sora_storage_quota_bytes;

ALTER TABLE users
    DROP COLUMN IF EXISTS sora_storage_used_bytes,
    DROP COLUMN IF EXISTS sora_storage_quota_bytes;

DROP TABLE IF EXISTS sora_generations;
