package service

import "context"

func selectBestAccountFromIndexedSnapshotPager(
	ctx context.Context,
	pager *schedulerIndexedAccountPager,
	pageSize int,
	initialSupported bool,
	selectBatch func(batch []*Account) (*Account, error),
	better func(candidate, current *Account) bool,
) (*Account, bool, error) {
	if pager == nil || selectBatch == nil || better == nil {
		return nil, initialSupported, nil
	}

	supported := initialSupported
	var selected *Account
	err := forEachIndexedSnapshotPage(ctx, pager, pageSize, func(batch []*Account) (bool, error) {
		supported = true
		batchBest, err := selectBatch(batch)
		if err != nil {
			return false, err
		}
		if better(batchBest, selected) {
			selected = batchBest
		}
		return false, nil
	})
	if err != nil {
		return nil, supported, err
	}
	return selected, supported, nil
}
