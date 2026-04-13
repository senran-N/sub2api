DROP INDEX IF EXISTS user_subscriptions_user_group_unique_active;
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'user_subscriptions'::regclass
          AND contype = 'u'
          AND conname = 'user_subscriptions_user_id_group_id_key'
    ) THEN
        ALTER TABLE user_subscriptions
            ADD CONSTRAINT user_subscriptions_user_id_group_id_key UNIQUE (user_id, group_id);
    END IF;
END
$$;

DROP INDEX IF EXISTS groups_name_unique_active;
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'groups'::regclass
          AND contype = 'u'
          AND conname = 'groups_name_key'
    ) THEN
        ALTER TABLE groups
            ADD CONSTRAINT groups_name_key UNIQUE (name);
    END IF;
END
$$;

DROP INDEX IF EXISTS users_email_unique_active;
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'users'::regclass
          AND contype = 'u'
          AND conname = 'users_email_key'
    ) THEN
        ALTER TABLE users
            ADD CONSTRAINT users_email_key UNIQUE (email);
    END IF;
END
$$;
