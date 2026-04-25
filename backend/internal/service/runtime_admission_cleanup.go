package service

type RuntimeAdmissionCleanupRequest struct {
	Account               *Account
	FailedAccountIDs      map[int64]struct{}
	AccountRelease        RuntimeForwardReleaseFunc
	QueueRelease          RuntimeForwardReleaseFunc
	ClearUpstreamAccepted RuntimeForwardReleaseFunc
}

type RuntimeAdmissionCleanupResult struct {
	AccountID    int64
	MarkedFailed bool
}

func CleanupRuntimeAdmissionDenied(req RuntimeAdmissionCleanupRequest) RuntimeAdmissionCleanupResult {
	if req.QueueRelease != nil {
		req.QueueRelease()
	}
	if req.ClearUpstreamAccepted != nil {
		req.ClearUpstreamAccepted()
	}
	if req.AccountRelease != nil {
		req.AccountRelease()
	}

	if req.Account == nil || req.FailedAccountIDs == nil {
		return RuntimeAdmissionCleanupResult{}
	}

	req.FailedAccountIDs[req.Account.ID] = struct{}{}
	return RuntimeAdmissionCleanupResult{
		AccountID:    req.Account.ID,
		MarkedFailed: true,
	}
}
