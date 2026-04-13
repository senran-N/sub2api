ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS cache_creation5m_tokens INT NOT NULL DEFAULT 0;

ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS cache_creation1h_tokens INT NOT NULL DEFAULT 0;

UPDATE usage_logs
SET cache_creation5m_tokens = cache_creation_5m_tokens
WHERE cache_creation5m_tokens = 0
  AND cache_creation_5m_tokens <> 0;

UPDATE usage_logs
SET cache_creation1h_tokens = cache_creation_1h_tokens
WHERE cache_creation1h_tokens = 0
  AND cache_creation_1h_tokens <> 0;
