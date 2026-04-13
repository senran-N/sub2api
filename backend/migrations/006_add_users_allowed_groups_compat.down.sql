DO $$
BEGIN
    RAISE NOTICE 'rollback for 006_add_users_allowed_groups_compat is a no-op compatibility migration';
END
$$;
