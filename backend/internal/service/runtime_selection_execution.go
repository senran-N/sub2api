package service

import (
	"context"
	"time"
)

type runtimeAcquireSelectionSpec struct {
	account      *Account
	prepare      func(account *Account) *Account
	acquire      func(account *Account) (*AcquireResult, error)
	finalize     func(account *Account) *Account
	allowSession func(account *Account) bool
	bind         func(account *Account)
	onAcquired   func(account *Account, acquired *AcquireResult) *AccountSelectionResult
}

const (
	runtimeAcquireMissUnavailable = "acquire_unavailable"
	runtimeAcquireMissFinalize    = "finalize_miss"
	runtimeAcquireMissSession     = "session_limit"
)

func tryAcquireRuntimeSelectionDetailed(
	ctx context.Context,
	spec runtimeAcquireSelectionSpec,
) (*AccountSelectionResult, string, error, bool) {
	account := spec.account
	if account == nil {
		return nil, "", nil, false
	}
	if spec.prepare != nil {
		account = spec.prepare(account)
	}
	if account == nil || spec.acquire == nil {
		return nil, "", nil, false
	}
	defaultSchedulingRuntimeKernelStats.runtimeAcquireAttempts.Add(1)

	result, err := spec.acquire(account)
	if err != nil {
		defaultSchedulingRuntimeKernelStats.runtimeAcquireErrors.Add(1)
		return nil, runtimeAcquireMissUnavailable, err, false
	}
	if result == nil || !result.Acquired {
		defaultSchedulingRuntimeKernelStats.runtimeAcquireMisses.Add(1)
		return nil, runtimeAcquireMissUnavailable, nil, false
	}

	selected := account
	if spec.finalize != nil {
		selected = spec.finalize(account)
	}
	if selected == nil {
		if result.ReleaseFunc != nil {
			result.ReleaseFunc()
		}
		defaultSchedulingRuntimeKernelStats.runtimeFinalizeMisses.Add(1)
		return nil, runtimeAcquireMissFinalize, nil, false
	}
	if spec.allowSession != nil && !spec.allowSession(selected) {
		if result.ReleaseFunc != nil {
			result.ReleaseFunc()
		}
		defaultSchedulingRuntimeKernelStats.runtimeSessionMisses.Add(1)
		return nil, runtimeAcquireMissSession, nil, false
	}
	if spec.bind != nil {
		spec.bind(selected)
	}
	defaultSchedulingRuntimeKernelStats.runtimeAcquireSuccess.Add(1)
	if spec.onAcquired != nil {
		return spec.onAcquired(selected, result), "", nil, true
	}
	return newAcquiredAccountSelection(selected, result.ReleaseFunc), "", nil, true
}

func tryAcquireRuntimeSelection(
	ctx context.Context,
	spec runtimeAcquireSelectionSpec,
) (*AccountSelectionResult, error, bool) {
	result, _, err, ok := tryAcquireRuntimeSelectionDetailed(ctx, spec)
	return result, err, ok
}

type runtimeWaitPlanSpec struct {
	account      *Account
	prepare      func(account *Account) *Account
	allowWait    func(account *Account) (string, bool)
	allowSession func(account *Account) bool
	finalize     func(account *Account) *Account
	timeout      time.Duration
	maxWaiting   int
}

func tryBuildRuntimeWaitPlan(
	ctx context.Context,
	spec runtimeWaitPlanSpec,
) (*AccountSelectionResult, string, bool) {
	_ = ctx

	account := spec.account
	if account == nil {
		return nil, "", false
	}
	if spec.prepare != nil {
		account = spec.prepare(account)
	}
	defaultSchedulingRuntimeKernelStats.runtimeWaitPlanAttempts.Add(1)
	if account == nil {
		defaultSchedulingRuntimeKernelStats.runtimeWaitPlanRejected.Add(1)
		return nil, "", false
	}
	if spec.allowWait != nil {
		reason, ok := spec.allowWait(account)
		if !ok {
			defaultSchedulingRuntimeKernelStats.runtimeWaitPlanRejected.Add(1)
			return nil, reason, false
		}
	}
	if spec.allowSession != nil && !spec.allowSession(account) {
		defaultSchedulingRuntimeKernelStats.runtimeWaitPlanRejected.Add(1)
		return nil, "session_limit", false
	}

	selected := account
	if spec.finalize != nil {
		selected = spec.finalize(account)
	}
	if selected == nil {
		defaultSchedulingRuntimeKernelStats.runtimeWaitPlanFinalMiss.Add(1)
		return nil, "hydrate_miss", false
	}
	defaultSchedulingRuntimeKernelStats.runtimeWaitPlanSuccess.Add(1)
	return newWaitPlanAccountSelection(selected, spec.timeout, spec.maxWaiting), "", true
}
