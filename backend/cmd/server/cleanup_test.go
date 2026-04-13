package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunCleanup_TimeoutDoesNotPanicWhenParallelStepReturnsLateError(t *testing.T) {
	parallelStepDone := make(chan struct{})
	var infraCalls atomic.Int32

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	runCleanup(
		ctx,
		[]cleanupStep{
			{
				name: "slow-error",
				fn: func() error {
					time.Sleep(30 * time.Millisecond)
					close(parallelStepDone)
					return errors.New("late cleanup error")
				},
			},
		},
		[]cleanupStep{
			{
				name: "infra",
				fn: func() error {
					infraCalls.Add(1)
					return nil
				},
			},
		},
	)

	select {
	case <-parallelStepDone:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("parallel cleanup step did not finish")
	}

	require.EqualValues(t, 1, infraCalls.Load())
}

func TestRunCleanup_ParallelCancellationRunsBeforeInfraCleanup(t *testing.T) {
	subscriberCtx, cancelSubscriber := context.WithCancel(context.Background())
	defer cancelSubscriber()

	var infraSawCanceled atomic.Bool

	runCleanup(
		context.Background(),
		[]cleanupStep{
			callbackStep("subscriber", cancelSubscriber),
		},
		[]cleanupStep{
			{
				name: "infra",
				fn: func() error {
					infraSawCanceled.Store(subscriberCtx.Err() == context.Canceled)
					return nil
				},
			},
		},
	)

	require.True(t, infraSawCanceled.Load())
}
