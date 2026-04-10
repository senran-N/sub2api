package service

func selectFirstOrderedRuntimeSelection[T any](
	ordered []*Account,
	trySelect func(account *Account) (*T, error, bool),
) (*T, error, bool) {
	if len(ordered) == 0 || trySelect == nil {
		return nil, nil, false
	}
	for _, account := range ordered {
		defaultSchedulingRuntimeKernelStats.orderedRuntimeProbes.Add(1)
		result, err, ok := trySelect(account)
		if err != nil {
			return nil, err, false
		}
		if ok {
			return result, nil, true
		}
	}
	return nil, nil, false
}

func selectFirstOrderedWaitPlan(
	ordered []*Account,
	trySelect func(account *Account) (*AccountSelectionResult, string, bool),
) (*AccountSelectionResult, bool) {
	if len(ordered) == 0 || trySelect == nil {
		return nil, false
	}
	for _, account := range ordered {
		defaultSchedulingRuntimeKernelStats.orderedWaitPlanProbes.Add(1)
		result, _, ok := trySelect(account)
		if ok {
			return result, true
		}
	}
	return nil, false
}
