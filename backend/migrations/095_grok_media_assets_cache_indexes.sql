CREATE INDEX IF NOT EXISTS idx_grok_media_assets_content_hash_ready
    ON grok_media_assets(content_hash)
    WHERE content_hash IS NOT NULL AND local_path IS NOT NULL AND status = 'ready';

CREATE INDEX IF NOT EXISTS idx_grok_media_assets_expires_at
    ON grok_media_assets(expires_at)
    WHERE expires_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_grok_media_assets_local_path
    ON grok_media_assets(local_path)
    WHERE local_path IS NOT NULL;
