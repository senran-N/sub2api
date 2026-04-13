DROP INDEX IF EXISTS usersubscription_deleted_at;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS deleted_at;
