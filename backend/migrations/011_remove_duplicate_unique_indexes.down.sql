CREATE UNIQUE INDEX IF NOT EXISTS apikey_key ON api_keys(key);
CREATE UNIQUE INDEX IF NOT EXISTS api_keys_key ON api_keys(key);
CREATE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key);

CREATE UNIQUE INDEX IF NOT EXISTS user_email ON users(email);
CREATE UNIQUE INDEX IF NOT EXISTS users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

CREATE UNIQUE INDEX IF NOT EXISTS settings_key ON settings(key);
CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(key);

CREATE UNIQUE INDEX IF NOT EXISTS redeemcode_code ON redeem_codes(code);
CREATE UNIQUE INDEX IF NOT EXISTS redeem_codes_code ON redeem_codes(code);
CREATE INDEX IF NOT EXISTS idx_redeem_codes_code ON redeem_codes(code);

CREATE UNIQUE INDEX IF NOT EXISTS group_name ON groups(name);
CREATE UNIQUE INDEX IF NOT EXISTS groups_name ON groups(name);
CREATE INDEX IF NOT EXISTS idx_groups_name ON groups(name);
