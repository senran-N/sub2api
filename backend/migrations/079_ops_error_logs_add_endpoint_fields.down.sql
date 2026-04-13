ALTER TABLE ops_error_logs DROP COLUMN IF EXISTS request_type;
ALTER TABLE ops_error_logs DROP COLUMN IF EXISTS upstream_model;
ALTER TABLE ops_error_logs DROP COLUMN IF EXISTS requested_model;
ALTER TABLE ops_error_logs DROP COLUMN IF EXISTS upstream_endpoint;
ALTER TABLE ops_error_logs DROP COLUMN IF EXISTS inbound_endpoint;
