ALTER TABLE channel_model_pricing DROP COLUMN IF EXISTS per_request_price;
ALTER TABLE channels DROP COLUMN IF EXISTS restrict_models;
