package service

import "context"

func forEachIndexedSnapshotPage(
	ctx context.Context,
	pager *schedulerIndexedAccountPager,
	pageSize int,
	visit func(batch []Account) (bool, error),
) error {
	if pager == nil || visit == nil {
		return nil
	}
	if pageSize <= 0 {
		pageSize = 1
	}

	for {
		batch, hasMore, err := pager.Next(ctx, pageSize)
		if err != nil {
			return err
		}
		if len(batch) > 0 {
			stop, err := visit(batch)
			if err != nil {
				return err
			}
			if stop {
				return nil
			}
		}
		if !hasMore {
			return nil
		}
	}
}
