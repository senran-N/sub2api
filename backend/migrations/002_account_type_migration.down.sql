-- sub2api:irreversible
DO $$
BEGIN
    RAISE EXCEPTION '002_account_type_migration is not safely reversible: original official/oauth/setup-token distinctions cannot be reconstructed automatically';
END
$$;
