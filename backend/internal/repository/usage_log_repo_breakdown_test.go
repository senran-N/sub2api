//go:build unit

package repository

import (
	"testing"

	"github.com/senran-N/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

func TestResolveEndpointColumn(t *testing.T) {
	tests := []struct {
		endpointType string
		want         string
	}{
		{"inbound", "ul.inbound_endpoint"},
		{"upstream", "ul.upstream_endpoint"},
		{"path", "ul.inbound_endpoint || ' -> ' || ul.upstream_endpoint"},
		{"", "ul.inbound_endpoint"},        // default
		{"unknown", "ul.inbound_endpoint"}, // fallback
	}

	for _, tc := range tests {
		t.Run(tc.endpointType, func(t *testing.T) {
			got := resolveEndpointColumn(tc.endpointType)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestResolveModelDimensionExpression(t *testing.T) {
	tests := []struct {
		modelType string
		want      string
	}{
		{usagestats.ModelSourceRequested, "COALESCE(NULLIF(TRIM(requested_model), ''), model)"},
		{usagestats.ModelSourceUpstream, "COALESCE(NULLIF(TRIM(upstream_model), ''), COALESCE(NULLIF(TRIM(requested_model), ''), model))"},
		{usagestats.ModelSourceMapping, "(COALESCE(NULLIF(TRIM(requested_model), ''), model) || ' -> ' || COALESCE(NULLIF(TRIM(upstream_model), ''), COALESCE(NULLIF(TRIM(requested_model), ''), model)))"},
		{"", "COALESCE(NULLIF(TRIM(requested_model), ''), model)"},
		{"invalid", "COALESCE(NULLIF(TRIM(requested_model), ''), model)"},
	}

	for _, tc := range tests {
		t.Run(tc.modelType, func(t *testing.T) {
			got := resolveModelDimensionExpression(tc.modelType)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestUsageLogActualCostAggregateExpr(t *testing.T) {
	tests := []struct {
		name       string
		tableAlias string
		userID     int64
		apiKeyID   int64
		accountID  int64
		want       string
	}{
		{
			name:       "account scoped uses account rate multiplier",
			tableAlias: "",
			accountID:  42,
			want:       "COALESCE(SUM(total_cost * COALESCE(account_rate_multiplier, 1)), 0) as actual_cost",
		},
		{
			name:       "account scoped with table alias",
			tableAlias: "ul",
			accountID:  42,
			want:       "COALESCE(SUM(ul.total_cost * COALESCE(ul.account_rate_multiplier, 1)), 0) as actual_cost",
		},
		{
			name:       "user scoped uses actual cost column",
			tableAlias: "ul",
			userID:     7,
			accountID:  42,
			want:       "COALESCE(SUM(ul.actual_cost), 0) as actual_cost",
		},
		{
			name:     "api key scoped uses actual cost column",
			apiKeyID: 5,
			want:     "COALESCE(SUM(actual_cost), 0) as actual_cost",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := usageLogActualCostAggregateExpr(tc.tableAlias, tc.userID, tc.apiKeyID, tc.accountID)
			require.Equal(t, tc.want, got)
		})
	}
}
