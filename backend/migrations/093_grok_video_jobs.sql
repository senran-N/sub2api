CREATE TABLE IF NOT EXISTS grok_video_jobs (
    id BIGSERIAL PRIMARY KEY,
    job_id TEXT NOT NULL UNIQUE,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    group_id BIGINT REFERENCES groups(id) ON DELETE SET NULL,
    requested_model TEXT NOT NULL DEFAULT '',
    canonical_model TEXT NOT NULL DEFAULT '',
    request_payload_snapshot JSONB,
    upstream_status TEXT NOT NULL DEFAULT '',
    normalized_status TEXT NOT NULL DEFAULT '',
    poll_after TIMESTAMPTZ NULL,
    error_code TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_grok_video_jobs_account_id ON grok_video_jobs(account_id);
CREATE INDEX IF NOT EXISTS idx_grok_video_jobs_group_id ON grok_video_jobs(group_id);
CREATE INDEX IF NOT EXISTS idx_grok_video_jobs_updated_at ON grok_video_jobs(updated_at DESC);
