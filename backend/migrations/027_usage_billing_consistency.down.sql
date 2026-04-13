DROP INDEX IF EXISTS idx_usage_logs_request_id_api_key_unique;

DROP TABLE IF EXISTS billing_usage_entries;

DO $$
BEGIN
    RAISE NOTICE 'rollback for 027_usage_billing_consistency does not restore normalized request_id duplicates';
END
$$;
