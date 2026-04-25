package service

import (
	"context"
	"errors"
)

type RuntimeForwardFunc func(context.Context) (*ForwardResult, error)
type RuntimeForwardWriterSizeFunc func() int
type RuntimeForwardReleaseFunc func()
type RuntimeForwardReleaseWindowCostFunc func(context.Context, *WindowCostReservation) error

type RuntimeForwardAttemptRequest struct {
	Account               *Account
	Forward               RuntimeForwardFunc
	WriterSize            RuntimeForwardWriterSizeFunc
	AccountRelease        RuntimeForwardReleaseFunc
	QueueRelease          RuntimeForwardReleaseFunc
	ClearUpstreamAccepted RuntimeForwardReleaseFunc
	WindowCostReservation *WindowCostReservation
	ReleaseWindowCost     RuntimeForwardReleaseWindowCostFunc
}

type RuntimeForwardAttemptResult struct {
	Account              *Account
	Result               *ForwardResult
	Err                  error
	WriterSizeBefore     int
	WriterSizeAfter      int
	ResponseStarted      bool
	WindowCostReleaseErr error
}

func ExecuteRuntimeForwardAttempt(ctx context.Context, req RuntimeForwardAttemptRequest) RuntimeForwardAttemptResult {
	if ctx == nil {
		ctx = context.Background()
	}

	result := RuntimeForwardAttemptResult{
		Account:          req.Account,
		WriterSizeBefore: runtimeForwardWriterSize(req.WriterSize),
	}

	if req.Forward == nil {
		result.Err = errors.New("runtime forward hook is not configured")
	} else {
		result.Result, result.Err = req.Forward(ctx)
	}

	if req.QueueRelease != nil {
		req.QueueRelease()
	}
	if req.ClearUpstreamAccepted != nil {
		req.ClearUpstreamAccepted()
	}
	if req.AccountRelease != nil {
		req.AccountRelease()
	}

	result.WriterSizeAfter = runtimeForwardWriterSize(req.WriterSize)
	result.ResponseStarted = result.WriterSizeAfter != result.WriterSizeBefore
	if result.Err != nil && !result.ResponseStarted && req.WindowCostReservation != nil && req.ReleaseWindowCost != nil {
		result.WindowCostReleaseErr = req.ReleaseWindowCost(ctx, req.WindowCostReservation)
	}

	return result
}

func runtimeForwardWriterSize(writerSize RuntimeForwardWriterSizeFunc) int {
	if writerSize == nil {
		return 0
	}
	return writerSize()
}
