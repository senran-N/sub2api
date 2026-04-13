-- sub2api:irreversible
DO $$
BEGIN
    RAISE EXCEPTION '052_migrate_upstream_to_apikey is not safely reversible: original upstream/apikey account intent cannot be inferred automatically';
END
$$;
