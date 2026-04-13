package service

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewScheduledTestRunnerService_DefaultTunables(t *testing.T) {
	svc := NewScheduledTestRunnerService(nil, nil, nil, nil, nil)

	if svc.startDelay != scheduledTestDefaultStartDelay {
		t.Fatalf("startDelay = %v, want %v", svc.startDelay, scheduledTestDefaultStartDelay)
	}
	if svc.runTimeout != scheduledTestDefaultRunTimeout {
		t.Fatalf("runTimeout = %v, want %v", svc.runTimeout, scheduledTestDefaultRunTimeout)
	}
	if svc.stopWait != scheduledTestDefaultStopWait {
		t.Fatalf("stopWait = %v, want %v", svc.stopWait, scheduledTestDefaultStopWait)
	}
	if svc.maxWorkers != scheduledTestDefaultMaxWorkers {
		t.Fatalf("maxWorkers = %d, want %d", svc.maxWorkers, scheduledTestDefaultMaxWorkers)
	}
}

func TestScheduledTestRunnerService_RunScheduledUsesTunables(t *testing.T) {
	svc := NewScheduledTestRunnerService(nil, nil, nil, nil, nil)
	svc.rateLimitSvc = &RateLimitService{}
	svc.startDelay = 17 * time.Millisecond
	svc.runTimeout = 200 * time.Millisecond
	svc.maxWorkers = 1

	fixedNow := time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return fixedNow }

	var slept time.Duration
	svc.sleep = func(d time.Duration) {
		slept = d
	}

	plans := []*ScheduledTestPlan{
		{ID: 1, AccountID: 101, ModelID: "model-a", CronExpression: "* * * * *", MaxResults: 3},
		{ID: 2, AccountID: 102, ModelID: "model-b", CronExpression: "* * * * *", MaxResults: 3},
	}

	svc.listDue = func(ctx context.Context, now time.Time) ([]*ScheduledTestPlan, error) {
		if now != fixedNow {
			t.Fatalf("listDue now = %v, want %v", now, fixedNow)
		}
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatalf("listDue context missing deadline")
		}
		remaining := time.Until(deadline)
		if remaining <= 0 || remaining > svc.runTimeout {
			t.Fatalf("deadline remaining = %v, want within (0,%v]", remaining, svc.runTimeout)
		}
		return plans, nil
	}

	var inFlight int32
	var maxInFlight int32
	svc.runTest = func(ctx context.Context, accountID int64, modelID string) (*ScheduledTestResult, error) {
		current := atomic.AddInt32(&inFlight, 1)
		for {
			seen := atomic.LoadInt32(&maxInFlight)
			if current <= seen || atomic.CompareAndSwapInt32(&maxInFlight, seen, current) {
				break
			}
		}
		time.Sleep(15 * time.Millisecond)
		atomic.AddInt32(&inFlight, -1)

		return &ScheduledTestResult{
			Status:     "success",
			StartedAt:  fixedNow,
			FinishedAt: fixedNow.Add(time.Second),
		}, nil
	}

	var (
		mu          sync.Mutex
		savedIDs    []int64
		updatedIDs  []int64
		recoveredID []int64
	)
	svc.saveResult = func(ctx context.Context, planID int64, maxResults int, result *ScheduledTestResult) error {
		mu.Lock()
		defer mu.Unlock()
		savedIDs = append(savedIDs, planID)
		return nil
	}
	svc.updateAfterRun = func(ctx context.Context, id int64, lastRunAt time.Time, nextRunAt time.Time) error {
		if lastRunAt != fixedNow {
			t.Fatalf("lastRunAt = %v, want %v", lastRunAt, fixedNow)
		}
		if nextRunAt.IsZero() {
			t.Fatalf("nextRunAt should be populated")
		}
		mu.Lock()
		defer mu.Unlock()
		updatedIDs = append(updatedIDs, id)
		return nil
	}
	svc.recoverAccount = func(ctx context.Context, accountID int64) (*SuccessfulTestRecoveryResult, error) {
		mu.Lock()
		defer mu.Unlock()
		recoveredID = append(recoveredID, accountID)
		return &SuccessfulTestRecoveryResult{ClearedError: true}, nil
	}

	plans[0].AutoRecover = true
	plans[1].AutoRecover = true

	svc.runScheduled()

	if slept != svc.startDelay {
		t.Fatalf("sleep = %v, want %v", slept, svc.startDelay)
	}
	if got := atomic.LoadInt32(&maxInFlight); got != 1 {
		t.Fatalf("maxInFlight = %d, want 1", got)
	}
	if len(savedIDs) != 2 {
		t.Fatalf("savedIDs len = %d, want 2", len(savedIDs))
	}
	if len(updatedIDs) != 2 {
		t.Fatalf("updatedIDs len = %d, want 2", len(updatedIDs))
	}
	if len(recoveredID) != 2 {
		t.Fatalf("recoveredID len = %d, want 2", len(recoveredID))
	}
}
