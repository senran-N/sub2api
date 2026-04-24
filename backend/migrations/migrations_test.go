package migrations

import (
	"io/fs"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var recentMigrationDownCompanionAllowlist = map[string]struct{}{
	"081_create_channels.sql":                                 {},
	"082_refactor_channel_pricing.sql":                        {},
	"083_channel_model_mapping.sql":                           {},
	"084_channel_billing_model_source.sql":                    {},
	"085_channel_restrict_and_per_request_price.sql":          {},
	"086_channel_platform_pricing.sql":                        {},
	"087_usage_log_billing_mode.sql":                          {},
	"088_channel_billing_model_source_channel_mapped.sql":     {},
	"089_usage_log_image_output_tokens.sql":                   {},
	"090_drop_sora.sql":                                       {},
	"092_payment_orders.sql":                                  {},
	"093_payment_audit_logs.sql":                              {},
	"094_removed_payment_channels.sql":                        {},
	"095_channel_features.sql":                                {},
	"095_subscription_plans.sql":                              {},
	"096_payment_provider_instances.sql":                      {},
	"097_fix_settings_updated_at_default.sql":                 {},
	"098_migrate_purchase_subscription_to_custom_menu.sql":    {},
	"099_fix_migrated_purchase_menu_label_icon.sql":           {},
	"100_remove_easypay_from_enabled_payment_types.sql":       {},
	"101_add_account_stats_pricing.sql":                       {},
	"101_add_balance_notify_fields.sql":                       {},
	"101_add_channel_features_config.sql":                     {},
	"101_add_payment_mode.sql":                                {},
	"102_add_balance_notify_threshold_type.sql":               {},
	"102_add_out_trade_no_to_payment_orders.sql":              {},
	"103_add_allow_user_refund.sql":                           {},
	"104_migrate_notify_emails_to_struct.sql":                 {},
	"105_migrate_websearch_emulation_to_tristate.sql":         {},
	"106_add_account_stats_pricing_intervals.sql":             {},
	"107_add_account_cost_to_dashboard_tables.sql":            {},
	"108_auth_identity_foundation_core.sql":                   {},
	"109_auth_identity_compat_backfill.sql":                   {},
	"110_pending_auth_and_provider_default_grants.sql":        {},
	"111_payment_routing_and_scheduler_flags.sql":             {},
	"112_add_payment_order_provider_key_snapshot.sql":         {},
	"113_normalize_legacy_wechat_provider_key.sql":            {},
	"114_auth_identity_migration_report_resolution.sql":       {},
	"115_auth_identity_legacy_external_backfill.sql":          {},
	"116_auth_identity_legacy_external_safety_reports.sql":    {},
	"117_add_payment_order_provider_snapshot.sql":             {},
	"118_wechat_dual_mode_and_auth_source_defaults.sql":       {},
	"119_enforce_payment_orders_out_trade_no_unique.sql":      {},
	"120_enforce_payment_orders_out_trade_no_unique_notx.sql": {},
	"121_auth_identity_migration_report_type_widen.sql":       {},
	"122_pending_auth_completion_token_cleanup.sql":           {},
	"123_fix_legacy_auth_source_grant_on_signup_defaults.sql": {},
	"124_backfill_legacy_oidc_security_flags.sql":             {},
	"125_add_channel_monitors.sql":                            {},
	"125_add_group_rpm_limit.sql":                             {},
	"126_add_channel_monitor_aggregation.sql":                 {},
	"126_add_user_rpm_limit.sql":                              {},
	"127_add_user_group_rpm_override.sql":                     {},
	"127_drop_channel_monitor_deleted_at.sql":                 {},
	"128_add_channel_monitor_request_templates.sql":           {},
	"129_seed_claude_code_template.sql":                       {},
}

func TestRecentMigrationsHaveDownCompanions(t *testing.T) {
	files, err := fs.Glob(FS, "*.sql")
	require.NoError(t, err)

	fileSet := make(map[string]struct{}, len(files))
	for _, name := range files {
		fileSet[name] = struct{}{}
	}

	for _, name := range files {
		if strings.HasSuffix(name, ".down.sql") {
			continue
		}
		version := migrationNumericPrefix(name)
		if version < 1 {
			continue
		}
		if _, ok := recentMigrationDownCompanionAllowlist[name]; ok {
			continue
		}
		downName := strings.TrimSuffix(name, ".sql") + ".down.sql"
		_, ok := fileSet[downName]
		require.Truef(t, ok, "missing down companion for %s", name)
	}
}

func TestRecentSchemaMigrationsPreferIdempotentDDL(t *testing.T) {
	files, err := fs.Glob(FS, "*.sql")
	require.NoError(t, err)

	for _, name := range files {
		if migrationNumericPrefix(name) < 74 {
			continue
		}

		raw, err := fs.ReadFile(FS, name)
		require.NoErrorf(t, err, "read %s", name)
		content := strings.ToUpper(string(raw))

		assertIdempotentPattern(t, name, content, "CREATE TABLE", "CREATE TABLE IF NOT EXISTS")
		assertOneOfPatterns(t, name, content, "CREATE INDEX", "CREATE INDEX IF NOT EXISTS", "CREATE INDEX CONCURRENTLY IF NOT EXISTS")
		assertIdempotentPattern(t, name, content, "DROP TABLE", "DROP TABLE IF EXISTS")
		assertOneOfPatterns(t, name, content, "DROP INDEX", "DROP INDEX IF EXISTS", "DROP INDEX CONCURRENTLY IF EXISTS")
		assertIdempotentPattern(t, name, content, "DROP COLUMN", "DROP COLUMN IF EXISTS")

		if strings.Contains(content, "ADD COLUMN") {
			require.Containsf(t, content, "ADD COLUMN IF NOT EXISTS", "%s should use IF NOT EXISTS for ADD COLUMN", name)
		}
	}
}

func assertIdempotentPattern(t *testing.T, name, content, keyword, expected string) {
	t.Helper()
	if strings.Contains(content, keyword) {
		require.Containsf(t, content, expected, "%s should prefer idempotent DDL: %s", name, expected)
	}
}

func assertOneOfPatterns(t *testing.T, name, content, keyword string, expected ...string) {
	t.Helper()
	if !strings.Contains(content, keyword) {
		return
	}
	for _, pattern := range expected {
		if strings.Contains(content, pattern) {
			return
		}
	}
	require.Failf(t, "missing idempotent DDL pattern", "%s should prefer one of: %s", name, strings.Join(expected, ", "))
}

func migrationNumericPrefix(name string) int {
	prefix, _, ok := strings.Cut(name, "_")
	if !ok {
		return 0
	}
	value, err := strconv.Atoi(prefix)
	if err != nil {
		return 0
	}
	return value
}
