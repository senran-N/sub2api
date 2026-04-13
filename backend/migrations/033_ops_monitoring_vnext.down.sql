DROP TABLE IF EXISTS ops_alert_events;
DROP TABLE IF EXISTS ops_alert_rules;
DROP TABLE IF EXISTS ops_job_heartbeats;
DROP TABLE IF EXISTS ops_system_metrics;
DROP TABLE IF EXISTS ops_retry_attempts;
DROP TABLE IF EXISTS ops_error_logs;
DROP TABLE IF EXISTS ops_metrics_daily;
DROP TABLE IF EXISTS ops_metrics_hourly;

DO $$
BEGIN
    RAISE NOTICE 'rollback for 033_ops_monitoring_vnext drops vNext ops tables but does not restore pre-vNext ops data/schema';
END
$$;
