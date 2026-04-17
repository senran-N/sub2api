package service

import "context"

func executeIndexedRuntimeSelection[T any](
	ctx context.Context,
	pager *schedulerIndexedAccountPager,
	pageSize int,
	visit func(batch []*Account) (bool, *T, error),
	better func(candidate, current *T) bool,
) (bool, *T, error) {
	if pager == nil || visit == nil || better == nil {
		return false, nil, nil
	}
	if pageSize <= 0 {
		pageSize = 1
	}

	var best *T
	scopedFound := false
	for {
		batch, hasMore, err := pager.NextRefs(ctx, pageSize)
		if err != nil {
			return scopedFound, best, err
		}
		if len(batch) > 0 {
			scopedFound = true
			stop, candidate, err := visit(batch)
			if err != nil {
				return scopedFound, best, err
			}
			if better(candidate, best) {
				best = candidate
			}
			if stop {
				return scopedFound, best, nil
			}
		}
		if !hasMore {
			return scopedFound, best, nil
		}
	}
}
