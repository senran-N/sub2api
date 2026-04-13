DO $$
BEGIN
    RAISE NOTICE 'rollback for 006b_guard_users_allowed_groups is a no-op compatibility migration';
END
$$;
