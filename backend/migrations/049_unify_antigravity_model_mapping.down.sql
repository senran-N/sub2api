-- sub2api:irreversible
DO $$
BEGIN
    RAISE EXCEPTION '049_unify_antigravity_model_mapping is not safely reversible: prior model_mapping/model_whitelist state was overwritten and removed';
END
$$;
