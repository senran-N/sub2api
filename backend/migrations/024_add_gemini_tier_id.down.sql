UPDATE accounts
SET credentials = credentials - 'tier_id'
WHERE platform = 'gemini'
  AND type = 'oauth'
  AND credentials ? 'tier_id';
