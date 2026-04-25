package service

import (
	"context"
	"errors"
	"time"
)

type RuntimeAccountSlotOutcome string

const (
	RuntimeAccountSlotSucceeded        RuntimeAccountSlotOutcome = "succeeded"
	RuntimeAccountSlotNoAvailable      RuntimeAccountSlotOutcome = "no_available_account"
	RuntimeAccountSlotQueueFull        RuntimeAccountSlotOutcome = "queue_full"
	RuntimeAccountSlotAcquireError     RuntimeAccountSlotOutcome = "acquire_error"
	RuntimeAccountSlotWaitAcquireError RuntimeAccountSlotOutcome = "wait_acquire_error"
)

type RuntimeAccountSlotAcquireOrQueueFunc func(context.Context, int64, int, int) (*AcquireOrQueueResult, error)
type RuntimeAccountSlotWaitFunc func(context.Context, RuntimeAccountSlotWaitRequest) (func(), error)
type RuntimeAccountSlotDecrementWaitFunc func(context.Context, int64)
type RuntimeAccountSlotBindStickyFunc func(context.Context, *int64, string, int64) error

type RuntimeAccountSlotWaitRequest struct {
	AccountID      int64
	MaxConcurrency int
	Timeout        time.Duration
}

type RuntimeAccountSlotRequest struct {
	GroupID        *int64
	SessionHash    string
	Selection      *AccountSelectionResult
	AcquireOrQueue RuntimeAccountSlotAcquireOrQueueFunc
	WaitForSlot    RuntimeAccountSlotWaitFunc
	DecrementWait  RuntimeAccountSlotDecrementWaitFunc
	BindSticky     RuntimeAccountSlotBindStickyFunc
}

type RuntimeAccountSlotResult struct {
	Outcome     RuntimeAccountSlotOutcome
	Account     *Account
	ReleaseFunc func()
	WaitPlan    *AccountWaitPlan
	Err         error
	BindErr     error
}

func AcquireRuntimeAccountSlot(ctx context.Context, req RuntimeAccountSlotRequest) RuntimeAccountSlotResult {
	if ctx == nil {
		ctx = context.Background()
	}
	if req.Selection == nil || req.Selection.Account == nil {
		return RuntimeAccountSlotResult{Outcome: RuntimeAccountSlotNoAvailable}
	}

	account := req.Selection.Account
	result := RuntimeAccountSlotResult{
		Outcome:  RuntimeAccountSlotNoAvailable,
		Account:  account,
		WaitPlan: req.Selection.WaitPlan,
	}

	if req.Selection.Acquired {
		result.Outcome = RuntimeAccountSlotSucceeded
		result.ReleaseFunc = req.Selection.ReleaseFunc
		return result
	}
	if req.Selection.WaitPlan == nil {
		return result
	}
	if req.AcquireOrQueue == nil {
		result.Outcome = RuntimeAccountSlotAcquireError
		result.Err = errors.New("runtime account slot acquire hook is not configured")
		return result
	}

	queueResult, err := req.AcquireOrQueue(ctx, account.ID, req.Selection.WaitPlan.MaxConcurrency, req.Selection.WaitPlan.MaxWaiting)
	if err != nil {
		result.Outcome = RuntimeAccountSlotAcquireError
		result.Err = err
		return result
	}
	if queueResult == nil {
		result.Outcome = RuntimeAccountSlotAcquireError
		result.Err = errors.New("runtime account slot acquire returned nil result")
		return result
	}
	if queueResult.Acquired {
		result.Outcome = RuntimeAccountSlotSucceeded
		result.ReleaseFunc = queueResult.ReleaseFunc
		result.BindErr = bindRuntimeAccountSlotSticky(ctx, req, account.ID)
		return result
	}
	if !queueResult.QueueAllowed {
		result.Outcome = RuntimeAccountSlotQueueFull
		return result
	}

	waitCounted := queueResult.WaitCounted
	releaseWait := func() {
		if waitCounted {
			if req.DecrementWait != nil {
				req.DecrementWait(ctx, account.ID)
			}
			waitCounted = false
		}
	}

	if req.WaitForSlot == nil {
		releaseWait()
		result.Outcome = RuntimeAccountSlotWaitAcquireError
		result.Err = errors.New("runtime account slot wait hook is not configured")
		return result
	}

	releaseFunc, err := req.WaitForSlot(ctx, RuntimeAccountSlotWaitRequest{
		AccountID:      account.ID,
		MaxConcurrency: req.Selection.WaitPlan.MaxConcurrency,
		Timeout:        req.Selection.WaitPlan.Timeout,
	})
	if err != nil {
		releaseWait()
		result.Outcome = RuntimeAccountSlotWaitAcquireError
		result.Err = err
		return result
	}

	releaseWait()
	result.Outcome = RuntimeAccountSlotSucceeded
	result.ReleaseFunc = releaseFunc
	result.BindErr = bindRuntimeAccountSlotSticky(ctx, req, account.ID)
	return result
}

func bindRuntimeAccountSlotSticky(ctx context.Context, req RuntimeAccountSlotRequest, accountID int64) error {
	if req.BindSticky == nil {
		return nil
	}
	return req.BindSticky(ctx, req.GroupID, req.SessionHash, accountID)
}
