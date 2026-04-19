CREATE TABLE IF NOT EXISTS grok_media_assets (
    asset_id TEXT PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    job_id TEXT,
    requested_model TEXT NOT NULL DEFAULT '',
    canonical_model TEXT NOT NULL DEFAULT '',
    asset_type TEXT NOT NULL,
    upstream_url TEXT NOT NULL,
    local_path TEXT,
    content_hash TEXT,
    mime_type TEXT,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'remote',
    expires_at TIMESTAMPTZ,
    last_access_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_grok_media_assets_account_id ON grok_media_assets(account_id);
CREATE INDEX IF NOT EXISTS idx_grok_media_assets_job_id ON grok_media_assets(job_id);
CREATE INDEX IF NOT EXISTS idx_grok_media_assets_updated_at ON grok_media_assets(updated_at DESC);

ALTER TABLE grok_video_jobs ADD COLUMN IF NOT EXISTS output_asset_id TEXT REFERENCES grok_media_assets(asset_id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_grok_video_jobs_output_asset_id ON grok_video_jobs(output_asset_id);
