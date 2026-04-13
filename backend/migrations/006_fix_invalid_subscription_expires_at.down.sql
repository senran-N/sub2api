DO $$
BEGIN
    RAISE NOTICE 'rollback for 006_fix_invalid_subscription_expires_at is a no-op; original invalid expires_at values are not recoverable';
END
$$;
