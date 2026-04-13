ALTER TABLE usage_logs DROP COLUMN IF EXISTS media_type;

ALTER TABLE groups
    DROP COLUMN IF EXISTS sora_video_price_per_request_hd,
    DROP COLUMN IF EXISTS sora_video_price_per_request,
    DROP COLUMN IF EXISTS sora_image_price_540,
    DROP COLUMN IF EXISTS sora_image_price_360;
