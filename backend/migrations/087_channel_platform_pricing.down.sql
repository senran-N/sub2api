-- Revert nested anthropic-only model_mapping back to flat format when the row
-- still matches the pure migrated shape {"anthropic": {...}}.
UPDATE channels
SET model_mapping = COALESCE(model_mapping->'anthropic', '{}'::jsonb)
WHERE jsonb_typeof(model_mapping) = 'object'
  AND model_mapping ? 'anthropic'
  AND jsonb_object_length(model_mapping) = 1
  AND jsonb_typeof(model_mapping->'anthropic') = 'object';

DROP INDEX IF EXISTS idx_channel_model_pricing_platform;
ALTER TABLE channel_model_pricing DROP COLUMN IF EXISTS platform;
