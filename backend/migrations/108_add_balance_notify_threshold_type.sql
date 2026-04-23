-- Add threshold type support (fixed / percentage) to balance notification.
ALTER TABLE users ADD COLUMN IF NOT EXISTS balance_notify_threshold_type VARCHAR(10) NOT NULL DEFAULT 'fixed';
ALTER TABLE users ADD COLUMN IF NOT EXISTS total_recharged DECIMAL(20,8) NOT NULL DEFAULT 0;
