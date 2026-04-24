-- Ensure previously soft-deleted API keys/accounts are not still marked active.
UPDATE api_keys
SET status = 'disabled',
    updated_at = NOW()
WHERE deleted_at IS NOT NULL
  AND status = 'active';

UPDATE accounts
SET status = 'disabled',
    schedulable = FALSE,
    updated_at = NOW()
WHERE deleted_at IS NOT NULL
  AND (status = 'active' OR schedulable = TRUE);
