DO $$
BEGIN
    RAISE NOTICE 'usage_logs partition bootstrap rollback is a no-op; manual cleanup required if partitions were created';
END
$$;
