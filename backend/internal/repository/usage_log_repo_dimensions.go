package repository

import (
	"fmt"

	"github.com/senran-N/sub2api/internal/pkg/usagestats"
)

func usageLogActualCostAggregateExpr(tableAlias string, userID, apiKeyID, accountID int64) string {
	columnPrefix := ""
	if tableAlias != "" {
		columnPrefix = tableAlias + "."
	}
	if accountID > 0 && userID == 0 && apiKeyID == 0 {
		return fmt.Sprintf("COALESCE(SUM(%stotal_cost * COALESCE(%saccount_rate_multiplier, 1)), 0) as actual_cost", columnPrefix, columnPrefix)
	}
	return fmt.Sprintf("COALESCE(SUM(%sactual_cost), 0) as actual_cost", columnPrefix)
}

// resolveModelDimensionExpression maps model source type to a safe SQL expression.
func resolveModelDimensionExpression(modelType string) string {
	requestedExpr := "COALESCE(NULLIF(TRIM(requested_model), ''), model)"
	switch usagestats.NormalizeModelSource(modelType) {
	case usagestats.ModelSourceUpstream:
		return fmt.Sprintf("COALESCE(NULLIF(TRIM(upstream_model), ''), %s)", requestedExpr)
	case usagestats.ModelSourceMapping:
		return fmt.Sprintf("(%s || ' -> ' || COALESCE(NULLIF(TRIM(upstream_model), ''), %s))", requestedExpr, requestedExpr)
	default:
		return requestedExpr
	}
}

// resolveEndpointColumn maps endpoint type to the corresponding DB column name.
func resolveEndpointColumn(endpointType string) string {
	switch endpointType {
	case "upstream":
		return "ul.upstream_endpoint"
	case "path":
		return "ul.inbound_endpoint || ' -> ' || ul.upstream_endpoint"
	default:
		return "ul.inbound_endpoint"
	}
}
