ALTER TABLE usage_logs DROP COLUMN IF EXISTS billing_tier;
ALTER TABLE usage_logs DROP COLUMN IF EXISTS model_mapping_chain;
ALTER TABLE usage_logs DROP COLUMN IF EXISTS channel_id;
ALTER TABLE channels DROP COLUMN IF EXISTS billing_model_source;
