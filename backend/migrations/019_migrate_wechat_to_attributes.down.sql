ALTER TABLE users ADD COLUMN IF NOT EXISTS wechat VARCHAR(100) DEFAULT '';

UPDATE users u
SET wechat = uav.value
FROM user_attribute_values uav
JOIN user_attribute_definitions uad ON uav.attribute_id = uad.id
WHERE uav.user_id = u.id
  AND uad.key = 'wechat'
  AND uad.deleted_at IS NULL;

DELETE FROM user_attribute_values
WHERE attribute_id IN (
    SELECT id FROM user_attribute_definitions WHERE key = 'wechat' AND deleted_at IS NULL
);

UPDATE user_attribute_definitions
SET deleted_at = NOW()
WHERE key = 'wechat' AND deleted_at IS NULL;
