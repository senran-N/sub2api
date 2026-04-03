package service

import (
	"errors"

	"github.com/senran-N/sub2api/internal/domain"
)

type OpsQueryMode = domain.OpsQueryMode

const (
	OpsQueryModeAuto   = domain.OpsQueryModeAuto
	OpsQueryModeRaw    = domain.OpsQueryModeRaw
	OpsQueryModePreagg = domain.OpsQueryModePreagg
)

// ErrOpsPreaggregatedNotPopulated indicates that raw logs exist for a window, but the
// pre-aggregation tables are not populated yet. This is primarily used to implement
// the forced `preagg` mode UX.
var ErrOpsPreaggregatedNotPopulated = errors.New("ops pre-aggregated tables not populated")

var ParseOpsQueryMode = domain.ParseOpsQueryMode

func shouldFallbackOpsPreagg(filter *OpsDashboardFilter, err error) bool {
	return filter != nil &&
		filter.QueryMode == OpsQueryModeAuto &&
		errors.Is(err, ErrOpsPreaggregatedNotPopulated)
}

func cloneOpsFilterWithMode(filter *OpsDashboardFilter, mode OpsQueryMode) *OpsDashboardFilter {
	if filter == nil {
		return nil
	}
	cloned := *filter
	cloned.QueryMode = mode
	return &cloned
}
