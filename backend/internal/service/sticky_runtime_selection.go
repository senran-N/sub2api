package service

type stickyRuntimeSelectionSpec struct {
	tryAcquire    func() (*AccountSelectionResult, string, bool)
	onAcquireMiss func(reason string)
	buildWaitPlan func() (*AccountSelectionResult, string, bool)
}

func trySelectStickyRuntimeSelection(
	spec stickyRuntimeSelectionSpec,
) (*AccountSelectionResult, string, bool) {
	if spec.tryAcquire == nil {
		return nil, "", false
	}

	result, missReason, ok := spec.tryAcquire()
	if ok {
		return result, "", true
	}
	if missReason != "" && spec.onAcquireMiss != nil {
		spec.onAcquireMiss(missReason)
	}
	if spec.buildWaitPlan == nil {
		return nil, missReason, false
	}
	return spec.buildWaitPlan()
}
